// Copyright (c) 2021-2023 https://www.haimacloud.com/
// SPDX-License-Identifier: MIT

//go:build android || linux || darwin
// +build android linux darwin

package plugins

import (
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"romstat/stat/data"
	"romstat/stat/utils"
)

func (t *SfLatencyStatPlugin) guessH5SurfaceView(pkgName string, allSurfaceList string) string {
	surfaceList := make([]string, 0)
	surfaceList = append(surfaceList, strings.Split(allSurfaceList, "\n")...)
	if pkgName == "com.tencent.mm" {
		for _, s := range surfaceList {
			if strings.HasPrefix(s, "com.tencent.mm/com.tencent.mm.plugin.webview.ui.tools.MMWebViewUI#") {
				return s
			}
		}
	} else if pkgName == "com.android.chrome" {
		for _, s := range surfaceList {
			if strings.HasPrefix(s, "com.android.chrome/ChromeChildSurface#") {
				return s
			}
		}
	}
	return ""
}
func (t *SfLatencyStatPlugin) guessSpecialAppView(pkgName string, allSurfaceList string) string {
	surfaceList := make([]string, 0)
	surfaceList = append(surfaceList, strings.Split(allSurfaceList, "\n")...)
	if pkgName == "com.ss.android.ugc.aweme" {
		for _, s := range surfaceList {
			if strings.HasPrefix(s, "com.ss.android.ugc.aweme/com.ss.android.ugc.aweme.splash.SplashActivity") {
				return s
			}
		}
	}
	return ""
}
func (t *SfLatencyStatPlugin) GetCurrentPkgSurface() (string, string) {
	return t.currentPkgName, t.currentSurfaceView
}
func (t *SfLatencyStatPlugin) getLockedSurfaceView() (string, error) {
	if t.lockedPkgSurface == nil {
		pkgName := t.shell.GetTopmostPackage(t.sdkVersion)
		t.debugLog.Println("topmost package: ", pkgName)
		if (data.GetCmdParameters().PkgName != "") && pkgName != t.monitorProcessName {
			return "", errors.New("process not match, monitor process name: " + t.monitorProcessName)
		}
		if t.lockedPkgSurface == nil {
			t.lockedPkgSurface = new(SfPkgSurfaceData)
		}
		t.lockedPkgSurface.PkgName = pkgName
		t.lockedPkgSurface.SurfaceView, _ = t.guessSurfaceView2(pkgName)
	}
	return t.lockedPkgSurface.SurfaceView, nil
}
func (t *SfLatencyStatPlugin) getLastLine(lines []string) string {
	var lastLine string
	idx := 0
	for {
		idx += 1
		lastCount := len(lines) - idx
		if lastCount < 0 {
			break
		}
		lastLine = lines[lastCount]
		if strings.TrimSpace(lastLine) != "" {
			break
		}
	}
	return lastLine
}
func (t *SfLatencyStatPlugin) guessSurfaceView2(pkgName string) (string, error) {
	t.debugLog.Println("guessSurfaceView2: start guess, pkgName=", pkgName)
	output := t.shell.RunShell("dumpsys SurfaceFlinger --list")
	pkgSurfaceViewLst := make([]string, 0)
	for _, surfaceView := range strings.Split(output, "\n") {
		if strings.Contains(surfaceView, pkgName) {
			pkgSurfaceViewLst = append([]string{surfaceView}, pkgSurfaceViewLst...)
		}
	}

	guessTargetViewMap := make(map[string]int64)
	//first record
	for _, sfView := range pkgSurfaceViewLst {
		latencyOutput := t.shell.RunShell(fmt.Sprintf("dumpsys SurfaceFlinger --latency '%s'", sfView))
		lines := strings.Split(latencyOutput, "\n")
		if len(lines) > 1 {
			var lastLine = t.getLastLine(lines)
			dataLst := strings.Split(lastLine, "\t")
			if len(dataLst) != 3 {
				continue
			}
			vals := make([]int64, 0)
			for _, v := range dataLst {
				val, _ := strconv.ParseInt(strings.TrimSpace(v), 10, 64)
				vals = append(vals, val)
			}
			if vals[0] == 0 {
				continue
			}
			guessTargetViewMap[sfView] = vals[0]
		}
	}
	time.Sleep(50 * time.Millisecond) //sleep 50 ms
	for sfView, val := range guessTargetViewMap {
		latencyOutput := t.shell.RunShell(fmt.Sprintf("dumpsys SurfaceFlinger --latency '%s'", sfView))
		lines := strings.Split(latencyOutput, "\n")
		if len(lines) > 1 {
			var lastLine = t.getLastLine(lines)
			dataLst := strings.Split(lastLine, "\t")
			if len(dataLst) != 3 {
				continue
			}
			vals := make([]int64, 0)
			for _, v := range dataLst {
				val, _ := strconv.ParseInt(strings.TrimSpace(v), 10, 64)
				vals = append(vals, val)
			}
			lastVsync := vals[0]
			if lastVsync > val { //vsync changed, guess this surface view is correct
				t.debugLog.Println("guessSurfaceView2: guess surface view is target:", sfView)
				return sfView, nil
			}
		}
	}
	return t.guessSurfaceView(pkgName)
}
func (t *SfLatencyStatPlugin) guessSurfaceView(pkgName string) (string, error) {
	t.debugLog.Println("guessSurfaceView: start guess, pkgName=", pkgName)
	output := t.shell.RunShell("dumpsys SurfaceFlinger --list")
	//BUGFIX： The Redmi mobile cannot get fps,
	//The Redmi mobile phone use SurfaceView named end with BLAST
	//This situation needs special treatment
	SurfaceViewResults := make([]string, 0)

	//For the special handling of some special surfaces, specify the surface acquisition conditions
	if data.GetCmdParameters().TargetSurface != "" {
		targetSurface := data.GetCmdParameters().TargetSurface
		for _, line := range strings.Split(output, "\n") {
			if pkgName != "" && strings.Contains(line, pkgName) && strings.Contains(line, targetSurface) {
				return line, nil
			}
		}
	}

	//BUGFIX: guess h5 view surface for WeiChat and Chrome browser
	if gSurface := t.guessH5SurfaceView(pkgName, output); gSurface != "" {
		return gSurface, nil
	}
	//BUGFIX: special apps view
	if gSurface := t.guessSpecialAppView(pkgName, output); gSurface != "" {
		return gSurface, nil
	}

	//BUGFIX: Add AppViewResults, to support use the last pkg activity
	AppViewResults := make([]string, 0)
	blastSurfaceViewResults := make([]string, 0)

	for _, line := range strings.Split(output, "\n") {
		if t.sdkVersion < 24 && line == "SurfaceView" {
			SurfaceViewResults = append(SurfaceViewResults, line)
		}
		if strings.HasPrefix(line, "SurfaceView") && strings.Index(line, pkgName) > 0 { //BUGFIX： 红米机器获取不到fps
			if strings.Index(line, "BLAST") > 0 {
				blastSurfaceViewResults = append(blastSurfaceViewResults, line)
			} else {
				SurfaceViewResults = append(SurfaceViewResults, line)
			}
		}
	}
	//BUGFIX: support 2 blast surface views return the last one
	if len(blastSurfaceViewResults) > 0 {
		return blastSurfaceViewResults[len(blastSurfaceViewResults)-1], nil
	}
	//BUGFIX: Fix the fps problem because the package name is obtained first and then the SurfaceView
	if len(SurfaceViewResults) > 0 {
		return SurfaceViewResults[len(SurfaceViewResults)-1], nil
	}
	for _, line := range strings.Split(output, "\n") {
		if pkgName != "" && strings.HasPrefix(line, pkgName) {
			AppViewResults = append(AppViewResults, line)
		}
	}
	if len(AppViewResults) > 0 {
		return AppViewResults[len(AppViewResults)-1], nil
	}
	return "", errors.New("no surfaceview found")
}
func (t *SfLatencyStatPlugin) getTopSurfaceView() (string, error) {
	pkgName := t.shell.GetTopmostPackage(t.sdkVersion)
	t.debugLog.Println("topmost package: ", pkgName)
	if (data.GetCmdParameters().PkgName != "") && pkgName != t.monitorProcessName {
		return "", errors.New("process not match, monitor process name: " + t.monitorProcessName)
	}
	if t.currentPkgName != pkgName {
		t.currentPkgName = pkgName
		return t.guessSurfaceView2(pkgName)
	}
	return t.guessSurfaceView(pkgName)
}

