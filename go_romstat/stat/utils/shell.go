// Copyright (c) 2021-2023 https://www.haimacloud.com/
// SPDX-License-Identifier: MIT

package utils

import (
	"encoding/base64"
	"errors"
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"

	"github.com/shirou/gopsutil/process"

	"romstat/apk"
)

type AndroidShell struct {
	debugLog Logger
}

func NewAndroidShell() *AndroidShell {
	return &AndroidShell{debugLog: DebugLogger}
}

func (t *AndroidShell) RunShell(command string) string {
	//BUGFIX: Some mobile phone shells are not in/bin/sh, but in/system/bin/sh,
	//so we simply use sh in the environment variable
	//In order not to change the previous code logic, keep/bin/sh as the first choice,
	//and use the sh command by default if it is not found
	var shellCmd = "/bin/sh"
	if !CheckFileIsExist("/bin/sh") {
		shellCmd = "sh"
	}
	if t.debugLog != nil {
		t.debugLog.Println("[CMD]", shellCmd, "-c", command)
	}
	cmd := exec.Command(shellCmd, "-c", command)
	output, _ := cmd.CombinedOutput()
	if t.debugLog != nil {
		t.debugLog.Println("----------")
		t.debugLog.Println(string(output))
	}
	return string(output)
}

func (t *AndroidShell) GetSdkVersion() int64 {
	output := t.RunShell("getprop ro.build.version.sdk")
	output = strings.TrimSpace(output)
	v, err := strconv.ParseInt(output, 10, 64)
	if err != nil {
		panic(err)
	}
	return v
}

func (t *AndroidShell) GetTopmostPackage(sdkVersion int64) string {
	var output string
	if sdkVersion >= 33 {
		output = t.RunShell("dumpsys activity activities |grep topResumedActivity")
	} else {
		output = t.RunShell("dumpsys activity activities |grep mResumedActivity")
	}

	r := regexp.MustCompile("ActivityRecord{(.*)}")
	sz := r.FindStringSubmatch(output)
	if len(sz) <= 1 {
		return ""
	}
	activityName := strings.Split(sz[1], " ")[2]
	return strings.Split(activityName, "/")[0]
}

func (t *AndroidShell) GetPackagePath(pkgName string) string {
	output := t.RunShell("pm list packages -f")
	allPackageLines := strings.Split(output, "\n")
	for _, line := range allPackageLines {
		pkgFile := strings.Join(strings.Split(line, ":")[1:], ":")
		if strings.HasSuffix(pkgFile, pkgName) {
			idx := strings.Index(pkgFile, ".apk")
			if idx > -1 {
				return pkgFile[0:idx] + ".apk"
			}
		}
	}
	return ""
}
func (t *AndroidShell) GetAllInstalledPackages() []string {
	output := t.RunShell("pm list packages")
	allPackageLines := strings.Split(output, "\n")
	installedPackages := make([]string, 0)
	for _, line := range allPackageLines {
		pkgFile := strings.Join(strings.Split(line, ":")[1:], ":")
		installedPackages = append(installedPackages, pkgFile)
	}
	return installedPackages
}

type PackageInfo struct {
	Name        string `json:"name"`
	Icon        string `json:"icon"`
	PackageName string `json:"package_name"`
}

func (t *AndroidShell) GetPackageInfo(pkgName string) (*PackageInfo, error) {
	if pkgName == "" {
		sdkVersion := t.GetSdkVersion()
		pkgName = t.GetTopmostPackage(sdkVersion)
	}
	pkgFilePath := t.GetPackagePath(pkgName)
	pkg, err := apk.OpenFile(pkgFilePath)
	if err != nil {
		return nil, err
	}
	defer pkg.Close()

	imgByte, err := pkg.IconJpeg(nil)
	baseImg := ""
	if err == nil {
		baseImg = "data:image/jpeg;base64," + base64.StdEncoding.EncodeToString(imgByte)
	}
	pkgLabel, _ := pkg.Label(nil)
	return &PackageInfo{
		Name:        pkgLabel,
		PackageName: pkgName,
		Icon:        baseImg,
	}, nil
}

type RunningPackageInfo struct {
	Name    string `json:"name"`
	Pid     int32  `json:"pid"`
	Topmost bool   `json:"topmost"`
}

func (t *AndroidShell) GetRecentApps() ([]string, error) {
	output := t.RunShell(" dumpsys activity recents |grep mRootProcess")
	r := regexp.MustCompile("ProcessRecord{(.*)}")
	retLst := r.FindAllStringSubmatch(output, -1)
	if len(retLst) <= 1 {
		return nil, errors.New("not found")
	}
	recentApps := make([]string, 0)
	for _, v := range retLst {
		if len(v) <= 1 {
			continue
		}
		startIdx := strings.Index(v[1], ":") + 1
		endIdx := strings.Index(v[1], "/")
		if startIdx < 0 || endIdx < 0 {
			continue
		}
		pkgName := v[1][startIdx:endIdx]
		recentApps = append(recentApps, pkgName)
	}
	return recentApps, nil
}
func (t *AndroidShell) GetAllRunningPackages() ([]*RunningPackageInfo, error) {
	allRecentPackages, err := t.GetRecentApps()
	if err != nil {
		allRecentPackages = t.GetAllInstalledPackages()
	}
	topMostPackage := t.GetTopmostPackage(t.GetSdkVersion())
	allRecentPackages = append(allRecentPackages, topMostPackage)
	processes, err := process.Processes()
	if err != nil {
		return nil, err
	}
	allRunningPackages := make([]*RunningPackageInfo, 0)
	for _, p := range processes {
		processName, _ := p.Name()
		if StringInSlice(processName, allRecentPackages) {
			isTopmost := false
			if topMostPackage == processName {
				isTopmost = true
			}
			allRunningPackages = append(allRunningPackages, &RunningPackageInfo{
				Name:    processName,
				Pid:     p.Pid,
				Topmost: isTopmost,
			})
		}
	}

	return allRunningPackages, nil
}

func (t *AndroidShell) GetPackagePid(pkgName string) (int32, error) {
	processes, err := process.Processes()
	if err != nil {
		return 0, err
	}
	for _, p := range processes {
		processName, _ := p.Name()
		if pkgName == processName {
			return p.Pid, nil
		}

	}
	return 0, errors.New("process not found")
}

type PingStat struct {
	SendPackages int
	RecvPackages int
	RssLst       []float64
}

func (t *AndroidShell) GetPingStat(domainAddr string, packCount int) (*PingStat, error) {
	output := t.RunShell(fmt.Sprintf("ping -c %d -i 0.2 -W 1000 %s", packCount, domainAddr))
	rRss := regexp.MustCompile("time=(.*) ms")
	matchLst := rRss.FindAllStringSubmatch(output, -1)
	retPingStat := &PingStat{
		SendPackages: packCount,
	}
	if len(matchLst) >= 1 {
		rssLst := make([]float64, 0)
		for _, match := range matchLst {
			rss, err := strconv.ParseFloat(match[1], 64)
			if err != nil {
				continue
			}
			rssLst = append(rssLst, rss)
		}
		retPingStat.RssLst = rssLst
	}

	rSendRecv := regexp.MustCompile("packets transmitted, (.*) received")
	recvPackages := rSendRecv.FindStringSubmatch(output)
	if len(recvPackages) >= 1 {
		retPingStat.SendPackages = packCount
		recvPacks, _ := strconv.ParseInt(recvPackages[1], 10, 64)
		retPingStat.RecvPackages = int(recvPacks)
	}
	return retPingStat, nil
}
