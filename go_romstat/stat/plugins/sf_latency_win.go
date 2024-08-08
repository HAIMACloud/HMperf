// Copyright (c) 2021-2023 https://www.haimacloud.com/
// SPDX-License-Identifier: MIT

//go:build windows
// +build windows

package plugins

import (
	"context"
	"errors"
	"fmt"
	"runtime"
	"time"

	"romstat/stat/utils"

	"github.com/kbinani/screenshot"

	"github.com/kirides/go-d3d"
	"github.com/kirides/go-d3d/d3d11"
	"github.com/kirides/go-d3d/outputduplication"
	"github.com/kirides/go-d3d/win"
)

type DxError struct {
	ErrCode int64
	ErrMsg  string
}

func (t *DxError) Error() string {
	return fmt.Sprintf("DirectX Error: %d, %s", t.ErrCode, t.ErrMsg)
}

func NewDesktopFramerateCounter(log utils.Logger, maxFrames *int) *DesktopFramerateCoutner {
	t := new(DesktopFramerateCoutner)
	t.FramesTimestamp = make([]int64, 0)
	if maxFrames == nil {
		t.maxFrames = 500
	} else {
		t.maxFrames = *maxFrames
	}
	t.debugLog = log
	t.ctx = context.Background()
	t.latestPresentTs = 0
	return t
}
func (t *DesktopFramerateCoutner) Stop() {
	t.ctx.Done()
}
func (t *DesktopFramerateCoutner) GetNewFramesTimestamp() []int64 {
	//read present time array for new
	lastIdx := 0
	t.frameTimestampLock.Lock()
	defer t.frameTimestampLock.Unlock()

	if len(t.FramesTimestamp) == 0 {
		return t.FramesTimestamp
	}
	for idx, v := range t.FramesTimestamp {
		if v > t.latestPresentTs {
			lastIdx = idx
			break
		}
	}
	t.latestPresentTs = t.FramesTimestamp[len(t.FramesTimestamp)-1]
	retFrames := append([]int64{}, t.FramesTimestamp[lastIdx:]...)
	t.FramesTimestamp = t.FramesTimestamp[lastIdx:]
	return retFrames
}

func (t *DesktopFramerateCoutner) Start() error {
	max := screenshot.NumActiveDisplays()
	n := max - 1
	//screenBounds := screenshot.GetDisplayBounds(n)
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	// Make thread PerMonitorV2 Dpi aware if supported on OS
	// allows to let windows handle BGRA -> RGBA conversion and possibly more things
	if win.IsValidDpiAwarenessContext(win.DpiAwarenessContextPerMonitorAwareV2) {
		_, err := win.SetThreadDpiAwarenessContext(win.DpiAwarenessContextPerMonitorAwareV2)
		if err != nil {
			t.debugLog.Printf("Could not set thread DPI awareness to PerMonitorAwareV2. %v\n", err)
			return &DxError{ErrCode: -1, ErrMsg: "Could not set thread DPI awareness to PerMonitorAwareV2. " + err.Error()}
		} else {
			t.debugLog.Printf("Enabled PerMonitorAwareV2 DPI awareness.\n")
		}
	}
	// Setup D3D11 stuff
	device, deviceCtx, err := d3d11.NewD3D11Device()
	if err != nil {
		t.debugLog.Printf("Could not create D3D11 Device. %v\n", err)
		return &DxError{ErrCode: -1, ErrMsg: "Could not create D3D11 Device. " + err.Error()}
	}
	defer device.Release()
	defer deviceCtx.Release()

	var ddup *outputduplication.OutputDuplicator
	var outputDupSuccess bool
	for n >= 0 {
		ddup, err = outputduplication.NewIDXGIOutputDuplication(device, deviceCtx, uint(n))
		if err != nil {
			n = n - 1
			continue
		} else {
			outputDupSuccess = true
			break
		}
	}
	if !outputDupSuccess {
		if err != nil {
			t.debugLog.Printf("Err NewIDXGIOutputDuplication: %v\n", err)
			return &DxError{ErrCode: -1, ErrMsg: "Err NewIDXGIOutputDuplication: " + err.Error()}
		}
	}
	defer ddup.Release()
	for {
		select {
		case <-t.ctx.Done():
			return nil
		default:
			break
		}
		// Grab an image.RGBA from the current output presenter
		err := ddup.Try2AcquireNextFrame(100)
		if errors.Is(err, outputduplication.ErrNoImageYet) {
			continue
		}
		if errors.Is(err, outputduplication.ErrDxAccessLost) {
			t.debugLog.Printf("Err ddup.Try2AcquireNextFrame: %v\n", err)
			return &DxError{ErrCode: int64(d3d.DXGI_ERROR_ACCESS_LOST), ErrMsg: "Err ddup.Try2AcquireNextFrame: " + err.Error()}
		}
		if err != nil {
			t.debugLog.Printf("Err ddup.GetImage: %v\n", err)
			return &DxError{ErrCode: -1, ErrMsg: "Err ddup.GetImage: " + err.Error()}
		}
		frameData := time.Now().Local().UnixNano()
		t.frameTimestampLock.Lock()
		if len(t.FramesTimestamp) >= t.maxFrames {
			t.FramesTimestamp = append(t.FramesTimestamp[len(t.FramesTimestamp)-t.maxFrames:], frameData)
		} else {
			t.FramesTimestamp = append(t.FramesTimestamp, frameData)
		}
		t.frameTimestampLock.Unlock()
	}
}

// For Windows Only
func (t *SfLatencyStatPlugin) Open() bool {
	t.secOuputFrameData = &OutputFrameData{}
	t.debugLog = utils.DebugLogger
	t.debugLog.Println("---start---")
	if t.d3dxLoopCounter == nil {
		t.d3dxLoopCounter = NewDesktopFramerateCounter(t.debugLog, nil)
	}
	go func() {
		for {
			err := t.d3dxLoopCounter.Start()
			if err != nil && err.(*DxError).ErrCode != int64(d3d.DXGI_ERROR_ACCESS_LOST) {
				fmt.Println(err)
				break
			}
		}
	}()
	return false
}

func (t *SfLatencyStatPlugin) Close() {
	if t.d3dxLoopCounter != nil {
		t.d3dxLoopCounter.Stop()
	}
}

func (t *SfLatencyStatPlugin) runCollectThread() {
	newSfLatencyDatas := t.d3dxLoopCounter.GetNewFramesTimestamp()
	for _, v := range newSfLatencyDatas {
		actualPresentTime := v
		if t.prevPresentTs == 0 { //Init here
			t.prevPresentTs = actualPresentTime
			continue
		}
		if actualPresentTime > t.prevPresentTs {
			frameData := &SfFrameData{
				DisplayTs: actualPresentTime,
			}
			frameData.FrameTime = actualPresentTime - t.prevPresentTs
			t.calcFrameTime(frameData)
		}
		t.prevPresentTs = actualPresentTime
	}
}

// This method for compatible only
func (t *SfLatencyStatPlugin) GetCurrentPkgSurface() (string, string) {
	return "", ""
}
