package ms

import (
	"embed"
	"fmt"
	"github.com/piupuer/go-helper/pkg/log"
	"io/ioutil"
	"os"
)

type ConfBox struct {
	Fs  embed.FS
	Dir string
}

func (c ConfBox) Get(filename string) (bs []byte) {
	if filename == "" {
		return
	}
	f := fmt.Sprintf("%s%s%s", c.Dir, string(os.PathSeparator), filename)
	var err error
	// read from system
	bs, err = ioutil.ReadFile(f)
	if err != nil {
		log.Warn("[conf box]read file %s from system failed: %v", f, err)
		err = nil
	}
	if len(bs) == 0 {
		// read from embed
		bs, err = c.Fs.ReadFile(f)
		if err != nil {
			log.Warn("[conf box]read file %s from embed failed: %v", f, err)
		}
	}
	return bs
}
