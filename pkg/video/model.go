package video

import (
	"fmt"
	"github.com/shopspring/decimal"
)

type Info struct {
	Duration     int64 // millisecond
	Bitrate      int64
	Width        int64
	Height       int64
	Fps          int64
	VideoBitrate int64
	AudioBitrate int64
	AudioHz      int64
}

func (i Info) SecondDuration() int64 {
	return decimal.NewFromInt(i.Duration / 1000).IntPart()
}

func (i Info) Duration2Str() (rp string) {
	milli := i.Duration % 1000
	s := i.Duration / 1000
	second := s % 60
	m := s / 60
	min := m % 60
	h := min / 60
	hour := h % 60
	if hour == 0 {
		rp = fmt.Sprintf("%02d:%02d.%02d", min, second, milli/10)
	} else {
		rp = fmt.Sprintf("%02d:%02d:%02d.%02d", hour, min, second, milli/10)
	}
	return
}

func (i Info) SecondDuration2Str() (rp string) {
	s := i.Duration / 1000
	second := s % 60
	m := s / 60
	min := m % 60
	h := min / 60
	hour := h % 60
	if hour == 0 {
		rp = fmt.Sprintf("%02d:%02d", min, second)
	} else {
		rp = fmt.Sprintf("%02d:%02d:%02d", hour, min, second)
	}
	return
}
