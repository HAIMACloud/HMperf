// Copyright (c) 2021-2023 https://www.haimacloud.com/
// SPDX-License-Identifier: MIT

//go:build windows
// +build windows

package plugins

import (
	"romstat/stat/utils"
	"time"

	"github.com/go-ping/ping"
)

func GetPingStat(count int) (*utils.PingStat, error) {
	stat := new(utils.PingStat)
	stat.RssLst = make([]float64, 0)
	pinger, err := ping.NewPinger("www.baidu.com")
	if err != nil {
		return nil, err
	}
	pinger.SetPrivileged(true)
	pinger.Count = count
	pinger.Interval = 200 * time.Millisecond
	pinger.Timeout = 1 * time.Second
	pinger.OnRecv = func(p *ping.Packet) {
		stat.RssLst = append(stat.RssLst, float64(p.Rtt.Abs().Seconds()*1000))
	}
	pinger.OnFinish = func(stats *ping.Statistics) {
		stat.RecvPackages = stats.PacketsRecv
		stat.SendPackages = stats.PacketsSent
	}

	err = pinger.Run()
	if err != nil {
		panic(err)
	}
	return stat, nil
}
