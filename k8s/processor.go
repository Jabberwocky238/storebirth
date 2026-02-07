package k8s

import (
	"log"
)

type Processor struct {
	JobQueue chan Job
	PoolSize int
}

type Job interface {
	Type() string
	ID() string
	Do() error
}

func NewProcessor(queueSize int, poolSize int) *Processor {
	return &Processor{
		JobQueue: make(chan Job, queueSize),
		PoolSize: poolSize,
	}
}

func (p *Processor) Submit(job Job) {
	p.JobQueue <- job
}

func (p *Processor) Start() {
	for i := 0; i < p.PoolSize; i++ {
		go func() {
			for job := range p.JobQueue {
				if err := job.Do(); err != nil {
					log.Printf("[processor] job failed (type=%s, id=%s): %v", job.Type(), job.ID(), err)
				}
			}
		}()
	}
	log.Println("[processor] started")
}
