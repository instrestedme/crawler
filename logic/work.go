package logic

import (
	"github.com/spf13/viper"
	"sync"
	"time"
)

// Worker 必须满足接口类型，
// 才能使用工作池
// 用来管理一个 goroutine 池来完成工作
type Worker interface {
	Work()
}

// Pool 提供一个 goroutine 池，这个池可以完成
// 任何已提交的 Worker 任务
type Pool struct {
	worker chan Worker
	wg     sync.WaitGroup
}

// NewPool 创建一个新工作池
func NewPool(maxGoroutines int) *Pool {
	p := Pool{
		worker: make(chan Worker),
	}

	p.wg.Add(maxGoroutines)
	for i := 0; i < maxGoroutines; i++ {
		go func() {
			for w := range p.worker {
				w.Work()
				// 睡眠一段时间，防止封ip
				time.Sleep(time.Duration(viper.GetInt("crawl.sleep")) * time.Second)
			}
			p.wg.Done()
		}()
	}

	return &p
}

// Run 提交工作到工作池
func (p *Pool) Run(w Worker) {
	p.worker <- w
}

// Shutdown 等待所有 goroutine 停止工作
func (p *Pool) Shutdown() {
	close(p.worker)
	p.wg.Wait()
}
