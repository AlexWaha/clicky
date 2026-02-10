package main

import (
	"runtime"
	"sync/atomic"
	"syscall"
	"time"
	"unsafe"
)

// ── Win32 constants ─────────────────────────────────────────────────────────

const (
	CS_HREDRAW = 0x0002
	CS_VREDRAW = 0x0001

	WS_CAPTION    = 0x00C00000
	WS_SYSMENU    = 0x00080000
	WS_VISIBLE    = 0x10000000
	WS_CHILD      = 0x40000000
	WS_TABSTOP    = 0x00010000
	WS_EX_TOPMOST = 0x00000008

	BS_OWNERDRAW = 0x0000000B

	SW_SHOW = 5

	WM_DESTROY    = 0x0002
	WM_ERASEBKGND = 0x0014
	WM_DRAWITEM   = 0x002B
	WM_SETICON    = 0x0080
	WM_COMMAND    = 0x0111
	WM_CTLCOLORBTN = 0x0135
	WM_HOTKEY     = 0x0312

	BN_CLICKED = 0

	SWP_NOZORDER   = 0x0004
	SWP_NOSIZE     = 0x0001
	SWP_NOMOVE     = 0x0002
	SWP_SHOWWINDOW = 0x0040

	HWND_TOPMOST = ^uintptr(0) // -1

	SPI_GETWORKAREA = 0x0030

	MOUSEEVENTF_LEFTDOWN = 0x0002
	MOUSEEVENTF_LEFTUP   = 0x0004

	ES_CONTINUOUS       = 0x80000000
	ES_DISPLAY_REQUIRED = 0x00000002
	ES_SYSTEM_REQUIRED  = 0x00000001

	MOD_CONTROL = 0x0002
	VK_Q        = 0x51

	IDC_ARROW     = 32512
	COLOR_BTNFACE = 15

	ICON_SMALL = 0
	ICON_BIG   = 1

	TRANSPARENT  = 1
	DT_CENTER    = 0x0001
	DT_VCENTER   = 0x0004
	DT_SINGLELINE = 0x0020
	FW_SEMIBOLD  = 600

	BTN_ID  = 1
	HK_QUIT = 1

	clientW = 300
	clientH = 300
	btnW    = 80
	btnH    = 30
	pad     = 10
)

// ── Win32 types ─────────────────────────────────────────────────────────────

type WNDCLASSEXW struct {
	CbSize        uint32
	Style         uint32
	LpfnWndProc   uintptr
	CbClsExtra    int32
	CbWndExtra    int32
	HInstance     syscall.Handle
	HIcon         syscall.Handle
	HCursor       syscall.Handle
	HbrBackground syscall.Handle
	LpszMenuName  *uint16
	LpszClassName *uint16
	HIconSm       syscall.Handle
}

type POINT struct {
	X, Y int32
}

type RECT struct {
	Left, Top, Right, Bottom int32
}

type MSG struct {
	Hwnd    syscall.Handle
	Message uint32
	WParam  uintptr
	LParam  uintptr
	Time    uint32
	Pt      POINT
}

type DRAWITEMSTRUCT struct {
	CtlType    uint32
	CtlID      uint32
	ItemID     uint32
	ItemAction uint32
	ItemState  uint32
	HwndItem   syscall.Handle
	HDC        syscall.Handle
	RcItem     RECT
	ItemData   uintptr
}

type ICONINFO struct {
	FIcon    uint32
	XHotspot uint32
	YHotspot uint32
	HbmMask  syscall.Handle
	HbmColor syscall.Handle
}

// ── DLL procs ───────────────────────────────────────────────────────────────

