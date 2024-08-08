// Copyright (c) 2021-2023 https://www.haimacloud.com/
// SPDX-License-Identifier: MIT

//go:build linux || android || darwin
// +build linux android darwin

package plugins

import "romstat/stat/utils"

func GetPingStat(count int) (*utils.PingStat, error) {
	shell := utils.NewAndroidShell()
	return shell.GetPingStat("www.baidu.com", count)
}
