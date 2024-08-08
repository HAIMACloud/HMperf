// Copyright (c) 2021-2023 https://www.haimacloud.com/
// SPDX-License-Identifier: MIT

package plugins

import (
	"fmt"
	"runtime"
	"time"

	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/mem"
	"github.com/shirou/gopsutil/process"

	"romstat/stat/data"
)

type MemoryStatData struct {
	UsedPercent float64
	SwapCached  uint64
}

type SystemStatPlugin struct {
	cpuUsage   float64
	memPercent float64
	swapCached uint64

	cpuUsageCh chan float64
	memInfoCh  chan *MemoryStatData

	collectSecTime time.Duration
}

func (c *SystemStatPlugin) Open() bool {
	c.cpuUsageCh = make(chan float64)
	c.memInfoCh = make(chan *MemoryStatData)
	c.collectSecTime = time.Millisecond * 500
	return true
}

func (c *SystemStatPlugin) cpuStat() {
	for {
		if pid := data.GetCmdParameters().GetPid(); pid != 0 {
			ps, _ := process.NewProcess(pid)
			percent, err := ps.Percent(c.collectSecTime)
			if err != nil {
				panic("cpu collect error")
			}
			cpuUsagePercent := percent / float64(runtime.NumCPU())
			c.cpuUsageCh <- cpuUsagePercent
		} else {
			percent, err := cpu.Percent(c.collectSecTime, false)
			if err != nil {
				panic("cpu collect error")
			}
			c.cpuUsageCh <- percent[0]
		}
	}
}

func (c *SystemStatPlugin) memStat() {
	for {
		if pid := data.GetCmdParameters().GetPid(); pid != 0 {
			ps, _ := process.NewProcess(pid)
			memPercent, _ := ps.MemoryPercent()
			memInfo, err := ps.MemoryInfo()
			if err != nil {
				panic("memory collect error")
			}
			time.Sleep(c.collectSecTime)
			c.memInfoCh <- &MemoryStatData{
				UsedPercent: float64(memPercent),
				SwapCached:  memInfo.Swap,
			}
		} else {
			memInfo, err := mem.VirtualMemory()
			if err != nil {
				panic("memory collect error")
			}
			time.Sleep(c.collectSecTime)
			c.memInfoCh <- &MemoryStatData{
				UsedPercent: memInfo.UsedPercent,
				SwapCached:  memInfo.SwapCached,
			}
		}
	}
}

func (c *SystemStatPlugin) Run() {
	go c.cpuStat()
	go c.memStat()
	go func() {
		for {
			select {
			case cpuUsage := <-c.cpuUsageCh:
				c.cpuUsage = cpuUsage
			case memUsage := <-c.memInfoCh:
				c.memPercent = memUsage.UsedPercent
				c.swapCached = memUsage.SwapCached
			}
		}
	}()
}

func (c *SystemStatPlugin) GetTypes() []*data.PluginType {
	return []*data.PluginType{
		{Name: "cpu_usg", DisplayName: "cpu%", IsCmdShow: true},
		{Name: "mem_usg", DisplayName: "mem%", IsCmdShow: true},
		{Name: "mem_swap", DisplayName: "swap/MB", IsCmdShow: true},
	}
}

func (c *SystemStatPlugin) GetData() map[string]string {
	c.collectSecTime = time.Second
	return map[string]string{
		"cpu_usg":  fmt.Sprintf("%.1f", c.cpuUsage),
		"mem_usg":  fmt.Sprintf("%.1f", c.memPercent),
		"mem_swap": fmt.Sprintf("%.2f", float64(c.swapCached)/1024/1024),
	}
}

func (c *SystemStatPlugin) Close() {

}
