package main

import (
	"log"
	"runtime"

	"github.com/slwk3000/post-it/internal/app"
)

func init() {
	// Cocoa requires main thread
	runtime.LockOSThread()
}

func main() {
	application, err := app.New()
	if err != nil {
		log.Fatalf("Failed to initialize Post-it application: %v", err)
	}

	if err := application.Start(); err != nil {
		log.Fatalf("Application runtime error: %v", err)
	}
}
