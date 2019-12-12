package main

import (
	"context"
	"flag"
	"github.com/okteto/supervisor/pkg/monitor"
	log "github.com/sirupsen/logrus"
	"os"
	"os/signal"
	"syscall"
)

// CommitString is the commit used to build the server
var CommitString string

func main() {
	log.WithField("commit", CommitString).Infof("supervisor started")

	remoteFlag := flag.Bool("remote", false, "start the remote server")
	flag.Parse()

	ctx, cancel := context.WithCancel(context.Background())
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigs
		cancel()
	}()

	m := monitor.NewMonitor(ctx)
	m.Add(monitor.NewProcess(
		"syncthing",
		"/var/okteto/bin/syncthing",
		[]string{"-home", "/var/syncthing", "-gui-address", "0.0.0.0:8384", "-verbose"}),
	)

	if *remoteFlag {
		m.Add(monitor.NewProcess("remote", "/var/okteto/bin/remote", nil))
	}

	log.Info("starting monitor")
	if err := m.Run(); err != nil {
		log.WithError(err).Error("monitor finished due to an error")
	}

	log.Info("stopping monitor")
	m.Stop()
}