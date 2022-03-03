package main

import (
	"fmt"
	"github.com/ai0by/kd-schedule/schedule"
	"time"
)

type edr struct {
	Num int
}
type edr2 struct {
	Num int
}

func main() {
	var edrS = &edr{Num: 0}
	//var edr2 = &edr2{Num: 0}
	wk := schedule.NewWorker(1, 100)
	//for i := 0; i < 100; i++ {
	//	_ = wk.Add(uint8(i), 50, edrS)
	//}
	//wk.ShowJob()

	//time.Sleep(10*time.Second)

	wk.Start()
	//time.Sleep(5*time.Second)

	for i := 0; i < 999; i++ {
		_ = wk.Add(uint8(i), 50, edrS)
	}
	//wk.Add(200, 50, edr2)

	//wk.ShowJob()
	wk.ShowStatus()

	wk.Add(1, 50, edrS)

	//time.Sleep(2*time.Second)

	//wk.ShowJob()
	//wk.ShowStatus()

	go func() {
		var num = 1
		for {
			time.Sleep(5*time.Second)
			wk.AddClosureFunc(1, 50, func(args ...interface{}) {
				fmt.Println("Go ",num)
			},num)
			num++
			fmt.Println("-------------- Add One --------------")
		}
	}()

	for  {
		fmt.Println("=======================")
		wk.ShowStatus()
		time.Sleep(5*time.Second)
	}

	time.Sleep(999 * time.Second)
}

func (e *edr) TaskFunc(args ...interface{}) {
	fmt.Println("Got it ",e.Num)
	e.Num++
	time.Sleep(500 * time.Millisecond)
}

func (e *edr2) TaskFunc(args ...interface{}) {
	fmt.Println("Check it ",e.Num)
	e.Num++
	time.Sleep(50 * time.Millisecond)
}