var (
	user32   = syscall.NewLazyDLL("user32.dll")
	kernel32 = syscall.NewLazyDLL("kernel32.dll")
	gdi32    = syscall.NewLazyDLL("gdi32.dll")

	pRegisterClassExW      = user32.NewProc("RegisterClassExW")
	pCreateWindowExW       = user32.NewProc("CreateWindowExW")
	pDefWindowProcW        = user32.NewProc("DefWindowProcW")
	pGetMessageW           = user32.NewProc("GetMessageW")
	pTranslateMessage      = user32.NewProc("TranslateMessage")
	pDispatchMessageW      = user32.NewProc("DispatchMessageW")
	pPostQuitMessage       = user32.NewProc("PostQuitMessage")
	pDestroyWindow         = user32.NewProc("DestroyWindow")
	pShowWindow            = user32.NewProc("ShowWindow")
	pUpdateWindow          = user32.NewProc("UpdateWindow")
	pMoveWindow            = user32.NewProc("MoveWindow")
	pSetCursorPos          = user32.NewProc("SetCursorPos")
	pClientToScreen        = user32.NewProc("ClientToScreen")
	pLoadCursorW           = user32.NewProc("LoadCursorW")
	pRegisterHotKey        = user32.NewProc("RegisterHotKey")
	pUnregisterHotKey      = user32.NewProc("UnregisterHotKey")
	pAdjustWindowRectEx    = user32.NewProc("AdjustWindowRectEx")
	pSystemParametersInfoW = user32.NewProc("SystemParametersInfoW")
	pSetWindowTextW        = user32.NewProc("SetWindowTextW")
	pMouseEvent            = user32.NewProc("mouse_event")
	pGetModuleHandleW      = kernel32.NewProc("GetModuleHandleW")
	pSetThreadExecutionState = kernel32.NewProc("SetThreadExecutionState")
	pSetProcessDPIAware    = user32.NewProc("SetProcessDPIAware")
	pGetCursorPos          = user32.NewProc("GetCursorPos")
	pFillRect              = user32.NewProc("FillRect")
	pDrawTextW             = user32.NewProc("DrawTextW")
	pCreateIconIndirect    = user32.NewProc("CreateIconIndirect")
	pSendMessageW          = user32.NewProc("SendMessageW")
	pGetClientRect         = user32.NewProc("GetClientRect")
	pInvalidateRect        = user32.NewProc("InvalidateRect")
	pSetWindowPos          = user32.NewProc("SetWindowPos")

	pCreateSolidBrush = gdi32.NewProc("CreateSolidBrush")
	pDeleteObject     = gdi32.NewProc("DeleteObject")
	pSetBkMode        = gdi32.NewProc("SetBkMode")
	pSetTextColor     = gdi32.NewProc("SetTextColor")
	pSelectObject     = gdi32.NewProc("SelectObject")
	pCreateFontW      = gdi32.NewProc("CreateFontW")
	pCreateBitmap     = gdi32.NewProc("CreateBitmap")
)

// ── Globals ─────────────────────────────────────────────────────────────────

var (
	hWndMain syscall.Handle
	hWndBtn  syscall.Handle
	active   atomic.Bool

	hFont       syscall.Handle
	hBrushBg    syscall.Handle // #2B2B2B dark background
	hBrushBlue  syscall.Handle // #0078D4 button idle
	hBrushGreen syscall.Handle // #107C10 button active
)

// ── Helpers ─────────────────────────────────────────────────────────────────

func utf16(s string) *uint16 {
	p, _ := syscall.UTF16PtrFromString(s)
	return p
}

func loword(l uintptr) uint16 { return uint16(l) }
func hiword(l uintptr) uint16 { return uint16(l >> 16) }

func rgb(r, g, b byte) uintptr {
	return uintptr(uint32(r) | uint32(g)<<8 | uint32(b)<<16)
}

// ── GDI resource creation ───────────────────────────────────────────────────

func createCustomFont() syscall.Handle {
	h, _, _ := pCreateFontW.Call(
		uintptr(uint32(0xFFFFFFEE)), // -18 (height)
		0, 0, 0,
		FW_SEMIBOLD,
		0, 0, 0, // italic, underline, strikeout
		1,        // DEFAULT_CHARSET
		0, 0, 4,  // out, clip, ANTIALIASED_QUALITY
		0,
		uintptr(unsafe.Pointer(utf16("Segoe UI"))),
	)
	return syscall.Handle(h)
}

