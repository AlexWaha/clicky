//go:build darwin

package main

/*
#cgo LDFLAGS: -framework Cocoa -framework CoreGraphics -framework IOKit
#include "objc_darwin.h"
*/
import "C"

import "unsafe"

// ── Go exports for Objective-C callbacks ────────────────────────────────────

//export goOnButtonClicked
func goOnButtonClicked() {
	if onButtonClicked != nil {
		onButtonClicked()
	}
}

//export goOnHotkeyQuit
func goOnHotkeyQuit() {
	if onHotkeyQuit != nil {
		onHotkeyQuit()
	}
}

// ── Platform interface implementation ───────────────────────────────────────

func platformRun() {
	C.setIconData(unsafe.Pointer(&iconPNG[0]), C.int(len(iconPNG)))
	C.createAndRunGUI()
}

func platformSetCursorPos(x, y int) {
	C.macSetCursorPos(C.int(x), C.int(y))
}

func platformGetCursorPos() (int, int) {
	var ox, oy C.int
	C.macGetCursorPos(&ox, &oy)
	return int(ox), int(oy)
}

func platformClick() {
	C.macClick()
}

func platformPreventSleep() {
	C.macPreventSleep()
}

func platformAllowSleep() {
	C.macAllowSleep()
}

func platformMoveButton(x, y int) {
	C.macMoveButton(C.int(x), C.int(y))
}

func platformClientToScreen(x, y int) (int, int) {
	var ox, oy C.int
	C.macClientToScreen(C.int(x), C.int(y), &ox, &oy)
	return int(ox), int(oy)
}

func platformSetButtonActive(isActive bool) {
	v := C.int(0)
	if isActive {
		v = 1
	}
	C.macSetButtonActive(v)
}

func platformReinforceTopmost() {
	C.macReinforceTopmost()
}

func platformQuit() {
	C.macQuit()
}
