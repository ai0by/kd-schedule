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
	var edrS = &edr{Num:0}
	var edr2 = &edr2{Num:0}
	wk := schedule.NewWorker(1, 100)
	for i := 1; i < 9999; i++ {
		_ = wk.Add(uint8(i), 50, edrS)
	}
	wk.Start()
	// 实现接口增加任务
	wk.Add(200,50, edr2)
	// 闭包带参数任务
	wk.AddClosureFunc(199,50, func(args ...interface{}) {
		for _,v := range args{
			fmt.Println(v.(string))
		}
	},"~~~~~~~~~~~~~~~~~~~~~~~~~~~~")
	wk.AddClosureFunc(199,50, func(args ...interface{}) {
		for _,v := range args{
			fmt.Println(v.(string))
		}
	},"---------------------------")
	time.Sleep(2 * time.Second)
	//wk.Stop()
	time.Sleep(10 * time.Second)
}

func (e *edr)TaskFunc(args ...interface{}) {
	fmt.Println("--------------",e.Num)
	e.Num++
	time.Sleep(500 * time.Millisecond)
}

func (e *edr2)TaskFunc(args ...interface{}){
	fmt.Println("+++++++++++++++++",e.Num)
	e.Num++
	time.Sleep(50 * time.Millisecond)
}