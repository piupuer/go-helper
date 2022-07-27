package v1

import (
	"bufio"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/piupuer/go-helper/pkg/log"
	"github.com/piupuer/go-helper/pkg/middleware"
	"github.com/piupuer/go-helper/pkg/req"
	"github.com/piupuer/go-helper/pkg/utils"
	"golang.org/x/crypto/ssh"
	"time"
	"unicode/utf8"
)

type ptyRequestMsg struct {
	Term     string
	Columns  uint32
	Rows     uint32
	Width    uint32
	Height   uint32
	Modelist string
}

// MachineShellWs start shell websocket
func MachineShellWs(options ...func(*Options)) gin.HandlerFunc {
	return func(c *gin.Context) {
		var r req.MachineShellWs
		req.ShouldBind(c, &r)

		conn, _ := middleware.WsUpgrader.Upgrade(c.Writer, c.Request, nil)
		defer conn.Close()

		active := time.Now()

		// get ssh client
		client, err := utils.GetSshClient(utils.SshConfig{
			Host:      r.Host,
			Port:      int(r.SshPort),
			LoginName: r.LoginName,
			LoginPwd:  r.LoginPwd,
		})
		if err != nil {
			log.WithContext(c).WithError(err).Error("connect ssh client failed")
			conn.WriteMessage(websocket.TextMessage, []byte("\n"+err.Error()))
			return
		}

		// open ssh channel
		channel, incomingRequests, err := client.Conn.OpenChannel("session", nil)
		if err != nil {
			log.WithContext(c).WithError(err).Error("connect ssh channel failed")
			conn.WriteMessage(websocket.TextMessage, []byte("\n"+err.Error()))
			return
		}
		defer channel.Close()
		defer client.Close()

		go func() {
			for item := range incomingRequests {
				if item.WantReply {
					// reply
					item.Reply(false, nil)
				}
			}
		}()

		modes := ssh.TerminalModes{
			ssh.ECHO:          1,
			ssh.TTY_OP_ISPEED: 14400,
			ssh.TTY_OP_OSPEED: 14400,
		}
		var modeList []byte
		for k, v := range modes {
			kv := struct {
				Key byte
				Val uint32
			}{k, v}
			modeList = append(modeList, ssh.Marshal(&kv)...)
		}

		modeList = append(modeList, 0)

		rows := uint32(r.Rows)
		cols := uint32(r.Cols)
		ptyReq := ptyRequestMsg{
			Term:     "xterm",
			Columns:  rows,
			Rows:     cols,
			Width:    rows,
			Height:   cols,
			Modelist: string(modeList),
		}
		ok, err := channel.SendRequest("pty-req", true, ssh.Marshal(&ptyReq))
		if !ok || err != nil {
			log.WithContext(c).WithError(err).Error("send pseudo terminal request failed")
			conn.WriteMessage(websocket.TextMessage, []byte("\n"+err.Error()))
			return
		}

		ok, err = channel.SendRequest("shell", true, nil)
		if !ok || err != nil {
			log.WithContext(c).WithError(err).Error("send shell failed")
			conn.WriteMessage(websocket.TextMessage, []byte("\n"+err.Error()))
			return
		}

		go func() {
			br := bufio.NewReader(channel)
			var buf []byte

			t := time.NewTimer(time.Millisecond * 100)
			defer t.Stop()
			cr := make(chan rune)

			go func() {
				for {
					x, size, e := br.ReadRune()
					if e != nil {
						log.WithContext(c).WithError(e).Warn("read shell failed")
						break
					}
					if size > 0 {
						cr <- x
					}
				}
			}()

			for {
				select {
				case <-t.C:
					if len(buf) != 0 {
						err = conn.WriteMessage(websocket.TextMessage, buf)
						buf = []byte{}
						if err != nil {
							log.WithContext(c).WithError(err).Error("write msg to %s failed", conn.RemoteAddr())
							return
						}
					}
					t.Reset(time.Millisecond * 100)
				case d := <-cr:
					if d != utf8.RuneError {
						p := make([]byte, utf8.RuneLen(d))
						utf8.EncodeRune(p, d)
						buf = append(buf, p...)
					} else {
						buf = append(buf, []byte("@")...)
					}
				}
			}

		}()

		// timeout handler
		go func() {
			for {
				time.Sleep(time.Minute * 5)
				cost := time.Since(active)
				if cost.Minutes() >= 30 {
					// if it is inactive for more than 30 minutes, the connection will be closed automatically
					conn.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf("\r\nif it is inactive for more than [ %s ] minutes, the connection will be closed automatically", cost.String())))
					conn.Close()
					break
				}
			}
		}()
		conn.WriteMessage(websocket.TextMessage, []byte("\r\n["+r.Host+"]connect success"))
		conn.WriteMessage(websocket.TextMessage, []byte("\r\nif it is inactive for more than 30 minutes, the connection will be closed automatically\r\n\r\n"))

		if r.InitCmd != "" {
			go func() {
				time.Sleep(time.Second * 1)
				channel.Write([]byte(r.InitCmd + "\r\n"))
			}()
		}

		oldCmd := ""
		// read user input
		for {
			m, p, e := conn.ReadMessage()
			active = time.Now()
			if e != nil {
				log.WithContext(c).WithError(e).Warn("connection %s lost", conn.RemoteAddr())
				break
			}

			if m == websocket.TextMessage {
				s := string(p)
				if s == "\r" {
					cmd := oldCmd
					oldCmd = ""
					if e1 := utils.IsSafetyCmd(cmd); e1 != nil {
						e1 = conn.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf("\r\n\r\n%v\r\n\r\n", e1)))
						if e1 != nil {
							log.WithContext(c).WithError(e1).Warn("write msg to %s failed", conn.RemoteAddr())
							break
						}
						// write Ctrl C
						if _, err := channel.Write([]byte{3}); nil != err {
							break
						}
						continue
					}
				} else {
					oldCmd += s
				}
				if _, e2 := channel.Write(p); nil != e2 {
					break
				}
			}
		}
	}
}
