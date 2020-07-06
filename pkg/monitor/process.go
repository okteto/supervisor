package monitor

import (
	"os"
	"strings"
	"time"

	"github.com/go-cmd/cmd"
	log "github.com/sirupsen/logrus"
)

type state string

const (
	maxRetries         = 10
	neverStarted state = "never"
	started      state = "started"
	stopping     state = "stopping"
	stopped      state = "stopped"
	fatal        state = "fatal"
)

// Process is process monitored
type Process struct {
	Name       string
	Path       string
	Args       []string
	logger     *log.Entry
	cmd        *cmd.Cmd
	started    time.Time
	startCount int
	state      state
}

// NewProcess returns an intialized process
func NewProcess(name string, path string, args []string) *Process {
	p := &Process{
		Name:  name,
		Path:  path,
		Args:  args,
		state: neverStarted,
	}

	p.logger = log.WithField("process", p.Name)

	if p.Args == nil {
		p.Args = []string{}
	}

	return p
}

func (p *Process) start() {
	p.startCount++
	p.cmd = cmd.NewCmdOptions(cmd.Options{
		Streaming: true,
	}, p.Path, p.Args...)

	p.cmd.Env = os.Environ()

	go func() {
		for line := range p.cmd.Stdout {
			p.logger.Info(line)
		}
	}()

	go func() {
		for line := range p.cmd.Stderr {
			p.logger.Info(line)
		}
	}()

	p.logger.Infof("starting %s %s", p.Path, strings.Join(p.Args, " "))
	p.cmd.Start()

	p.started = time.Now()
	status := p.cmd.Status()
	if status.Error == nil {
		d := time.Duration(5) * time.Second
		<-time.After(d)
		if p.isRunning() {
			if status.PID > 0 {
				p.logger = p.logger.WithField("pid", status.PID)
			}

			p.logger.Info("process started")
			p.state = started
			go p.monitor()
			return
		}

		p.state = fatal
		p.logger.Errorf("process wasn't running after %s", d.String())
		return

	}

	p.logger.Errorf("process didn't start: %s", status.Error)
	p.state = fatal
}

func (p *Process) isRunning() bool {
	if p.cmd == nil {
		return false
	}

	if status := p.cmd.Status(); !status.Complete {
		return true
	}

	return false
}

func (p *Process) shouldStart() bool {
	return p.startCount < maxRetries
}

func (p *Process) monitor() {
	<-p.cmd.Done()
	if p.state == stopping {
		return
	}

	status := p.cmd.Status()
	if status.Error == nil {
		p.logger.Infof("process exited with status %d", status.Exit)
		p.state = stopped
	} else {
		p.logger.WithError(status.Error).Errorf("process exited with error status %d", status.Exit)
		p.state = fatal
	}
}

func (p *Process) stop() {
	if p.cmd == nil {
		p.logger.Info("process hasn't started")
		return
	}

	p.state = stopping

	p.logger.Info("stopping process")
	err := p.cmd.Stop()
	if err != nil {
		p.logger.WithError(err).Error("failed to stop process")
		return
	}

	<-p.cmd.Done()
	p.logger.Info("process stopped")
	p.state = stopped
	p.cmd = nil
}