func createAppIcon() syscall.Handle {
	const sz = 32
	// Build 32x32 pixel buffers (1 bit per pixel for mask, 32 bits per pixel for color)
	maskBits := make([]byte, sz*sz/8)
	colorBits := make([]byte, sz*sz*4) // BGRA

	// Draw a blue circle with white center dot
	cx, cy := float64(sz/2), float64(sz/2)
	for y := 0; y < sz; y++ {
		for x := 0; x < sz; x++ {
			dx := float64(x) - cx + 0.5
			dy := float64(y) - cy + 0.5
			dist := dx*dx + dy*dy
			off := (y*sz + x) * 4

			if dist <= 14*14 { // outer circle radius=14
				// Mask: 0 = visible
				if dist <= 3*3 { // center dot radius=3
					// White dot
					colorBits[off+0] = 0xFF // B
					colorBits[off+1] = 0xFF // G
					colorBits[off+2] = 0xFF // R
					colorBits[off+3] = 0xFF // A
				} else {
					// Blue #0078D4
					colorBits[off+0] = 0xD4 // B
					colorBits[off+1] = 0x78 // G
					colorBits[off+2] = 0x00 // R
					colorBits[off+3] = 0xFF // A
				}
			} else {
				// Outside circle — transparent, set mask bit to 1
				byteIdx := (y*sz + x) / 8
				bitIdx := uint(7 - (x % 8))
				maskBits[byteIdx] |= 1 << bitIdx
			}
		}
	}

	hMask, _, _ := pCreateBitmap.Call(sz, sz, 1, 1, uintptr(unsafe.Pointer(&maskBits[0])))
	hColor, _, _ := pCreateBitmap.Call(sz, sz, 1, 32, uintptr(unsafe.Pointer(&colorBits[0])))

	ii := ICONINFO{
		FIcon:    1,
		HbmMask:  syscall.Handle(hMask),
		HbmColor: syscall.Handle(hColor),
	}
	hIcon, _, _ := pCreateIconIndirect.Call(uintptr(unsafe.Pointer(&ii)))

	pDeleteObject.Call(hMask)
	pDeleteObject.Call(hColor)

	return syscall.Handle(hIcon)
}

// ── WndProc ─────────────────────────────────────────────────────────────────

func wndProc(hwnd syscall.Handle, msg uint32, wParam, lParam uintptr) uintptr {
	switch msg {
	case WM_ERASEBKGND:
		var rc RECT
		pGetClientRect.Call(uintptr(hwnd), uintptr(unsafe.Pointer(&rc)))
		pFillRect.Call(wParam, uintptr(unsafe.Pointer(&rc)), uintptr(hBrushBg))
		return 1

	case WM_CTLCOLORBTN:
		return 0

	case WM_DRAWITEM:
		di := (*DRAWITEMSTRUCT)(unsafe.Pointer(lParam))
		// Choose brush based on active state
		brush := hBrushBlue
		text := "Alive"
		if active.Load() {
			brush = hBrushGreen
			text = "\u25CF Active"
		}
		pFillRect.Call(uintptr(di.HDC), uintptr(unsafe.Pointer(&di.RcItem)), uintptr(brush))
		pSetBkMode.Call(uintptr(di.HDC), TRANSPARENT)
		pSetTextColor.Call(uintptr(di.HDC), rgb(0xFF, 0xFF, 0xFF))
		pSelectObject.Call(uintptr(di.HDC), uintptr(hFont))
		t := utf16(text)
		pDrawTextW.Call(
			uintptr(di.HDC),
			uintptr(unsafe.Pointer(t)),
			uintptr(uint32(0xFFFFFFFF)), // -1
			uintptr(unsafe.Pointer(&di.RcItem)),
			DT_CENTER|DT_VCENTER|DT_SINGLELINE,
		)
		return 1

	case WM_COMMAND:
		if hiword(wParam) == BN_CLICKED && loword(wParam) == BTN_ID {
			if !active.Load() {
				active.Store(true)
				pInvalidateRect.Call(uintptr(hWndBtn), 0, 1)
				go aliveLoop()
			}
		}
		return 0

	case WM_HOTKEY:
		if wParam == HK_QUIT {
			active.Store(false)
			pDestroyWindow.Call(uintptr(hWndMain))
		}
		return 0

	case WM_DESTROY:
		active.Store(false)
		pUnregisterHotKey.Call(uintptr(hWndMain), HK_QUIT)
		pSetThreadExecutionState.Call(ES_CONTINUOUS)
		pDeleteObject.Call(uintptr(hFont))
		pDeleteObject.Call(uintptr(hBrushBg))
		pDeleteObject.Call(uintptr(hBrushBlue))
		pDeleteObject.Call(uintptr(hBrushGreen))
		pPostQuitMessage.Call(0)
		return 0
	}

	ret, _, _ := pDefWindowProcW.Call(uintptr(hwnd), uintptr(msg), wParam, lParam)
	return ret
}

