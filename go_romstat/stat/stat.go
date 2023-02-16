// Copyright (c) 2021-2023 https://www.haimacloud.com/
// SPDX-License-Identifier: MIT

package stat

import (
	"fmt"
	"os"
	"strings"
	"time"

	"romstat/stat/data"
	"romstat/stat/utils"
)

type ItemData struct {
	TimeStamp int64
	Data      map[string]map[string]string
}

type Header struct {
	TypeLst     []string
	PluginTypes []*data.PluginType
}

type PluginManager struct {
	data            map[string]Plugin
	currentRunTypes []string
	itemDataChan    chan *ItemData
	header          *Header
	displayLogger   utils.Logger
	debugLogger     utils.Logger
}

var sep = "\t"
var csvSep = ","
var lineNum = 30
var headerWriteFlag = false

var fpWriter *utils.RsaWriter

type Plugin interface {
	Open() bool
	Close()
	Run()
	GetTypes() []*data.PluginType
	GetData() map[string]string
}

func (t *PluginManager) outputHeaderLines() {
	lines := []string{"--------"}
	for _, v := range t.header.TypeLst {
		if v == "system" {
			lines = append(lines, "   --------system-------")
		} else if v == "display" {
			lines = append(lines, "-----------display-----------")
		} else if v == "network" {
			lines = append(lines, "----network----")
		}
	}
	t.displayLogger.Println(strings.Join(lines, "    "))
	cmdPluginTypes := make([]string, 0)
	allPluginTypes := make([]string, 0)
	for _, tp := range t.header.PluginTypes {
		if tp.IsCmdShow {
			cmdPluginTypes = append(cmdPluginTypes, tp.DisplayName)
		}
		allPluginTypes = append(allPluginTypes, tp.DisplayName)
	}
	t.displayLogger.Println("  时间  " + sep + strings.Join(cmdPluginTypes, sep))
	if !headerWriteFlag {
		fpWriter.WriteString("时间" + csvSep)
		fpWriter.WriteString(strings.Join(allPluginTypes, csvSep) + "\n")
		fpWriter.Flush()
		headerWriteFlag = true
	}
}

func initFpOutput() {
	var pubKey *string
	if data.GetCmdParameters().PemFile != "" {
		fContent, err := os.ReadFile(data.GetCmdParameters().PemFile)
		if err != nil {
			panic(err)
		}
		pemContent := string(fContent)
		pubKey = &pemContent
	}
	fpWriter = utils.NewRsaWriter("/data/local/tmp/out.hmp", pubKey)
}

func InitStatByType(typeLst []string) *PluginManager {
	LoadAllPlugins()

	initFpOutput()

	mgmt := new(PluginManager)
	mgmt.displayLogger = utils.DisplayLogger
	mgmt.debugLogger = utils.DebugLogger
	mgmt.currentRunTypes = typeLst
	mgmt.data = make(map[string]Plugin)
	pluginTypes := make([]*data.PluginType, 0)
	for _, pluginName := range typeLst {
		plugin := registerPlugins[pluginName]
		plugin.Open()
		pluginTypes = append(pluginTypes, plugin.GetTypes()...)
		mgmt.data[pluginName] = plugin
	}
	mgmt.itemDataChan = make(chan *ItemData)

	mgmt.header = &Header{TypeLst: typeLst, PluginTypes: pluginTypes}
	mgmt.outputHeaderLines()
	return mgmt
}

func (t *PluginManager) dataCollection() {
	itemData := new(ItemData)
	itemData.TimeStamp = time.Now().Unix()
	for _, pluginName := range t.currentRunTypes {
		plugin := registerPlugins[pluginName]
		if itemData.Data == nil {
			itemData.Data = make(map[string]map[string]string)
		}
		itemData.Data[pluginName] = plugin.GetData()
	}
	t.itemDataChan <- itemData
}

func (t *PluginManager) Start(seconds int) {
	for _, pluginName := range t.currentRunTypes {
		t.data[pluginName].Run()
	}
	go utils.SetTimer(seconds, t.dataCollection)
	tmpLineNum := lineNum
	for {
		if tmpLineNum != 0 {
			tmpLineNum = tmpLineNum - 1
		} else {
			tmpLineNum = lineNum
			t.outputHeaderLines()
		}
		printData := <-t.itemDataChan
		timeSecFmt := time.Now().In(time.FixedZone("CST", 8*3600)).Format("15:04:05")
		cmdOutputLine := []string{timeSecFmt}
		fileOutputLine := []string{timeSecFmt}

		mItem := new(MemDataItem)
		mItem.TimeStamp = time.Now().Unix()
		mItem.ItemData = make(map[string]string)
		for _, pluginName := range t.currentRunTypes {
			mapData := printData.Data[pluginName]
			types := t.data[pluginName].GetTypes()
			for _, k := range types {
				val := mapData[k.Name]
				fileOutputLine = append(fileOutputLine, val)
				if k.IsCmdShow {
					cmdOutputLine = append(cmdOutputLine, val)
				}
				key := fmt.Sprintf("%s.%s", pluginName, k.Name)
				mItem.ItemData[key] = mapData[k.Name]
			}
		}
		t.displayLogger.Println(strings.Join(cmdOutputLine, sep))
		fpWriter.WriteString(strings.Join(fileOutputLine, csvSep) + "\n")
		fpWriter.Flush()
	}
}
