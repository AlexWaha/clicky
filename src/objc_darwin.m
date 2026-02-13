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

// ── Window delegate (close button) ─────────────────────────────────────────

@interface WindowDelegate : NSObject <NSWindowDelegate>
@end

@implementation WindowDelegate
- (BOOL)windowShouldClose:(NSWindow *)sender {
    [NSApp terminate:nil];
    return NO;
}
@end

static WindowDelegate *winDel = nil;

// ── App delegate (Cmd+Q) ───────────────────────────────────────────────────

@interface AppDelegate : NSObject <NSApplicationDelegate>
@end

@implementation AppDelegate
- (NSApplicationTerminateReply)applicationShouldTerminate:(NSApplication *)sender {
    goOnHotkeyQuit();
    return NSTerminateNow;
}
@end

static AppDelegate *appDel = nil;

// ── C functions called from Go ──────────────────────────────────────────────

static const void *iconPNGData = NULL;
static int iconPNGLength = 0;

void setIconData(const void *data, int length) {
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

void createAndRunGUI() {
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

        [mainWindow setTitle:@"Clicky"];
        winDel = [[WindowDelegate alloc] init];
        [mainWindow setDelegate:winDel];
        [mainWindow setLevel:NSFloatingWindowLevel];
        [mainWindow setBackgroundColor:[NSColor colorWithRed:0x2B/255.0 green:0x2B/255.0 blue:0x2B/255.0 alpha:1.0]];

        // App icon: macOS uses icon.icns from the .app bundle automatically

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

        // Quit hint label — bottom-right corner
        NSTextField *hintLabel = [NSTextField labelWithString:@"To close the App press Cmd+Q"];
        [hintLabel setTextColor:[NSColor colorWithWhite:1.0 alpha:0.35]];
        [hintLabel setFont:[NSFont systemFontOfSize:12]];
        [hintLabel sizeToFit];
        CGFloat hintPad = 12;
        CGFloat hintX = winW - hintLabel.frame.size.width - hintPad;
        CGFloat hintY = hintPad;
        [hintLabel setFrameOrigin:NSMakePoint(hintX, hintY)];
        [content addSubview:hintLabel];

        // Cmd+Q menu item
        NSMenu *menuBar = [[NSMenu alloc] init];
        NSMenuItem *appMenuItem = [[NSMenuItem alloc] init];
        [menuBar addItem:appMenuItem];
        NSMenu *appMenu = [[NSMenu alloc] init];
        NSMenuItem *quitItem = [[NSMenuItem alloc]
            initWithTitle:@"Quit Clicky"
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

void macSetCursorPos(int x, int y) {
    NSScreen *screen = [NSScreen mainScreen];
    CGFloat screenH = screen.frame.size.height;
    CGPoint pt = CGPointMake((CGFloat)x, (CGFloat)y);
    CGWarpMouseCursorPosition(pt);
}

void macGetCursorPos(int *outX, int *outY) {
    NSPoint loc = [NSEvent mouseLocation];
    NSScreen *screen = [NSScreen mainScreen];
    CGFloat screenH = screen.frame.size.height;
    *outX = (int)loc.x;
    *outY = (int)(screenH - loc.y);
}

void macClick() {
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

void macPreventSleep() {
    if (sleepAssertionID == 0) {
        IOPMAssertionCreateWithName(
            kIOPMAssertionTypeNoDisplaySleep,
            kIOPMAssertionLevelOn,
            CFSTR("KeepAlive active"),
            &sleepAssertionID);
    }
}

void macAllowSleep() {
    if (sleepAssertionID != 0) {
        IOPMAssertionRelease(sleepAssertionID);
        sleepAssertionID = 0;
    }
}

void macMoveButton(int x, int y) {
    dispatch_async(dispatch_get_main_queue(), ^{
        NSRect frame = [aliveButton frame];
        // macOS: y is from bottom
        CGFloat contentH = [[mainWindow contentView] frame].size.height;
        [aliveButton setFrameOrigin:NSMakePoint((CGFloat)x, contentH - (CGFloat)y - frame.size.height)];
    });
}

void macClientToScreen(int cx, int cy, int *outX, int *outY) {
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

void macSetButtonActive(int isActive) {
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

void macReinforceTopmost() {
    dispatch_async(dispatch_get_main_queue(), ^{
        [mainWindow setLevel:NSFloatingWindowLevel];
    });
}

void macQuit() {
    dispatch_async(dispatch_get_main_queue(), ^{
        [NSApp terminate:nil];
    });
}
