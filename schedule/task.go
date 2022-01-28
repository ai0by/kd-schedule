package schedule

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

/**
最大支持 0 ~ 65535 个协程控制
*/

type Worker struct {
	Kind    uint8              `json:"kind"`
	Cap     uint16             `json:"cap"`
	C       chan *Job          `json:"channel"`
	Entry   []*WorkerEntry     `json:"Entry"`
	Code    uint16             `json:"code"`
	Running bool               `json:"running"`
	Monitor bool               `json:"monitor"`
	ExecNum int32              `json:"exec_num"`
	ExecSig chan uint          `json:"exec_sig"`
	Ctx     context.Context    `json:"ctx"`
	Cancel  context.CancelFunc `json:"cancel"`
}
type WorkerApi struct {
	f WorkerFuncS
	args []interface{}
}

type WorkerFuncS func(args ...interface{})

type WorkerFunc interface {
	TaskFunc(args ...interface{})
}

// 任务单条队列
type WorkerEntry struct {
	code       uint16
	Head       *Job
	ExecTime   time.Duration
	Running    bool
	WaitJobNum int32
	Lock       sync.Mutex
}

// 任务结构体
type Job struct {
	Func     WorkerFunc
	Weight   uint8
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
	worker.ExecSig = make(chan uint)
	worker.Ctx, worker.Cancel = context.WithCancel(context.Background())

	go worker.monitorAdd()

	fmt.Println("New Worker Success ...")
	return worker
}

// monitor add task
func (w *Worker) monitorAdd() {
	var i uint16
	for i = 0; i < w.Cap; i++ {
		w.Entry = append(w.Entry, &WorkerEntry{code: i, Head: &Job{}, WaitJobNum: 0})
	}
	w.Monitor = true
	for {
		job := <-w.C
		if job.Weight <= 0 {
			fmt.Println("monitor channel exit ...")
			w.Monitor = false
			return
		}
		if w.Code < w.Cap {
			w.Entry[w.Code].addJob(job)
		} else {
			w.Code = 0x00
			w.Entry[w.Code].addJob(job)
		}
		w.Code = w.Code + 1
	}
}

// 增加任务
func (w *Worker) Add(weight uint8, time int, cmd WorkerFunc) error {
	if weight <= 0 {
		return fmt.Errorf("Add job error,weight can not is 0 ... ")
	}
	job := &Job{
		Func:   cmd,
		Weight: weight,
	}
	if w.Running {
		err := w.realTimeAdd(job)
		if err != nil {
			return err
		}
	} else {
		w.C <- job
	}

	return nil
}

// 增加闭包任务
func (w *Worker) AddClosureFunc(weight uint8, time int, cmd func(args ...interface{}), args ...interface{}) error {
	var wa WorkerApi
	wa.f = cmd
	wa.args = args
	return w.Add(weight, time, &wa)
}

// 按队列增加任务
func (s *WorkerEntry) addJob(job *Job) {
	err := s.insertNode(job)
	if err != nil {
		fmt.Println(err)
	}
	atomic.AddInt32(&s.WaitJobNum, 1)
}

// 动态加载
func (w *Worker) realTimeAdd(job *Job) error {
	// find free queue or find small amount of task
	var fewCode uint16
	var fewNum int32
	for k := range w.Entry {
		if !w.Entry[k].Running {
			err := w.Entry[k].insertNode(job)
			return err
		}
		if w.Entry[k].WaitJobNum < fewNum {
			fewCode = w.Entry[k].code
			fewNum = w.Entry[k].WaitJobNum
		}
	}
	err := w.Entry[fewCode].insertNode(job)
	return err
}

// 开始执行
func (w *Worker) Start() error {
	if w.Running {
		return fmt.Errorf("Worker is already running ... ")
	} else {
		w.Running = !w.Running
		// close monitor
		w.C <- &Job{}
	}
	go func() {
		for {
			select {
			case <-w.Ctx.Done():
				return
			case <-w.ExecSig:
				for {
					if uint16(w.ExecNum) < w.Cap && w.ExecNum >= 0 {
						w.C <- &Job{}
						break
					}
				}
			}
		}
	}()
	go func() {
		for {
			if w.Monitor {
				continue
			}
			for _, v := range w.Entry {
				//v.showCode()
				w.ExecSig <- uint(0x01)
				select {
				case <-w.C:
					go func(s *WorkerEntry) {
						if s.WaitJobNum <= 0 || s.Head.next == nil {
							s.Running = false
							s.WaitJobNum = 0
							return
						} else if s.Running {
							return
						} else {
							atomic.AddInt32(&w.ExecNum, 1)
							s.list(w.Ctx)
							atomic.AddInt32(&w.ExecNum, -1)
						}
					}(v)
				case <-w.Ctx.Done():
					return
				}
			}
		}
	}()
	return nil
}

func (w *Worker) ShowJob() {
	for key := range w.Entry {
		fmt.Println(w.Entry[key].getJobNode())
	}
}

// 队列执行
func (s *WorkerEntry) list(ctx context.Context) {
	s.Running = true
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
		select {
		case <-ctx.Done():
			return
		default:
			show = show.next
			before := time.Now()
			//s.showCode()
			show.Func.TaskFunc()
			show.ExecTime = time.Now().Sub(before)
			s.delLinksNode(show)
			if s.WaitJobNum != 0 {
				atomic.AddInt32(&s.WaitJobNum, -1)
			}
			fmt.Println(s.code, ".........", s.WaitJobNum)
		}
	}
}

// show routine code
func (s *WorkerEntry) showCode() {
	fmt.Println("Execute routine code :", s.code, " Job number is ", s.WaitJobNum)
}

func (w *Worker) SetCap(size uint16) error {
	if size <= w.Cap {
		return fmt.Errorf("set size error,excessive capacity ... ")
	}
	w.Cap = size
	return nil
}

func (w *Worker) Stop() {
	w.Monitor = true
	w.Running = false
	w.ExecNum = 0
	w.Cancel()
}

// 销毁
func (w *Worker) Destroy() {
	w.Stop()
	for k, v := range w.Entry {
		for {
			temp := v.Head
			if temp == nil {
				v.Head = nil
				break
			}
			delLinkNode := temp
			temp = temp.next
			delLinkNode.next = nil
			delLinkNode.pre = nil
		}
		w.Entry[k] = nil
	}
	w.Entry = nil
	return
}

// ================== Linked list ==================

// add node end
func (s *WorkerEntry) addNodeEnd(jobs *Job) error {
	temp := s.Head
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
func (s *WorkerEntry) insertNode(job *Job) error {
	temp := s.Head
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
func (s *WorkerEntry) getJobNode() []*Job {
	var jobs []*Job
	show := s.Head
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
func (s *WorkerEntry) delLinksNode(job *Job) *Job {
	s.Lock.Lock()
	defer s.Lock.Unlock()
	fmt.Println("del node ", job.Weight)
	temp := s.Head
	if temp.next == nil {
		return job
	}
	// is last
	last := false
	for {
		//s.showCode()
		if temp.next == nil {
			fmt.Println("Del node error , node not exist ... error pointer is ", s.code, " error pointer num is ", s.WaitJobNum, " weight is ", job.Weight)
			return job
		} else if temp.next == job {
			// is last
			if temp.next.next == nil {
				last = true
			}
			break
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

func (e *WorkerApi) TaskFunc(args ...interface{}) {
	e.f(e.args...)
}
