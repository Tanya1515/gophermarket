package intaccrual

type Semaphore struct {
	semaphoreChan chan struct{}
}

func NewSemaphore(limit int) *Semaphore {
	return &Semaphore{
		semaphoreChan: make(chan struct{}, limit),
	}
}

func (s *Semaphore) Lock() {
	s.semaphoreChan <- struct{}{}
}

func (s *Semaphore) Acquire() {
	<-s.semaphoreChan
}
