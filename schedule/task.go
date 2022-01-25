package schedule

import (
	"fmt"
	"time"
)

/**
最大支持 0 ~ 65535 个协程控制
*/

type Worker struct {
	Kind    uint8          `json:"kind"`
	Cap     uint16         `json:"cap"`
	C       chan *Job      `json:"channel"`
	Entry   []*WorkerEntry `json:"Entry"`
	Code    uint16         `json:"code"`
	Running bool           `json:"running"`
}

type WorkerApi interface {
}

// 任务单条队列
type WorkerEntry struct {
	code     uint16
	Head     *Job
	ExecTime time.Duration
}

// 任务结构体
type Job struct {
	Func     func()
	Weight   int8
	ExecTime time.Duration
	pre      *Job
	next     *Job
}

func NewWorker(kind uint8, length uint16) *Worker {
	worker := new(Worker)
	worker.Kind = kind
	worker.Cap = length
	worker.Code = 0x00
	worker.Running = false
	worker.C = make(chan *Job)

	go worker.monitorAdd()

	fmt.Println("New Worker Success ...")
	return worker
}

// monitor add task
func (w *Worker) monitorAdd() {
	var i uint16
	for i = 0; i < w.Cap; i++ {
		w.Entry = append(w.Entry,&WorkerEntry{code:i,Head:&Job{}})
	}

	for {
		job := <-w.C
		if job.Weight <= 0 {
			fmt.Println("monitor channel exit ...")
			return
		}
		if w.Code+1 < w.Cap {
			w.Code = w.Code + 1
			w.Entry[w.Code].addJob(job)
		} else {
			w.Code = 0x00
			w.Entry[w.Code].addJob(job)
		}
	}
}

// 增加任务
func (w *Worker) Add(weight int8, time int, cmd func()) error {
	if weight <= 0 {
		return fmt.Errorf("Add job error,weight can not is 0 ... ")
	}
	job := &Job{
		Func:   cmd,
		Weight: weight,
	}
	if w.Running {
		if w.Code+1 >= w.Cap {
			w.Code = 0x00
		}
		w.Entry[w.Code+1].addJob(job)
	} else {
		w.C <- job
	}

	return nil
}

// 动态加载
func (s *WorkerEntry) addJob(job *Job) {
	err := insertNode(s.Head, job)
	if err !=nil {
		fmt.Println(err)
	}
}

// 开始执行
func (w *Worker) Start() error {
	if w.Running {
		return fmt.Errorf("Worker is already running ... ")
	} else {
		w.Running = true
		// close monitor
		w.C <- &Job{}
	}
	go func() {
		for  {
			for _, v := range w.Entry {
				go v.list()
			}
		}
	}()
	return nil
}

// 队列执行
func (s *WorkerEntry) list() {
	listBegin := time.Now()
	show := s.Head
	// is empty
	if show.next == nil {
		fmt.Println("Links List is empty ...")
		return
	}
	for {
		if show.next == nil {
			s.ExecTime = time.Now().Sub(listBegin)
			return
		}
		show = show.next
		before := time.Now()
		show.Func()
		show.ExecTime = time.Now().Sub(before)
	}
}

func (w *Worker) Stop() {

}

// 销毁
func (w *Worker) Destroy() {

}

// ================== Linked list ==================

// add node end
func addNodeEnd(head *Job, jobs *Job) error {
	temp := head
	for {
		if temp.next == nil {
			break
		}
		temp = temp.next
	}
	temp.next = jobs
	jobs.pre = temp
	return nil
}

// insert node
func insertNode(head *Job, job *Job) error {
	temp := head
	last := false
	for {
		if temp.next == nil {
			last = true
			break
		} else if temp.next.Weight < job.Weight {
			break
		}
		temp = temp.next
	}
	if last {
		temp.next = job
		job.pre = temp
	} else {
		job.next = temp.next
		temp.next = job
		job.next.pre = job
		job.pre = temp
	}
	return nil
}

// show all links node
func getJobNode(head *Job) []*Job {
	var jobs []*Job
	show := head
	// is empty
	if show.next == nil {
		fmt.Println("Links is empty ...")
		return jobs
	}
	for {
		if show.next == nil {
			return jobs
		}
		jobs = append(jobs, show)
		show = show.next
	}
}

// delete node and show
func delLinksNode(head *Job, job *Job) *Job {
	temp := head
	if temp.next == nil {
		return job
	}
	// is last
	last := false
	for {
		if temp.next.Weight == job.Weight {
			// is last
			if temp.next.next == nil {
				last = true
			}
			break
		} else if temp.next == nil {
			fmt.Println("Del node error , node not exist ...")
			return job
		}
		temp = temp.next
	}

	if last {
		temp.next = nil
	} else {
		temp.next = temp.next.next
		temp.next.pre = temp
	}
	return job
}
