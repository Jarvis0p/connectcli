package punchmonitor

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strconv"
	"syscall"
	"time"

	"connectcli/internal/api"
	"connectcli/internal/config"
	"connectcli/internal/credentials"
	"connectcli/internal/notifications"
	"connectcli/internal/utils"
)

// LogPath returns the path to the monitor log file.
func LogPath() (string, error) {
	d, err := Dir()
	if err != nil {
		return "", err
	}
	return filepath.Join(d, "punch_monitor.log"), nil
}

// Spawn starts a detached background monitor (re-execs this binary with __punch-monitor).
func Spawn() error {
	if err := Stop(); err != nil {
		return err
	}

	exe, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to resolve executable: %w", err)
	}

	logPath, err := LogPath()
	if err != nil {
		return err
	}
	logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		return fmt.Errorf("failed to open monitor log: %w", err)
	}

	cmd := exec.Command(exe, "__punch-monitor")
	cmd.Stdout = logFile
	cmd.Stderr = logFile
	cmd.Stdin = nil

	if err := cmd.Start(); err != nil {
		_ = logFile.Close()
		return fmt.Errorf("failed to start punch monitor: %w", err)
	}

	if err := WritePID(cmd.Process.Pid); err != nil {
		_ = terminateProcess(cmd.Process)
		_ = logFile.Close()
		return fmt.Errorf("failed to write monitor pid file: %w", err)
	}

	go func() {
		_ = cmd.Wait()
		_ = logFile.Close()
	}()

	return nil
}

// RunMonitor is the main loop for the background process (every 10 min Slack + session check).
func RunMonitor() error {
	logPath, err := LogPath()
	if err != nil {
		return err
	}
	f, err := os.OpenFile(logPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	defer f.Close()
	logger := log.New(f, "", log.LstdFlags)
	logger.Printf("punch monitor started pid=%d", os.Getpid())

	creds, err := credentials.LoadCredentials()
	if err != nil {
		logger.Printf("credentials error: %v", err)
		return err
	}
	if creds.SlackWebhook == "" {
		err := fmt.Errorf("slack_webhook not set")
		logger.Printf("%v", err)
		return err
	}

	if err := utils.EnsureObjectID(); err != nil {
		logger.Printf("object id: %v", err)
		return err
	}

	cfg, err := config.LoadConfig()
	if err != nil {
		logger.Printf("config: %v", err)
		return err
	}

	objectID, err := strconv.Atoi(cfg.PunchClockObjectID)
	if err != nil {
		logger.Printf("object id int: %v", err)
		return err
	}

	statusClient := api.NewClockStatusClient()
	slack := notifications.NewSlackClient(creds.SlackWebhook)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	ticker := time.NewTicker(10 * time.Minute)
	defer ticker.Stop()

	me := os.Getpid()
	defer func() {
		_ = RemovePIDFileIfMatches(me)
		logger.Printf("punch monitor exiting pid=%d", me)
	}()

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			resp, err := statusClient.GetStatus(creds, objectID)
			if err != nil {
				logger.Printf("clock status error: %v", err)
				continue
			}
			if !resp.IsClockedIn() {
				logger.Printf("no open punch (clocked out elsewhere); stopping monitor")
				return nil
			}

			punchInTime := time.Unix(resp.PunchInTimestamp(), 0)
			elapsed := time.Since(punchInTime)
			clientName := resp.ClientName()
			if clientName == "" {
				clientName = "unknown client"
			}

			msg := fmt.Sprintf("clocked-in %s for %s", clientName, notifications.FormatDuration(elapsed))
			if err := slack.Send(msg); err != nil {
				logger.Printf("slack error: %v", err)
			} else {
				logger.Printf("slack: %s", msg)
			}
		}
	}
}
