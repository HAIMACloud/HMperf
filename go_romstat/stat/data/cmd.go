// Copyright (c) 2021-2023 https://www.haimacloud.com/
// SPDX-License-Identifier: MIT

package data

import (
	"flag"
)

type CmdlineParameters struct {
	Pid       int
	IsDebug   bool
	PemFile   string
	IsVersion bool
	IsPInfo   bool
	PkgName   string
}

var cmdParameters CmdlineParameters

func InitCmdParser() {
	flag.IntVar(&cmdParameters.Pid, "p", 0, "process id, default all system(0)")
	flag.BoolVar(&cmdParameters.IsDebug, "d", false, "is debug mode, default false")
	flag.StringVar(&cmdParameters.PemFile, "pem", "", "pem file path, ras pubkey file")
	flag.BoolVar(&cmdParameters.IsVersion, "v", false, "print version information")
	flag.BoolVar(&cmdParameters.IsPInfo, "pinfo", false, "print package information, default topmost package")
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
