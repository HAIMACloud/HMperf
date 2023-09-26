// Copyright (c) 2021-2023 https://www.haimacloud.com/
// SPDX-License-Identifier: MIT

package data

import (
	"flag"

	"github.com/shirou/gopsutil/process"
)

type CmdlineParameters struct {
	IsDebug       bool
	PemFile       string
	IsVersion     bool
	IsPInfo       bool
	IsListRunning bool
	PkgName       string
	TargetSurface string
	LockSurface   bool
	Ask           string
}

func (t *CmdlineParameters) getPkgRunningPid() int32 {
	allProcesses, err := process.Processes()
	if err != nil {
		return 0
	}
	for _, p := range allProcesses {
		pkgName, err := p.Name()
		if err != nil {
			continue
		}
		if pkgName == t.PkgName {
			return p.Pid
		}
	}
	return 0
}
func (t *CmdlineParameters) GetPid() int32 {
	return t.getPkgRunningPid()
}

var cmdParameters CmdlineParameters

func InitCmdParser() {
	flag.StringVar(&cmdParameters.PkgName, "p", "", "application package name, default all system")
	flag.BoolVar(&cmdParameters.IsDebug, "d", false, "is debug mode, default false")
	flag.StringVar(&cmdParameters.TargetSurface, "ts", "", "specify target surface, default for auto")
	flag.BoolVar(&cmdParameters.LockSurface, "lock", false, "lock collected package surface, cannot transfer to application")
	flag.StringVar(&cmdParameters.PemFile, "pem", "", "pem file path, ras pubkey file")
	flag.BoolVar(&cmdParameters.IsVersion, "v", false, "print version information")
	flag.BoolVar(&cmdParameters.IsPInfo, "pinfo", false, "print package information, default topmost package")
	flag.BoolVar(&cmdParameters.IsListRunning, "running", false, "print all running package name")
	flag.StringVar(&cmdParameters.Ask, "ask", "", "ask for master process from pipeline: current_pkg_surface")
	flag.Parse()
	if cmdParameters.IsPInfo {
		if len(flag.Args()) >= 1 {
			cmdParameters.PkgName = flag.Args()[0]
		}
	}
}

func GetCmdParameters() *CmdlineParameters {
	return &cmdParameters
}
