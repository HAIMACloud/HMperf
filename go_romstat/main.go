// Copyright (c) 2021-2023 https://www.haimacloud.com/
// SPDX-License-Identifier: MIT

package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	data2 "romstat/build"
	"romstat/stat"
	"romstat/stat/data"
	"romstat/stat/utils"
)

func main() {
	data.InitCmdParser()
	if data.GetCmdParameters().IsVersion {
		fmt.Println("hmp:", data2.HmFileVersion)
		fmt.Println("romstat:", data2.RomStatVersion)
		return
	}
	if data.GetCmdParameters().IsPInfo {
		if data.GetCmdParameters().PkgName == "" {
			data.GetCmdParameters().PkgName = utils.NewAndroidShell().GetTopmostPackage(utils.NewAndroidShell().GetSdkVersion())
		}
		if data.GetCmdParameters().PkgName == "" {
			fmt.Println("ERROR: cannot find package to get information")
			return
		}
		pkgInfo, err := utils.NewAndroidShell().GetPackageInfo(data.GetCmdParameters().PkgName)
		if err != nil {
			fmt.Println("ERROR:", err.Error())
			return
		}
		info, _ := json.MarshalIndent(pkgInfo, "", "  ")
		fmt.Println(string(info))
		return
	} else if data.GetCmdParameters().IsListRunning {
		pkgInfo, err := utils.NewAndroidShell().GetAllRunningPackages()
		if err != nil {
			fmt.Println("ERROR:", err.Error())
			return
		}
		info, _ := json.MarshalIndent(pkgInfo, "", "  ")
		fmt.Println(string(info))
		return
	} else if data.GetCmdParameters().Ask != "" {
		answer, err := stat.AskPipelineServer(data.GetCmdParameters().Ask)
		if err != nil {
			fmt.Println("ERROR:", err.Error())
			return
		}
		fmt.Println(answer)
		return
	}

	utils.InitLogger()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	mgmt := stat.InitStatByType([]string{"system", "display", "network", "ping"})
	go mgmt.Start(1)
	defer stat.UnloadPlugins()

	go stat.NewPipelineServerListen()
	<-c
}