func (t *SfLatencyStatPlugin) Open() bool {
	if data.GetCmdParameters().PkgName != "" {
		t.monitorProcessName = data.GetCmdParameters().PkgName
	}
	t.shell = utils.NewAndroidShell()
	t.sdkVersion = t.shell.GetSdkVersion()
	t.secOuputFrameData = &OutputFrameData{}
	t.debugLog = utils.DebugLogger
	t.debugLog.Println("---start---")
	return true
}

func (t *SfLatencyStatPlugin) Close() {
}

func (t *SfLatencyStatPlugin) getSFLatencyData() [][]int64 {
	ret := make([][]int64, 0)
	output := t.shell.RunShell(fmt.Sprintf("dumpsys SurfaceFlinger --latency '%s'", t.currentSurfaceView))
	lines := strings.Split(output, "\n")
	t.vSyncPeriod, _ = strconv.ParseInt(lines[0], 10, 64) //记录帧间隔数据

	for _, line := range lines[1:] {
		if strings.TrimSpace(line) == "" {
			continue
		}
		dataLst := strings.Split(line, "\t")
		//BUGFIX: Judge whether it is 3 frames of data to ensure that the subsequent data obtained is accurate
		if len(dataLst) != 3 {
			continue
		}
		vals := make([]int64, 0)
		for _, v := range dataLst {
			val, _ := strconv.ParseInt(strings.TrimSpace(v), 10, 64)
			vals = append(vals, val)
		}
		ret = append(ret, vals)
	}
	if len(ret) == 0 { //If the above method does not get data, use 'gfxinfo framestats' to get data
		if packageName := t.shell.GetTopmostPackage(t.sdkVersion); packageName != "" {
			output := t.shell.RunShell(fmt.Sprintf("dumpsys gfxinfo %s framestats", packageName))
			lines := strings.Split(output, "\n")
			findTimestamps := false
			for _, line := range lines {
				if !findTimestamps {
					if strings.Contains(line, "---PROFILEDATA---") {
						findTimestamps = true
						continue
					}
				} else {
					if strings.TrimSpace(line) == "" {
						continue
					}
					dataLst := strings.Split(line, " ")
					if len(dataLst) != 3 {
						continue
					}
					vals := make([]int64, 0)
					for _, v := range dataLst {
						val, _ := strconv.ParseInt(strings.TrimSpace(v), 10, 64)
						vals = append(vals, val)
					}
					ret = append(ret, vals)
				}
			}
		}
	}

	return ret
}

