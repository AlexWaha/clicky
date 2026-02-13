package main

import (
	"math"
	"math/rand"
	"sync/atomic"
	"time"
)

// ── Platform contract ───────────────────────────────────────────────────────
// Each platform_*.go must implement these functions:
//
//   platformRun()                              – create GUI + run event loop (blocks)
//   platformSetCursorPos(x, y int)             – move cursor (screen coords)
//   platformGetCursorPos() (int, int)           – get cursor position
//   platformClick()                             – simulate left click
//   platformPreventSleep()                      – prevent system sleep
//   platformAllowSleep()                        – allow system sleep
//   platformMoveButton(x, y int)                – move button (client coords)
//   platformClientToScreen(x, y int) (int, int) – convert client → screen coords
//   platformSetButtonActive(isActive bool)       – change button appearance
//   platformReinforceTopmost()                  – reinforce always-on-top
//   platformQuit()                              – quit application

// ── Callbacks (set here, called by platform) ────────────────────────────────

var onButtonClicked func()
var onHotkeyQuit func()

// ── Shared state ────────────────────────────────────────────────────────────

var active atomic.Bool

// ── Layout constants ────────────────────────────────────────────────────────

const (
	clientW = 300
	clientH = 300
	btnW    = 80
	btnH    = 30
	pad     = 10
)

// Button corner positions inside the client area
var btnCorners = [4][2]int{
	{pad, pad},                               // top-left
	{clientW - btnW - pad, pad},              // top-right
	{clientW - btnW - pad, clientH - btnH - pad}, // bottom-right
	{pad, clientH - btnH - pad},              // bottom-left
}

// ── Init ────────────────────────────────────────────────────────────────────

func initApp() {
	onButtonClicked = handleButtonClick
	onHotkeyQuit = handleQuit
}

func handleButtonClick() {
	if !active.Load() {
		active.Store(true)
		platformSetButtonActive(true)
		go aliveLoop()
	}
}

func handleQuit() {
	active.Store(false)
	platformQuit()
}

// ── Random delay ────────────────────────────────────────────────────────────

// sleepWithCancel sleeps for the given duration, checking active every 100ms.
func sleepWithCancel(d time.Duration) {
	end := time.Now().Add(d)
	for time.Now().Before(end) && active.Load() {
		remaining := time.Until(end)
		if remaining > 100*time.Millisecond {
			time.Sleep(100 * time.Millisecond)
		} else {
			time.Sleep(remaining)
		}
	}
}

// randomDelay waits 1-5 seconds with cancellation support.
func randomDelay() {
	delay := time.Duration(1000+rand.Intn(4001)) * time.Millisecond
	sleepWithCancel(delay)
}

// ── Bézier curve movement ───────────────────────────────────────────────────

func moveCursorAlongCurve(toX, toY int) {
	fromX, fromY := platformGetCursorPos()

	dx := float64(toX - fromX)
	dy := float64(toY - fromY)
	dist := math.Sqrt(dx*dx + dy*dy)

	if dist < 2 {
		platformSetCursorPos(toX, toY)
		return
	}

	// Perpendicular unit vector
	perpX := -dy / dist
	perpY := dx / dist

	// Random offset amplitude: [-dist/3, +dist/3]
	amplitude := (rand.Float64()*2 - 1) * dist / 3

	// Control point = midpoint + perpendicular offset
	midX := float64(fromX) + dx/2
	midY := float64(fromY) + dy/2
	cpX := midX + perpX*amplitude
	cpY := midY + perpY*amplitude

	const steps = 25
	const stepDelay = 4 * time.Millisecond

	for i := 1; i <= steps; i++ {
		t := float64(i) / float64(steps)
		inv := 1 - t

		// Quadratic Bézier: B(t) = (1-t)²·P0 + 2·(1-t)·t·P1 + t²·P2
		x := inv*inv*float64(fromX) + 2*inv*t*cpX + t*t*float64(toX)
		y := inv*inv*float64(fromY) + 2*inv*t*cpY + t*t*float64(toY)

		platformSetCursorPos(int(math.Round(x)), int(math.Round(y)))
		time.Sleep(stepDelay)
	}
}

// ── Alive loop ──────────────────────────────────────────────────────────────

func aliveLoop() {
	platformPreventSleep()
	defer platformAllowSleep()

	idx := 0
	for active.Load() {
		c := btnCorners[idx%4]

		// Move button to next corner
		platformMoveButton(c[0], c[1])

		// Reinforce topmost
		platformReinforceTopmost()

		sleepWithCancel(400 * time.Millisecond)
		if !active.Load() {
			break
		}

		// Get button center in screen coordinates
		sx, sy := platformClientToScreen(c[0]+btnW/2, c[1]+btnH/2)

		// Move cursor along Bézier curve to button center
		moveCursorAlongCurve(sx, sy)

		sleepWithCancel(200 * time.Millisecond)
		if !active.Load() {
			break
		}

		// Click
		platformClick()

		// Random delay 1-5 seconds
		randomDelay()
		idx++
	}
}