// ── Button positions inside the 300x300 client area ─────────────────────────

var btnCorners = [4]POINT{
	{pad, pad},                                    // top-left
	{clientW - btnW - pad, pad},                   // top-right
	{clientW - btnW - pad, clientH - btnH - pad},  // bottom-right
	{pad, clientH - btnH - pad},                   // bottom-left
}

// ── Smooth cursor movement ──────────────────────────────────────────────────

func moveCursorSmooth(toX, toY int32) {
	var from POINT
	pGetCursorPos.Call(uintptr(unsafe.Pointer(&from)))

	const steps = 20
	const dt = 5 * time.Millisecond

	for i := int32(1); i <= steps; i++ {
		x := from.X + (toX-from.X)*i/steps
		y := from.Y + (toY-from.Y)*i/steps
		pSetCursorPos.Call(uintptr(x), uintptr(y))
		time.Sleep(dt)
	}
}

// ── Alive loop ──────────────────────────────────────────────────────────────

func aliveLoop() {
	runtime.LockOSThread()

	// Prevent sleep
	pSetThreadExecutionState.Call(ES_CONTINUOUS | ES_DISPLAY_REQUIRED | ES_SYSTEM_REQUIRED)

	idx := 0
	for active.Load() {
		c := btnCorners[idx%4]

		// Move button inside the window (MoveWindow for child = client coords)
		pMoveWindow.Call(uintptr(hWndBtn), uintptr(c.X), uintptr(c.Y), btnW, btnH, 1)
		// Reinforce topmost status
		pSetWindowPos.Call(uintptr(hWndMain), HWND_TOPMOST, 0, 0, 0, 0, SWP_NOMOVE|SWP_NOSIZE)
		time.Sleep(400 * time.Millisecond)
		if !active.Load() {
			break
		}

		// Get button center in screen coordinates
		var pt POINT
		pt.X = c.X + btnW/2
		pt.Y = c.Y + btnH/2
		pClientToScreen.Call(uintptr(hWndMain), uintptr(unsafe.Pointer(&pt)))

		// Move cursor to button center (smooth)
		moveCursorSmooth(pt.X, pt.Y)
		time.Sleep(200 * time.Millisecond)
		if !active.Load() {
			break
		}

		// Simulate click
		pMouseEvent.Call(MOUSEEVENTF_LEFTDOWN, 0, 0, 0, 0)
		pMouseEvent.Call(MOUSEEVENTF_LEFTUP, 0, 0, 0, 0)

		time.Sleep(3 * time.Second)
		idx++
	}
}

// ── Main ────────────────────────────────────────────────────────────────────

