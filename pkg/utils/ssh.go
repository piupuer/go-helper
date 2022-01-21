package utils

import (
	"bytes"
	"fmt"
	"github.com/piupuer/go-helper/pkg/log"
	"github.com/pkg/errors"
	"golang.org/x/crypto/ssh"
	"net"
	"path"
	"strings"
	"time"
)

const DefaultSshTimeout = 5

type SshConfig struct {
	LoginName string
	LoginPwd  string
	Host      string
	Port      int
	Timeout   int
}

type SshResult struct {
	Connect bool   `json:"connect"`
	Result  string `json:"result"`
	Err     error  `json:"err"`
}

// check cmd is safe(rm *, rm /*)
func IsSafetyCmd(cmd string) error {
	c := path.Clean(strings.ToLower(cmd))
	if strings.Contains(c, "rm") {
		if len(strings.Split(c, "/")) <= 1 {
			return errors.Errorf("rm command %s cannot delete files smaller than level 2 dir", cmd)
		}
	}
	return nil
}

func ExecRemoteShell(config SshConfig, cmds []string) SshResult {
	return ExecRemoteShellWithTimeout(config, cmds, 0)
}

func ExecRemoteShellWithTimeout(config SshConfig, cmds []string, timeout int64) SshResult {
	var session *ssh.Session
	client, err := GetSshClient(config)
	if err != nil {
		return SshResult{
			Connect: false,
			Err:     err,
		}
	}
	// create session
	if session, err = client.NewSession(); err != nil {
		return SshResult{
			Connect: false,
			Err:     errors.Wrapf(err, "create ssh %s session failed", config.Host),
		}
	}
	defer closeClient(session, client)

	go func() {
		if timeout > 0 {
			sleep, err := time.ParseDuration(fmt.Sprintf("%ds", timeout))
			if err != nil {
				log.Error("[exec remote shell]close ssh session failed: %v", err)
				return
			}
			time.Sleep(sleep)
			closeClient(session, client)
		}
	}()

	command := ""

	for i, cmd := range cmds {
		if err := IsSafetyCmd(cmd); err != nil {
			return SshResult{
				Connect: true,
				Err:     err,
			}
		}
		if i == 0 {
			command = cmd
		} else {
			command = command + " && " + cmd
		}
	}

	var e bytes.Buffer
	var b bytes.Buffer
	session.Stdout = &b
	session.Stderr = &e
	if command != "" {
		if err := session.Run(command); err != nil {
			return SshResult{
				Connect: true,
				Err:     errors.Wrapf(err, "exec cmd: %s failed", command),
				Result:  e.String(),
			}
		}
		log.Error("[exec remote shell]exec cmd: %s", command)
	}
	return SshResult{
		Result:  b.String(),
		Err:     nil,
		Connect: true,
	}
}

func closeClient(session *ssh.Session, client *ssh.Client) {
	err := client.Close()
	if err != nil {
		log.Error("[exec remote shell]close ssh client failed: %v", err)
	}
	session.Close()
}

func GetSshClient(config SshConfig) (*ssh.Client, error) {
	var (
		auth         []ssh.AuthMethod
		addr         string
		clientConfig *ssh.ClientConfig
		client       *ssh.Client
		err          error
	)
	// get auth method
	auth = make([]ssh.AuthMethod, 0)
	auth = append(auth, ssh.Password(config.LoginPwd))

	if config.Timeout == 0 {
		config.Timeout = DefaultSshTimeout
	}
	clientConfig = &ssh.ClientConfig{
		User:    config.LoginName,
		Auth:    auth,
		Timeout: time.Second * time.Duration(config.Timeout),
		HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error {
			return nil
		},
	}

	// connect to ssh
	addr = fmt.Sprintf("%s:%d", config.Host, config.Port)
	if client, err = ssh.Dial("tcp", addr, clientConfig); err != nil {
		return nil, errors.Wrapf(err, "connect ssh %s failed", addr)
	}
	return client, nil
}
