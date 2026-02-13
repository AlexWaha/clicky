//go:build darwin

package main

/*
#cgo CFLAGS: -x objective-c
#cgo LDFLAGS: -framework Cocoa -framework CoreGraphics -framework IOKit

#include <stdlib.h>
#import <Cocoa/Cocoa.h>
#import <CoreGraphics/CoreGraphics.h>
#import <IOKit/pwr_mgt/IOPMLib.h>

// ── Globals ─────────────────────────────────────────────────────────────────

static NSWindow     *mainWindow   = nil;
static NSButton     *aliveButton  = nil;
static IOPMAssertionID sleepAssertionID = 0;

// Forward declarations for Go callbacks
extern void goOnButtonClicked();
extern void goOnHotkeyQuit();

// ── Button action target ────────────────────────────────────────────────────

@interface ButtonTarget : NSObject
- (void)buttonClicked:(id)sender;
@end

@implementation ButtonTarget
- (void)buttonClicked:(id)sender {
    goOnButtonClicked();
}
@end

static ButtonTarget *btnTarget = nil;

// ── App delegate (Cmd+Q) ───────────────────────────────────────────────────

@interface AppDelegate : NSObject <NSApplicationDelegate>
@end

@implementation AppDelegate
- (NSApplicationTerminateReply)applicationShouldTerminate:(NSApplication *)sender {
    goOnHotkeyQuit();
    return NSTerminateCancel;
}
@end

static AppDelegate *appDel = nil;

// ── C functions called from Go ──────────────────────────────────────────────

static const void *iconPNGData = NULL;
static int iconPNGLength = 0;

static void setIconData(const void *data, int length) {
    iconPNGData = data;
    iconPNGLength = length;
}

static void macSetAppIconFromPNG(const void *data, int length) {
    NSData *pngData = [NSData dataWithBytes:data length:(NSUInteger)length];
    NSImage *icon = [[NSImage alloc] initWithData:pngData];
    if (icon) {
        [NSApp setApplicationIconImage:icon];
    }
}

static void createAndRunGUI() {
    @autoreleasepool {
        [NSApplication sharedApplication];
        [NSApp setActivationPolicy:NSApplicationActivationPolicyRegular];

        appDel = [[AppDelegate alloc] init];
        [NSApp setDelegate:appDel];

        // Screen geometry (macOS: origin = bottom-left)
        NSScreen *screen = [NSScreen mainScreen];
        CGFloat screenH = screen.frame.size.height;
        CGFloat winW = 300, winH = 300;
        CGFloat startX = (screen.frame.size.width - winW) / 2;
        CGFloat startY = (screenH - winH) / 2;

        NSRect frame = NSMakeRect(startX, startY, winW, winH);
        mainWindow = [[NSWindow alloc]
            initWithContentRect:frame
            styleMask:(NSWindowStyleMaskTitled | NSWindowStyleMaskClosable)
            backing:NSBackingStoreBuffered
            defer:NO];

        [mainWindow setTitle:@"KeepAlive (Cmd+Q = quit)"];
        [mainWindow setLevel:NSFloatingWindowLevel];
        [mainWindow setBackgroundColor:[NSColor colorWithRed:0x2B/255.0 green:0x2B/255.0 blue:0x2B/255.0 alpha:1.0]];

        // App icon from embedded PNG
        if (iconPNGData != NULL) {
            macSetAppIconFromPNG(iconPNGData, iconPNGLength);
        }

        NSView *content = [mainWindow contentView];

        // Alive button — centered
        CGFloat btnW = 80, btnH = 30;
        CGFloat btnX = (winW - btnW) / 2;
        CGFloat btnY = (winH - btnH) / 2;
        aliveButton = [[NSButton alloc] initWithFrame:NSMakeRect(btnX, btnY, btnW, btnH)];
        [aliveButton setTitle:@"Alive"];
        [aliveButton setBezelStyle:NSBezelStyleRounded];
        [aliveButton setWantsLayer:YES];
        [aliveButton.layer setBackgroundColor:[[NSColor colorWithRed:0.0 green:0x78/255.0 blue:0xD4/255.0 alpha:1.0] CGColor]];
        [aliveButton.layer setCornerRadius:4];
        NSMutableAttributedString *attrTitle = [[NSMutableAttributedString alloc] initWithString:@"Alive"];
        [attrTitle addAttribute:NSForegroundColorAttributeName value:[NSColor whiteColor] range:NSMakeRange(0, attrTitle.length)];
        [aliveButton setAttributedTitle:attrTitle];

        btnTarget = [[ButtonTarget alloc] init];
        [aliveButton setTarget:btnTarget];
        [aliveButton setAction:@selector(buttonClicked:)];
        [content addSubview:aliveButton];

        // Cmd+Q menu item
        NSMenu *menuBar = [[NSMenu alloc] init];
        NSMenuItem *appMenuItem = [[NSMenuItem alloc] init];
        [menuBar addItem:appMenuItem];
        NSMenu *appMenu = [[NSMenu alloc] init];
        NSMenuItem *quitItem = [[NSMenuItem alloc]
            initWithTitle:@"Quit KeepAlive"
            action:@selector(terminate:)
            keyEquivalent:@"q"];
        [appMenu addItem:quitItem];
        [appMenuItem setSubmenu:appMenu];
        [NSApp setMainMenu:menuBar];

        [mainWindow makeKeyAndOrderFront:nil];
        [NSApp activateIgnoringOtherApps:YES];

        [NSApp run];
    }
}

static void macSetCursorPos(int x, int y) {
    NSScreen *screen = [NSScreen mainScreen];
    CGFloat screenH = screen.frame.size.height;
    CGPoint pt = CGPointMake((CGFloat)x, (CGFloat)y);
    CGWarpMouseCursorPosition(pt);
}

static void macGetCursorPos(int *outX, int *outY) {
    NSPoint loc = [NSEvent mouseLocation];
    NSScreen *screen = [NSScreen mainScreen];
    CGFloat screenH = screen.frame.size.height;
    *outX = (int)loc.x;
    *outY = (int)(screenH - loc.y);
}

static void macClick() {
    NSPoint loc = [NSEvent mouseLocation];
    NSScreen *screen = [NSScreen mainScreen];
    CGFloat screenH = screen.frame.size.height;
    CGPoint pt = CGPointMake(loc.x, screenH - loc.y);

    CGEventRef down = CGEventCreateMouseEvent(NULL, kCGEventLeftMouseDown, pt, kCGMouseButtonLeft);
    CGEventRef up   = CGEventCreateMouseEvent(NULL, kCGEventLeftMouseUp,   pt, kCGMouseButtonLeft);
    CGEventPost(kCGHIDEventTap, down);
    CGEventPost(kCGHIDEventTap, up);
    CFRelease(down);
    CFRelease(up);
}

static void macPreventSleep() {
    if (sleepAssertionID == 0) {
        IOPMAssertionCreateWithName(
            kIOPMAssertionTypeNoDisplaySleep,
            kIOPMAssertionLevelOn,
            CFSTR("KeepAlive active"),
            &sleepAssertionID);
    }
}

static void macAllowSleep() {
    if (sleepAssertionID != 0) {
        IOPMAssertionRelease(sleepAssertionID);
        sleepAssertionID = 0;
    }
}

static void macMoveButton(int x, int y) {
    dispatch_async(dispatch_get_main_queue(), ^{
        NSRect frame = [aliveButton frame];
        // macOS: y is from bottom
        CGFloat contentH = [[mainWindow contentView] frame].size.height;
        [aliveButton setFrameOrigin:NSMakePoint((CGFloat)x, contentH - (CGFloat)y - frame.size.height)];
    });
}

static void macClientToScreen(int cx, int cy, int *outX, int *outY) {
    dispatch_sync(dispatch_get_main_queue(), ^{
        NSRect contentRect = [[mainWindow contentView] frame];
        CGFloat contentH = contentRect.size.height;
        // Convert client coords (top-left origin) to window coords (bottom-left origin)
        NSPoint winPt = NSMakePoint((CGFloat)cx, contentH - (CGFloat)cy);
        NSRect winRect = [mainWindow convertRectToScreen:NSMakeRect(winPt.x, winPt.y, 0, 0)];
        NSScreen *screen = [NSScreen mainScreen];
        CGFloat screenH = screen.frame.size.height;
        *outX = (int)winRect.origin.x;
        *outY = (int)(screenH - winRect.origin.y);
    });
}

static void macSetButtonActive(int isActive) {
    dispatch_async(dispatch_get_main_queue(), ^{
        if (isActive) {
            [aliveButton.layer setBackgroundColor:[[NSColor colorWithRed:0x10/255.0 green:0x7C/255.0 blue:0x10/255.0 alpha:1.0] CGColor]];
            NSMutableAttributedString *attrTitle = [[NSMutableAttributedString alloc] initWithString:@"\u25CF Active"];
            [attrTitle addAttribute:NSForegroundColorAttributeName value:[NSColor whiteColor] range:NSMakeRange(0, attrTitle.length)];
            [aliveButton setAttributedTitle:attrTitle];
        } else {
            [aliveButton.layer setBackgroundColor:[[NSColor colorWithRed:0.0 green:0x78/255.0 blue:0xD4/255.0 alpha:1.0] CGColor]];
            NSMutableAttributedString *attrTitle = [[NSMutableAttributedString alloc] initWithString:@"Alive"];
            [attrTitle addAttribute:NSForegroundColorAttributeName value:[NSColor whiteColor] range:NSMakeRange(0, attrTitle.length)];
            [aliveButton setAttributedTitle:attrTitle];
        }
    });
}

static void macReinforceTopmost() {
    dispatch_async(dispatch_get_main_queue(), ^{
        [mainWindow setLevel:NSFloatingWindowLevel];
    });
}

static void macQuit() {
    dispatch_async(dispatch_get_main_queue(), ^{
        [NSApp terminate:nil];
    });
}
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
