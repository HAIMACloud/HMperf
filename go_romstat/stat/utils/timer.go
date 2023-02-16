// Copyright (c) 2021-2023 https://www.haimacloud.com/
// SPDX-License-Identifier: MIT

package utils

import "time"

type Callback func()

func SetTimer(sec int, callfunc Callback) {
	t := time.NewTicker(time.Duration(sec) * time.Second)
	for range t.C {
		callfunc()
	}
}

func SetTimerMilliSecond(milliSec int, callfunc Callback) {
	t := time.NewTicker(time.Duration(milliSec) * time.Millisecond)
	for range t.C {
		callfunc()
	}
}
