// Copyright (c) 2021-2023 https://www.haimacloud.com/
// SPDX-License-Identifier: MIT

package stat

import (
	"fmt"

	"romstat/stat/plugins"
)

var registerPlugins map[string]Plugin

type MemDataItem struct {
	TimeStamp int64
	ItemData  map[string]string
}

var AllPluginMonitorItem []string

func RegPlugin(name string, plugin Plugin) {
	if registerPlugins == nil {
		registerPlugins = make(map[string]Plugin)
	}
	registerPlugins[name] = plugin

	for _, t := range plugin.GetTypes() {
		AllPluginMonitorItem = append(AllPluginMonitorItem, fmt.Sprintf("%s.%s", name, t.Name))
	}
}

func LoadAllPlugins() {
	if AllPluginMonitorItem == nil {
		AllPluginMonitorItem = make([]string, 0)
	}

	RegPlugin("system", new(plugins.SystemStatPlugin))
	RegPlugin("display", new(plugins.SfLatencyStatPlugin))
	RegPlugin("network", new(plugins.NetworkStatPlugin))
}

func UnloadPlugins() {
	for pluginName := range registerPlugins {
		registerPlugins[pluginName].Close()
	}
}
