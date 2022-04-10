package video

import (
	"github.com/piupuer/go-helper/pkg/log"
	"github.com/shopspring/decimal"
	"os/exec"
	"regexp"
	"strings"
)

var (
	durationReg = regexp.MustCompile(`.*Duration:\s(.*?),.*bitrate:\s(\S+)`)
	videoReg    = regexp.MustCompile(`Stream.*Video.*\s(\d+)x(\d+)(?:.*?(\S+)\skb/s)?.*?(\S+)\sfps`)
	audioReg    = regexp.MustCompile(`Stream.*Audio.*?(\d+)\sHz.*\s(\S+)\skb/s`)
)

func parseDuration(str string) (duration, bitrate int64) {
	res := durationReg.FindStringSubmatch(str)
	if len(res) == 3 {
		arr1 := strings.Split(res[1], ":")
		b, _ := decimal.NewFromString(res[2])
		bitrate = b.IntPart()
		arr2 := strings.Split(arr1[2], ".")
		arr1[2] = arr2[0]
		arr1 = append(arr1, arr2[1])
		hour, _ := decimal.NewFromString(arr1[0])
		min, _ := decimal.NewFromString(arr1[1])
		sec, _ := decimal.NewFromString(arr1[2])
		milli, _ := decimal.NewFromString(arr1[3])
		duration = hour.Mul(decimal.NewFromInt(3600000)).
			Add(min.Mul(decimal.NewFromInt(60000))).
			Add(sec.Mul(decimal.NewFromInt(1000))).
			Add(milli.Mul(decimal.NewFromInt(10))).IntPart()
	}
	return
}

func parseVideo(str string) (width, height, bitrate , fps int64) {
	res := videoReg.FindStringSubmatch(str)
	if len(res) == 5 {
		w, _ := decimal.NewFromString(res[1])
		h, _ := decimal.NewFromString(res[2])
		b, _ := decimal.NewFromString(res[3])
		f, _ := decimal.NewFromString(res[4])
		width = w.IntPart()
		height = h.IntPart()
		bitrate = b.IntPart()
		fps = f.IntPart()
	}
	return
}

func parseAudio(str string) (hz, bitrate int64) {
	res := audioReg.FindStringSubmatch(str)
	if len(res) == 3 {
		h, _ := decimal.NewFromString(res[1])
		b, _ := decimal.NewFromString(res[2])
		hz = h.IntPart()
		bitrate = b.IntPart()
	}
	return
}

func Stat(url string) (info Info) {
	if !hasFfmpeg() {
		log.Warn("ffmpeg not install")
		return
	}
	cmd := exec.Command("ffmpeg", "-i", url)
	res, _ := cmd.CombinedOutput()
	s := string(res)
	info.Duration, info.Bitrate = parseDuration(s)
	info.Width, info.Height, info.VideoBitrate, info.Fps = parseVideo(s)
	info.AudioHz, info.AudioBitrate = parseAudio(s)
	return
}