func (t *SfLatencyStatPlugin) refreshSFLatencyData(SurfaceChanged bool) [][]int64 {
	currentLatencyData := t.getSFLatencyData()
	sfTimestamps := make([][]int64, 0)
	//In two cases, the frame rate needs to be recalculated
	//1. No previous Vsync frame record (first record)
	//2. The Vsync frame has not been obtained for more than 1s :
	//   - Vsync frames may be lost such as switching back after 1s of screen lock
	if t.prevMaxVsyncTimestamp == 0 ||
		(SurfaceChanged && len(currentLatencyData) > 0 && currentLatencyData[len(currentLatencyData)-1][1]-t.prevMaxVsyncTimestamp > int64(time.Second)) {
		t.debugLog.Println("reset params: ", currentLatencyData[len(currentLatencyData)-1][0], t.prevMaxVsyncTimestamp, SurfaceChanged, len(currentLatencyData))
		for i := 1; i <= len(currentLatencyData); i++ { //Calculate the last legal data as the last vsync frame
			if currentLatencyData[len(currentLatencyData)-i][1] != math.MaxInt64 {
				t.prevMaxVsyncTimestamp = currentLatencyData[len(currentLatencyData)-i][1]
				break
			}
		}
		t.lastJank3Frames = []*SfFrameData{}
		t.lastSmallJank3Frames = []*SfFrameData{}
		t.prevPresentTs = 0 //Not been processed for a long time, exit
		return sfTimestamps
	}
	for _, model := range currentLatencyData {
		//If it has been recorded before, it is not necessary to record again
		if t.prevMaxVsyncTimestamp != 0 && t.prevMaxVsyncTimestamp >= model[1] {
			t.prevPresentTs = model[1] //last validate present frame
			continue
		}
		if model[1] == math.MaxInt64 { //Illegal data, representing incomplete rendering
			continue
		}
		sfTimestamps = append(sfTimestamps, model)
	}
	if len(sfTimestamps) == 0 {
		return sfTimestamps
	}
	t.prevMaxVsyncTimestamp = sfTimestamps[len(sfTimestamps)-1][1]
	return sfTimestamps
}

func (t *SfLatencyStatPlugin) runCollectThread() {
	var newSfLatencyDatas [][]int64
	if !data.GetCmdParameters().LockSurface {
		oldSurfaceView := t.currentSurfaceView
		t.currentSurfaceView, _ = t.getTopSurfaceView()
		if t.currentSurfaceView == "" {
			return
		}
		newSfLatencyDatas = t.refreshSFLatencyData(oldSurfaceView != t.currentSurfaceView)
	} else {
		t.currentSurfaceView, _ = t.getLockedSurfaceView()
		if t.currentSurfaceView == "" {
			return
		}
		newSfLatencyDatas = t.refreshSFLatencyData(false)
	}

	//Calculate on-frame screen time
	for idx, v := range newSfLatencyDatas {
		actualPresentTime := v[1]
		t.debugLog.Println(idx,
			fmt.Sprintf("+%d", (actualPresentTime-t.prevPresentTs)/1000000),
			fmt.Sprintf("%d %d %d", v[0], v[1], v[2]))

		if t.prevPresentTs == 0 { //Init here
			t.prevPresentTs = actualPresentTime
			continue
		}

		//Judge whether it is a new swap buffer frame
		//	Yes: press the queue to be displayed
		if (actualPresentTime / 1000000) >= (t.prevPresentTs / 1000000) {
			frameData := &SfFrameData{
				DisplayTs:    actualPresentTime,
				RefTimestamp: v,
			}
			//Calculate display duration
			frameData.FrameTime = actualPresentTime - t.prevPresentTs
			t.calcFrameTime(frameData)
		}
		t.prevPresentTs = actualPresentTime
	}
}
