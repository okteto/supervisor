package main

import (
	"context"
	"flag"
	"os"
	"os/exec"
	"os/signal"
	"syscall"

	"github.com/okteto/supervisor/pkg/monitor"
	"github.com/okteto/supervisor/pkg/setup"
	reaper "github.com/ramr/go-reaper"
	log "github.com/sirupsen/logrus"
)

// CommitString is the commit used to build the server
var CommitString string

const (
	defaultSyncthingData   = "/var/syncthing/data"
	defaultSyncthingConfig = "/var/syncthing"
	defaultSyncthingSecret = "/var/syncthing/secret"
)

func main() {
	log.WithField("commit", CommitString).Infof("supervisor started")

	//  Start background reaping of orphaned child processes.
	reaper.Reap()

	remoteFlag := flag.Bool("remote", false, "start the remote server")
	resetFlag := flag.Bool("reset", false, "reset syncthing database")
	verboseFlag := flag.Bool("verbose", true, "syncthing verbosity")

	flag.Parse()

	ctx, cancel := context.WithCancel(context.Background())
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigs
		cancel()
	}()

	if *resetFlag {
		if err := setup.Setup(defaultSyncthingSecret, defaultSyncthingConfig); err != nil {
			log.WithError(err).Error("error setting up syncthing")
			os.Exit(1)
		}
		cmd := exec.Command(monitor.SyncthingBin, "--config", defaultSyncthingConfig, "--data", defaultSyncthingData, "--reset-database")
		output, err := cmd.CombinedOutput()
		if err != nil {
			log.WithError(err).Errorf("error resetting syncthing database: %s", output)
		}
	}

	m := monitor.NewMonitor(ctx, monitor.NewSyncthingConfig(defaultSyncthingConfig, defaultSyncthingSecret, defaultSyncthingData))

	syncthingArgs := []string{"--config", defaultSyncthingConfig, "--data", defaultSyncthingData, "--gui-address", "0.0.0.0:8384"}
	if *verboseFlag {
		syncthingArgs = append(syncthingArgs, "--verbose")
	}
	m.Add(monitor.NewProcess("syncthing", monitor.SyncthingBin, syncthingArgs))

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