func main() {
	runtime.LockOSThread()

	pSetProcessDPIAware.Call()

	// Create GDI resources
	hFont = createCustomFont()
	hBrushBg, _, _ = createSolidBrush(rgb(0x2B, 0x2B, 0x2B))
	hBrushBlue, _, _ = createSolidBrush(rgb(0x00, 0x78, 0xD4))
	hBrushGreen, _, _ = createSolidBrush(rgb(0x10, 0x7C, 0x10))

	hIcon := createAppIcon()

	hInst, _, _ := pGetModuleHandleW.Call(0)

	className := utf16("KeepAliveClass")
	cursor, _, _ := pLoadCursorW.Call(0, IDC_ARROW)

	wc := WNDCLASSEXW{
		CbSize:        uint32(unsafe.Sizeof(WNDCLASSEXW{})),
		Style:         CS_HREDRAW | CS_VREDRAW,
		LpfnWndProc:   syscall.NewCallback(wndProc),
		HInstance:     syscall.Handle(hInst),
		HCursor:       syscall.Handle(cursor),
		HbrBackground: hBrushBg,
		LpszClassName: className,
		HIcon:         syscall.Handle(hIcon),
		HIconSm:       syscall.Handle(hIcon),
	}
	pRegisterClassExW.Call(uintptr(unsafe.Pointer(&wc)))

	// Calculate window size for 300x300 client area
	style := uint32(WS_CAPTION | WS_SYSMENU)
	exStyle := uint32(WS_EX_TOPMOST)

	rc := RECT{0, 0, clientW, clientH}
	pAdjustWindowRectEx.Call(
		uintptr(unsafe.Pointer(&rc)),
		uintptr(style), 0, uintptr(exStyle),
	)
	winW := rc.Right - rc.Left
	winH := rc.Bottom - rc.Top

	// Center on screen
	var wa RECT
	pSystemParametersInfoW.Call(SPI_GETWORKAREA, 0, uintptr(unsafe.Pointer(&wa)), 0)
	startX := (wa.Right-wa.Left-winW)/2 + wa.Left
	startY := (wa.Bottom-wa.Top-winH)/2 + wa.Top

	hwnd, _, _ := pCreateWindowExW.Call(
		uintptr(exStyle),
		uintptr(unsafe.Pointer(className)),
		uintptr(unsafe.Pointer(utf16("KeepAlive (Ctrl+Q = quit)"))),
		uintptr(style|WS_VISIBLE),
		uintptr(startX), uintptr(startY),
		uintptr(winW), uintptr(winH),
		0, 0, hInst, 0,
	)
	hWndMain = syscall.Handle(hwnd)

	// Set icon on window
	pSendMessageW.Call(uintptr(hWndMain), WM_SETICON, ICON_SMALL, uintptr(hIcon))
	pSendMessageW.Call(uintptr(hWndMain), WM_SETICON, ICON_BIG, uintptr(hIcon))

	// Create button — starts at center of client area, owner-draw style
	btnStartX := (clientW - btnW) / 2
	btnStartY := (clientH - btnH) / 2
	btn, _, _ := pCreateWindowExW.Call(
		0,
		uintptr(unsafe.Pointer(utf16("BUTTON"))),
		uintptr(unsafe.Pointer(utf16("Alive"))),
		uintptr(WS_CHILD|WS_VISIBLE|WS_TABSTOP|BS_OWNERDRAW),
		uintptr(btnStartX), uintptr(btnStartY),
		btnW, btnH,
		uintptr(hWndMain),
		BTN_ID,
		hInst, 0,
	)
	hWndBtn = syscall.Handle(btn)

	// Register Ctrl+Q global hotkey
	pRegisterHotKey.Call(uintptr(hWndMain), HK_QUIT, MOD_CONTROL, VK_Q)

	pShowWindow.Call(uintptr(hWndMain), SW_SHOW)
	pUpdateWindow.Call(uintptr(hWndMain))

	// Message loop
	var msg MSG
	for {
		ret, _, _ := pGetMessageW.Call(uintptr(unsafe.Pointer(&msg)), 0, 0, 0)
		if ret == 0 || int32(ret) == -1 {
			break
		}
		pTranslateMessage.Call(uintptr(unsafe.Pointer(&msg)))
		pDispatchMessageW.Call(uintptr(unsafe.Pointer(&msg)))
	}
}

// createSolidBrush wraps the GDI call, returning handle + raw results.
func createSolidBrush(color uintptr) (syscall.Handle, uintptr, error) {
	h, _, err := pCreateSolidBrush.Call(color)
	return syscall.Handle(h), h, err
}
