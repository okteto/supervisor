package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"syscall"

	"github.com/okteto/supervisor/pkg/monitor"
	reaper "github.com/ramr/go-reaper"
	log "github.com/sirupsen/logrus"
)

// CommitString is the commit used to build the server
var CommitString string

func main() {
	log.WithField("commit", CommitString).Infof("supervisor started")

	//  Start background reaping of orphaned child processes.
	reaper.Reap()

	remoteFlag := flag.Bool("remote", false, "start the remote server")
	resetFlag := flag.Bool("reset", false, "reset syncthing database")
	flag.Parse()

	ctx, cancel := context.WithCancel(context.Background())
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigs
		cancel()
	}()

	m := monitor.NewMonitor(ctx)
	reset := "-reset-deltas"
	if *resetFlag {
		reset = "-reset-database"
	}
	m.Add(monitor.NewProcess(
		"syncthing",
		"/var/okteto/bin/syncthing",
		[]string{"-home", "/var/syncthing", "-gui-address", "0.0.0.0:8384", "-verbose", reset}),
	)

	if *remoteFlag {
		m.Add(monitor.NewProcess("remote", "/var/okteto/bin/okteto-remote", nil))
	}

	log.Info("starting monitor")
	if err := m.Run(); err != nil {
		log.WithError(err).Error("monitor finished due to an error")
	}

	log.Info("stopping monitor")
	m.Stop()
}
