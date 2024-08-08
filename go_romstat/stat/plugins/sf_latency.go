// Copyright (c) 2021-2023 https://www.haimacloud.com/
// SPDX-License-Identifier: MIT
package plugins

import (
	"context"
	"fmt"
	"math"
	"romstat/stat/data"
	"romstat/stat/utils"
	"sync"
	"time"
)

type SfFrameData struct {
	RefTimestamp []int64 //timestamp list
	DisplayTs    int64   //vsync time interval
	FrameTime    int64   //display time interval
	Jank         bool    //is normal jank frame
	BigJank      bool    //is big jank frame
	SmallJank    bool    //is small jank frame
}

type OutputFrameData struct {
	StartTs     int   //start timestamp per second
	Fps         int   //frame count per second
	Jank        int   //count of normal jank
	BigJank     int   //count of big jank
	SmallJank   int   //count of small jank
	JankTotalTs int64 //duration of total jank times per second
}

// For Android Only
type SfPkgSurfaceData struct {
	PkgName     string
	SurfaceView string
}

// For Windows Only
type DesktopFramerateCoutner struct {
	FramesTimestamp    []int64
	maxFrames          int
	debugLog           utils.Logger
	ctx                context.Context
	latestPresentTs    int64
	frameTimestampLock sync.Mutex
}

type SfLatencyStatPlugin struct {
	lastSmallJank3Frames []*SfFrameData //Data of the last small jank frames
	lastJank3Frames      []*SfFrameData //Data of the last 3 frames
	secOuputFrameData    *OutputFrameData

	prevPresentTs         int64  //Last display time
	prevMaxVsyncTimestamp int64  //Last vsync time
	currentSurfaceView    string //Current surface view name
	currentPkgName        string //Current application package name
	vSyncPeriod           int64  //Frame interval

	lastFpsTimestamp int64

	monitorProcessName string

	debugLog         utils.Logger
	shell            *utils.AndroidShell
	sdkVersion       int64
	lockedPkgSurface *SfPkgSurfaceData

	d3dxLoopCounter *DesktopFramerateCoutner
}

func (t *SfLatencyStatPlugin) Run() {
	go utils.SetTimerMilliSecond(200, t.runCollectThread)
}

func (t *SfLatencyStatPlugin) GetTypes() []*data.PluginType {
	return []*data.PluginType{
		{Name: "fps", DisplayName: "fps", IsCmdShow: true},
		{Name: "jank", DisplayName: "jank", IsCmdShow: true},
		{Name: "Bjank", DisplayName: "Bjank", IsCmdShow: true},
		{Name: "jankTime", DisplayName: "jT(ms)", IsCmdShow: true},
		{Name: "Sjank", DisplayName: "Sjank", IsCmdShow: true},
		{Name: "jankPercent", DisplayName: "jT(%)", IsCmdShow: false},
	}
}

func (t *SfLatencyStatPlugin) GetData() map[string]string {
	secData := t.secOuputFrameData
	fps := secData.Fps
	var jankPercent float64
	if t.lastFpsTimestamp != 0 {
		dt := float64(t.prevPresentTs-t.lastFpsTimestamp) / float64(time.Second) //The frame rate is calculated once per acquisition cycle
		if dt == 0 {
			fps = 0
		} else {
			fps = int(math.Floor(float64(secData.Fps)/dt + 0.1)) //Only the frame rate deviation within 0.1 is processed rounded up to eliminate the impact of error
		}
		jankPercent = float64(secData.JankTotalTs) * 100 / float64(time.Second) / dt
		if jankPercent > 100 {
			jankPercent = 1.0
		} else if math.IsNaN(jankPercent) {
			jankPercent = 0.0
		}
	}

	ret := map[string]string{
		"fps":         fmt.Sprintf("%d", fps),
		"jank":        fmt.Sprintf("%d", secData.Jank),
		"Bjank":       fmt.Sprintf("%d", secData.BigJank),
		"Sjank":       fmt.Sprintf("%d", secData.SmallJank),
		"jankTime":    fmt.Sprintf("%d", secData.JankTotalTs/1000000),
		"jankPercent": fmt.Sprintf("%.1f", jankPercent),
	}
	t.secOuputFrameData = &OutputFrameData{}
	t.lastFpsTimestamp = t.prevPresentTs
	return ret
}

