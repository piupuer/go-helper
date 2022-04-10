package video

import (
	"fmt"
	"testing"
)

func TestStat(t *testing.T) {
	info := Stat("1.mp4")
	fmt.Println(info, info.Duration2Str(), info.SecondDuration2Str())
}
