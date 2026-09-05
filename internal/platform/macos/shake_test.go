package macos

import (
	"sync/atomic"
	"testing"
	"time"
)

func TestShakeDetectorStraightLine(t *testing.T) {
	var triggered int32
	sd := NewShakeDetector(func() {
		atomic.AddInt32(&triggered, 1)
	})

	now := time.Now()
	// Move in one continuous direction (no reversals)
	for i := 0; i < 20; i++ {
		now = now.Add(20 * time.Millisecond)
		sd.ProcessCoord(float64(i*50), 200, now)
	}

	if atomic.LoadInt32(&triggered) != 0 {
		t.Errorf("straight line movement should not trigger shake")
	}
}

func TestShakeDetectorRapidShake(t *testing.T) {
	var triggered int32
	sd := NewShakeDetector(func() {
		atomic.AddInt32(&triggered, 1)
	})

	now := time.Now()
	// Rapid back and forth movement: 0 -> 100 -> 0 -> 100 -> 0 -> 100
	positions := []float64{0, 100, 0, 100, 0, 100, 0, 100}
	for _, pos := range positions {
		now = now.Add(30 * time.Millisecond)
		sd.ProcessCoord(pos, 200, now)
	}

	// Give a tiny moment for callback goroutine
	time.Sleep(20 * time.Millisecond)

	if atomic.LoadInt32(&triggered) == 0 {
		t.Errorf("rapid back-and-forth movement should trigger shake")
	}
}

func TestShakeDetectorCooldown(t *testing.T) {
	var triggered int32
	sd := NewShakeDetector(func() {
		atomic.AddInt32(&triggered, 1)
	})

	now := time.Now()
	// First shake
	for _, pos := range []float64{0, 100, 0, 100, 0, 100, 0, 100} {
		now = now.Add(30 * time.Millisecond)
		sd.ProcessCoord(pos, 200, now)
	}

	time.Sleep(10 * time.Millisecond)
	if atomic.LoadInt32(&triggered) != 1 {
		t.Fatalf("expected 1 trigger, got %d", atomic.LoadInt32(&triggered))
	}

	// Immediate second shake within 200ms (cooldown is 1200ms)
	for _, pos := range []float64{0, 100, 0, 100, 0, 100, 0, 100} {
		now = now.Add(30 * time.Millisecond)
		sd.ProcessCoord(pos, 200, now)
	}

	time.Sleep(10 * time.Millisecond)
	if atomic.LoadInt32(&triggered) != 1 {
		t.Errorf("cooldown should prevent immediate re-trigger, got %d", atomic.LoadInt32(&triggered))
	}

	// Wait past cooldown
	now = now.Add(1500 * time.Millisecond)
	for _, pos := range []float64{0, 100, 0, 100, 0, 100, 0, 100} {
		now = now.Add(30 * time.Millisecond)
		sd.ProcessCoord(pos, 200, now)
	}

	time.Sleep(10 * time.Millisecond)
	if atomic.LoadInt32(&triggered) != 2 {
		t.Errorf("expected second trigger after cooldown, got %d", atomic.LoadInt32(&triggered))
	}
}
