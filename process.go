package ifrit

import (
	"os"
	"sync"
)

type Process interface {
	Wait() error
	Signal(os.Signal)
}

func Envoke(r Runner) Process {
	switch r := r.(type) {
	case RunGroup:
		return envokeGroup(r)
	default:
		return envokeProcess(r)
	}

}

func envokeProcess(r Runner) Process {
	p := &process{
		runner:         r,
		sig:            make(chan os.Signal),
		exitStatusChan: make(chan error, 1),
		ready:          make(chan struct{}),
	}
	go p.run()
	<-p.ready
	return p
}

type process struct {
	runner         Runner
	sig            chan os.Signal
	exitStatus     error
	exitStatusChan chan error
	ready          chan struct{}
	exitOnce       sync.Once
}

func (p *process) run() {
	p.exitStatusChan <- p.runner.Run(p.sig, p.ready)
}

func (p *process) Wait() error {
	p.exitOnce.Do(func() {
		p.exitStatus = <-p.exitStatusChan
	})
	return p.exitStatus
}

func (p *process) Signal(signal os.Signal) {
	go func() {
		p.sig <- signal
	}()
}
