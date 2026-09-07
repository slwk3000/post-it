package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"syscall"

	"github.com/slwk3000/post-it/internal/app"
)

func init() {
	// Cocoa requires main thread
	runtime.LockOSThread()
}

func ensureSingleInstance() {
	home, err := os.UserHomeDir()
	if err != nil {
		return
	}
	dir := filepath.Join(home, ".config", "post-it")
	_ = os.MkdirAll(dir, 0755)
	pidFile := filepath.Join(dir, "post-it.pid")

	if data, err := os.ReadFile(pidFile); err == nil {
		oldPid, err := strconv.Atoi(string(data))
		if err == nil && oldPid != os.Getpid() && oldPid > 0 {
			// Check if old process is still alive and terminate it so new version takes over
			if err := syscall.Kill(oldPid, syscall.SIGTERM); err == nil {
				log.Printf("Replaced previous post-it instance (PID %d)", oldPid)
			}
		}
	}

	_ = os.WriteFile(pidFile, []byte(fmt.Sprintf("%d", os.Getpid())), 0644)
}

func main() {
	ensureSingleInstance()

	application, err := app.New()
	if err != nil {
		log.Fatalf("Failed to initialize Post-it application: %v", err)
	}

	if err := application.Start(); err != nil {
		log.Fatalf("Application runtime error: %v", err)
	}
}
