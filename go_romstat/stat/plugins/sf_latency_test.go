package plugins

import (
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"romstat/stat/utils"
)

func readFrameTimes() []float64 {
	fp, _ := os.Open("frametime_data.txt")
	txtLines, _ := io.ReadAll(fp)
	lines := strings.Split(string(txtLines), "\n")
	retFloats := make([]float64, 0)
	for _, line := range lines {
		f, _ := strconv.ParseFloat(line, 64)
		retFloats = append(retFloats, f)
	}
	return retFloats
}
func TestJankCount(t *testing.T) {
	plugin := &SfLatencyStatPlugin{
		secOuputFrameData: &OutputFrameData{},
		debugLog:          utils.NewDebugLogger(),
	}
	frameTimes := readFrameTimes()
	//fmt.Println(frameTimes)
	for _, ft := range frameTimes {
		frameData := &SfFrameData{
			FrameTime: int64(ft * float64(time.Millisecond)),
		}
		plugin.calcFrameTime(frameData)
	}
	fmt.Println("smallJank", plugin.secOuputFrameData.SmallJank)
	fmt.Println("jank", plugin.secOuputFrameData.Jank)
	fmt.Println("bigJank", plugin.secOuputFrameData.BigJank)
	if plugin.secOuputFrameData.SmallJank != 200 {
		t.Errorf("ERROR: plugin.secOuputFrameData.SmallJank=%d, expect 200", plugin.secOuputFrameData.SmallJank)
	}
	if plugin.secOuputFrameData.Jank != 102 {
		t.Errorf("ERROR: plugin.secOuputFrameData.Jank=%d, expect 102", plugin.secOuputFrameData.Jank)
	}
	if plugin.secOuputFrameData.BigJank != 45 {
		t.Errorf("ERROR: plugin.secOuputFrameData.BigJank=%d, expect 45", plugin.secOuputFrameData.BigJank)
	}
}
