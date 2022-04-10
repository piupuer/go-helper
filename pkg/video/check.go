package video

import (
	"os/exec"
)

func hasFfmpeg() (ok bool) {
	cmd := exec.Command("command", "-v", "ffmpeg")
	_, err := cmd.Output()
	if err != nil {
		return
	}
	ok = true
	return
}
