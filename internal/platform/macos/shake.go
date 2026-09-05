package macos

import (
	"math"
	"sync"
	"time"
)

type ShakeCallback func()

type ShakeDetector struct {
	mu           sync.Mutex
	enabled      bool
	minDisp      float64
	reversalsReq int
	windowDur    time.Duration
	cooldownDur  time.Duration

	lastX        float64
	lastY        float64
	hasPos       bool
	lastDir      int // -1 for left/down, +1 for right/up, 0 for none
	reversals    []time.Time
	lastTrigger  time.Time
	onShake      ShakeCallback
	stopCh       chan struct{}
	running      bool
}

func NewShakeDetector(onShake ShakeCallback) *ShakeDetector {
	sd := &ShakeDetector{
		enabled:      true,
		minDisp:      18.0,
		reversalsReq: 4,
		windowDur:    450 * time.Millisecond,
		cooldownDur:  1200 * time.Millisecond,
		onShake:      onShake,
		stopCh:       make(chan struct{}),
	}
	return sd
}

func (sd *ShakeDetector) SetEnabled(enabled bool) {
	sd.mu.Lock()
	defer sd.mu.Unlock()
	sd.enabled = enabled
}

func (sd *ShakeDetector) SetSensitivity(level string) {
	sd.mu.Lock()
	defer sd.mu.Unlock()

	switch level {
	case "high":
		sd.minDisp = 12.0
		sd.reversalsReq = 3
		sd.windowDur = 500 * time.Millisecond
	case "low":
		sd.minDisp = 26.0
		sd.reversalsReq = 5
		sd.windowDur = 400 * time.Millisecond
	default: // normal
		sd.minDisp = 18.0
		sd.reversalsReq = 4
		sd.windowDur = 450 * time.Millisecond
	}
}

// ProcessCoord is the core pure algorithm, perfect for testability
func (sd *ShakeDetector) ProcessCoord(x, y float64, now time.Time) bool {
	sd.mu.Lock()
	defer sd.mu.Unlock()

	if !sd.enabled {
		return false
	}

	if !sd.hasPos {
		sd.lastX = x
		sd.lastY = y
		sd.hasPos = true
		return false
	}

	dx := x - sd.lastX
	sd.lastX = x
	sd.lastY = y

	// Filter out tiny cursor jitter
	if math.Abs(dx) < sd.minDisp {
		return false
	}

	currentDir := 1
	if dx < 0 {
		currentDir = -1
	}

	// If direction reversed
	if sd.lastDir != 0 && currentDir != sd.lastDir {
		// Clean up old reversals outside time window
		cutoff := now.Add(-sd.windowDur)
		valid := sd.reversals[:0]
		for _, t := range sd.reversals {
			if t.After(cutoff) {
				valid = append(valid, t)
			}
		}
		sd.reversals = append(valid, now)

		// Check if threshold reached and cooldown expired
		if len(sd.reversals) >= sd.reversalsReq && now.Sub(sd.lastTrigger) > sd.cooldownDur {
			sd.lastTrigger = now
			sd.reversals = nil
			if sd.onShake != nil {
				go sd.onShake()
			}
			return true
		}
	}

	sd.lastDir = currentDir
	return false
}

func (sd *ShakeDetector) Start() {
	sd.mu.Lock()
	if sd.running {
		sd.mu.Unlock()
		return
	}
	sd.running = true
	sd.stopCh = make(chan struct{})
	sd.mu.Unlock()

	go func() {
		ticker := time.NewTicker(20 * time.Millisecond) // 50 Hz
		defer ticker.Stop()

		for {
			select {
			case <-sd.stopCh:
				return
			case now := <-ticker.C:
				x, y := GetMousePos()
				sd.ProcessCoord(x, y, now)
			}
		}
	}()
}

func (sd *ShakeDetector) Stop() {
	sd.mu.Lock()
	defer sd.mu.Unlock()

	if !sd.running {
		return
	}
	sd.running = false
	close(sd.stopCh)
}
