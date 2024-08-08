// Copyright (c) 2021-2023 https://www.haimacloud.com/
// SPDX-License-Identifier: MIT

package plugins

import (
	"fmt"
	"runtime"
	"strings"
	"time"

	"github.com/shirou/gopsutil/net"
	"github.com/shirou/gopsutil/process"

	"romstat/stat/data"
	"romstat/stat/utils"
)

type NetData struct {
	BytesSend uint64
	BytesRecv uint64
}

type NetworkStatPlugin struct {
	netInfo         []net.IOCountersStat
	lastTimestamp   int64
	lastNetStatData map[int]*NetData
	sendPerSec      float64
	recvPerSec      float64
}

func (t *NetworkStatPlugin) Open() bool {
	return true
}

func (t *NetworkStatPlugin) Close() {
}

func (t *NetworkStatPlugin) Run() {
	if t.lastNetStatData == nil {
		t.lastNetStatData = make(map[int]*NetData)
	}
	t.netInfo, _ = net.IOCounters(true)
	for idx, v := range t.netInfo {
		t.lastNetStatData[idx] = &NetData{BytesSend: v.BytesSent, BytesRecv: v.BytesRecv}
	}
	t.lastTimestamp = time.Now().UnixNano()
	go utils.SetTimer(1, t.goNetworkStatBySeconds)
}

func (t *NetworkStatPlugin) GetTypes() []*data.PluginType {
	return []*data.PluginType{
		{Name: "net_in", DisplayName: "in(KB)", IsCmdShow: true},
		{Name: "net_out", DisplayName: "out(KB)", IsCmdShow: true},
	}
}

func (t *NetworkStatPlugin) goNetworkStatBySeconds() {
	oldTs := t.lastTimestamp
	var sendDert, recvDert uint64
	if pid := data.GetCmdParameters().GetPid(); pid != 0 {
		//Network data related for given process
		ps, _ := process.NewProcess(pid)
		t.netInfo, _ = ps.NetIOCounters(true)
	} else {
		t.netInfo, _ = net.IOCounters(true)
	}
	for idx, v := range t.netInfo {
		if _, ok := t.lastNetStatData[idx]; ok {
			if runtime.GOOS != "windows" {
				//BUGFIX:
				//Only retain the rmnet and wlan, except the ccmni Network
				//Avoid double counting on some mobile phones
				if !strings.HasPrefix(v.Name, "rmnet_data") &&
					!strings.HasPrefix(v.Name, "ccmni") &&
					!strings.HasPrefix(v.Name, "wlan") {
					continue
				}
				//BUGFIX: log collected network interface for debug information
				utils.DebugLogger.Println("NETWORK", v.Name, !strings.HasPrefix(v.Name, "rmnet_data"),
					!strings.HasPrefix(v.Name, "ccmni"),
					!strings.HasPrefix(v.Name, "wlan"))
			} else {
				if strings.HasPrefix(v.Name, "Loopback Pseudo-Interface") {
					continue
				}
			}
			sendDert += v.BytesSent - t.lastNetStatData[idx].BytesSend
			recvDert += v.BytesRecv - t.lastNetStatData[idx].BytesRecv
			t.lastNetStatData[idx] = &NetData{BytesSend: v.BytesSent, BytesRecv: v.BytesRecv}
		}
	}
	t.lastTimestamp = time.Now().UnixNano()

	timeDert := float64(t.lastTimestamp-oldTs) / float64(time.Second)
	t.sendPerSec = float64(sendDert) / timeDert
	t.recvPerSec = float64(recvDert) / timeDert
}

func (t *NetworkStatPlugin) GetData() map[string]string {
	return map[string]string{
		"net_in":  fmt.Sprintf("%.6f", t.recvPerSec/1024)[0:6],
		"net_out": fmt.Sprintf("%.6f", t.sendPerSec/1024)[0:6],
	}
}
