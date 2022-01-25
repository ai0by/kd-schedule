package main

import (
	"bytes"
	"fmt"
	"github.com/ai0by/kd-schedule/schedule"
	"runtime"
	"strconv"
	"time"
)

func main() {
	wk := schedule.NewWorker(1, 10)
	for i := 0; i < 28; i++ {
		_ = wk.Add(int8(i), 50, test)
	}
	wk.Start()
	time.Sleep(10 * time.Second)
}

func test() {
	b := make([]byte, 64)
	b = b[:runtime.Stack(b, false)]
	b = bytes.TrimPrefix(b, []byte("goroutine "))
	b = b[:bytes.IndexByte(b, ' ')]
	_, _ = strconv.ParseUint(string(b), 10, 64)
	fmt.Println("--------------")
}