func (t *SfLatencyStatPlugin) calcFrameTime(frameData *SfFrameData) {
	//Calculate whether there is Jank
	frameData = t.calcFrameJank(frameData)
	//Calculate what is displayed
	t.calcLastSecondFrames(frameData)

	//According to Perfdog algorithm:
	//  - detected Jank: reset the statistics of the last three frames and recalculate jank data
	if frameData.Jank || frameData.BigJank {
		//clear the last 3 frame small jank flag
		if len(t.lastSmallJank3Frames) == 3 {
			frameData.SmallJank = true
			t.secOuputFrameData.SmallJank += 1
		}
		t.lastJank3Frames = make([]*SfFrameData, 0)
		t.lastSmallJank3Frames = make([]*SfFrameData, 0)
	} else {
		//Push in the last three frames of display data to ensure that there are three data
		t.lastJank3Frames = append(t.lastJank3Frames, frameData)
		if len(t.lastJank3Frames) > 3 {
			t.lastJank3Frames = t.lastJank3Frames[len(t.lastJank3Frames)-3 : len(t.lastJank3Frames)]
		}

		//For small jank detect
		frameData = t.calcFrameSmallJank(frameData)
		if frameData.SmallJank {
			t.lastSmallJank3Frames = make([]*SfFrameData, 0)
			t.secOuputFrameData.SmallJank += 1
		} else {
			t.lastSmallJank3Frames = append(t.lastSmallJank3Frames, frameData)
			if len(t.lastSmallJank3Frames) > 3 {
				t.lastSmallJank3Frames = t.lastSmallJank3Frames[len(t.lastSmallJank3Frames)-3 : len(t.lastSmallJank3Frames)]
			}
		}
	}
}

func (t *SfLatencyStatPlugin) calcFrameJank(current *SfFrameData) *SfFrameData {
	if len(t.lastJank3Frames) < 3 { //按照Perfdog算法，未完成3帧统计的数据，不算jank
		return current
	}
	var totalDisplayTimes int64
	for _, v := range t.lastJank3Frames {
		totalDisplayTimes += v.FrameTime
	}
	lastAvgFrameTime := totalDisplayTimes / int64(len(t.lastJank3Frames))

	if current.FrameTime > lastAvgFrameTime*2 && current.FrameTime > int64(84*float64(time.Millisecond)) {
		current.Jank = true
	}
	if current.FrameTime > lastAvgFrameTime*2 && current.FrameTime > int64(125*float64(time.Millisecond)) {
		current.BigJank = true
	}
	if current.Jank || current.BigJank {
		sz := fmt.Sprintf("%d, [", len(t.lastJank3Frames))
		for _, v := range t.lastJank3Frames {
			sz += fmt.Sprintf(" %d,%d", v.DisplayTs, v.FrameTime)
		}
		t.debugLog.Println(current, sz, "]")
	}
	return current
}

func (t *SfLatencyStatPlugin) calcLastSecondFrames(frameData *SfFrameData) {
	t.secOuputFrameData.Fps += 1
	if frameData.Jank {
		t.secOuputFrameData.Jank += 1
	}
	if frameData.BigJank {
		t.secOuputFrameData.BigJank += 1
	}
	if frameData.Jank || frameData.BigJank {
		t.secOuputFrameData.JankTotalTs += frameData.FrameTime
	}
}

func (t *SfLatencyStatPlugin) calcFrameSmallJank(current *SfFrameData) *SfFrameData {
	if len(t.lastSmallJank3Frames) < 3 { //按照Perfdog算法，未完成3帧统计的数据，不算jank
		return current
	}
	var totalDisplayTimes int64
	for _, v := range t.lastSmallJank3Frames {
		totalDisplayTimes += v.FrameTime
	}
	lastAvgFrameTime := totalDisplayTimes / int64(len(t.lastSmallJank3Frames))
	if current.FrameTime > lastAvgFrameTime*2 && current.FrameTime > int64(41*float64(time.Millisecond)) {
		current.SmallJank = true
	}
	return current
}
