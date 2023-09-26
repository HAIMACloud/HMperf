// Copyright (c) 2021-2023 https://www.haimacloud.com/
// SPDX-License-Identifier: MIT

package stat

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"reflect"
	"strings"

	"github.com/firstrow/tcp_server"

	"romstat/stat/plugins"
)

var localAddress = "127.0.0.1:38421"

func NewPipelineServerListen() {
	log.Default().SetOutput(io.Discard)
	server := tcp_server.New(localAddress)
	server.OnNewMessage(func(c *tcp_server.Client, message string) {
		// new message received
		cmdOperator(c.Conn(), strings.Trim(message, "\n"))
	})
	server.Listen()
}
func cmdOperator(writer io.Writer, cmdLine string) {
	if cmdLine == "current_pkg_surface" {
		sfLatencyStatPlugin := reflect.ValueOf(registerPlugins["display"]).Interface().(*plugins.SfLatencyStatPlugin)
		pkgName, surfaceView := sfLatencyStatPlugin.GetCurrentPkgSurface()
		bz, _ := json.Marshal(map[string]string{"pkg_name": pkgName, "surface": surfaceView})
		writer.Write([]byte(fmt.Sprintf("%s\n", string(bz))))
	}
}

func AskPipelineServer(cmd string) (string, error) {
	f, err := net.Dial("tcp", localAddress)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
		return "", err
	}
	defer f.Close()
	f.Write([]byte(fmt.Sprintf("%s\n", cmd)))
	reader := bufio.NewReader(f)
	output, err := reader.ReadBytes('\n')
	if err != nil {
		return "", err
	}
	return strings.Trim(string(output), "\n"), nil
}
