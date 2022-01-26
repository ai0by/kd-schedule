package main

import (
	"fmt"
	"github.com/ai0by/kd-schedule/schedule"
	"time"
)

func main() {
	wk := schedule.NewWorker(1, 100)
	for i := 1; i < 999; i++ {
		_ = wk.Add(uint8(i), 50, test)
	}
	wk.Start()
	wk.Add(200,50,test1)
	time.Sleep(10 * time.Second)
}

func test() {
	fmt.Println("--------------")
	time.Sleep(500 * time.Millisecond)
}

func test1(){
	fmt.Println("+++++++++++++++++")
	time.Sleep(50 * time.Millisecond)
}