// Copyright (c) 2021-2023 https://www.haimacloud.com/
// SPDX-License-Identifier: MIT

package utils

import (
	"encoding/base64"
	"os/exec"
	"regexp"
	"strconv"
	"strings"

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
	if sdkVersion == 33 {
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
