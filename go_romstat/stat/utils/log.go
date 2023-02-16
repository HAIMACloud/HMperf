// Copyright (c) 2021-2023 https://www.haimacloud.com/
// SPDX-License-Identifier: MIT

package utils

import (
	"log"
	"os"

	"romstat/stat/data"
)

type Logger interface {
	Println(v ...interface{})
	Printf(format string, v ...interface{})
}

var (
	DisplayLogger Logger
	DebugLogger   Logger
)

func InitLogger() {
	DisplayLogger = log.New(os.Stdout, "", 0)
	DebugLogger = NewDebugLogger()
}

type DebugLoggerInstance struct {
	debugLogger *log.Logger
}

func NewDebugLogger() *DebugLoggerInstance {
	if data.GetCmdParameters().IsDebug {
		file, err := os.OpenFile("/data/local/tmp/romstat_d.log", os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0666)
		if err != nil {
			log.Fatal(err)
		}
		debugLogger := log.New(file, "DEBUG: ", log.Ldate|log.Ltime|log.Lshortfile)
		return &DebugLoggerInstance{debugLogger: debugLogger}
	} else {
		return &DebugLoggerInstance{debugLogger: nil}
	}
}

func (t *DebugLoggerInstance) Println(v ...interface{}) {
	if t.debugLogger != nil {
		t.debugLogger.Println(v...)
	}
}
func (t *DebugLoggerInstance) Printf(format string, v ...interface{}) {
	if t.debugLogger != nil {
		t.debugLogger.Printf(format, v...)
	}
}
