# 🖱️ Clicky — KeepAlive

> Tiny Windows utility that keeps your PC awake by simulating mouse clicks.

## ✨ Features

- 🟢 **Auto-clicker** — smooth cursor movement + periodic clicks
- 🔒 **Prevents sleep** — blocks display & system idle timeout
- 📌 **Always on top** — small 300×300 window stays visible
- ⌨️ **Ctrl+Q** — global hotkey to quit instantly
- 🎨 **Dark theme** — flat UI, no external dependencies

## 🚀 Usage

1. Launch `clicky.exe`
2. Click the **Alive** button — it turns green (**● Active**)
3. The cursor moves to button corners every ~4 seconds, simulating clicks
4. Press **Ctrl+Q** anywhere to quit

## 🔨 Build

```bash
go build -ldflags="-H windowsgui"
```

## 🎨 Design

| Element | Color |
|---------|-------|
| Background | `#2B2B2B` |
| Button (idle) | `#0078D4` |
| Button (active) | `#107C10` |
| Text | `#FFFFFF` |
| Font | Segoe UI, semi-bold |

Icon: blue circle with white center dot (embedded in .exe via `rsrc_windows_amd64.syso`).

## 📋 Requirements

- Windows 10+
- Go 1.21+
- No CGO, no external libraries — pure Win32 API via `syscall`
