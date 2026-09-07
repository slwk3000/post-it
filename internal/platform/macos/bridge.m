#import <Cocoa/Cocoa.h>
#import <WebKit/WebKit.h>
#import <Carbon/Carbon.h>
#import "bridge.h"

// Forward declarations of Go exported callbacks
extern void onWebMessage(char* panelId, char* messageJson);
extern void onTrayNewNote(void);
extern void onTrayToggleNotes(void);
extern void onTrayOpenMenu(void);
extern void onAppReopen(void);
extern void onHotKeyTriggered(int hotkeyId);
extern void onPanelMoved(char* panelId, double x, double y);
extern void onPanelResized(char* panelId, double w, double h);
extern void onAppWillTerminate(void);

// Custom NSPanel that allows key/focus input while borderless
@interface PostItPanel : NSPanel
@property (nonatomic, strong) NSString *panelId;
@property (nonatomic, strong) WKWebView *webView;
- (void)onWindowMoved:(NSNotification *)note;
- (void)onWindowResized:(NSNotification *)note;
@end

@implementation PostItPanel
- (BOOL)canBecomeKeyWindow {
    return YES;
}
- (BOOL)canBecomeMainWindow {
    return YES;
}
- (void)mouseDown:(NSEvent *)event {
    [NSApp activateIgnoringOtherApps:YES];
    [super mouseDown:event];
}
- (void)mouseEntered:(NSEvent *)event {
    [super mouseEntered:event];
    [NSApp activateIgnoringOtherApps:YES];
    [self makeKeyWindow];
}
- (BOOL)performKeyEquivalent:(NSEvent *)event {
    NSEventModifierFlags flags = [event modifierFlags] & NSEventModifierFlagDeviceIndependentFlagsMask;
    if (flags == NSEventModifierFlagCommand) {
        if ([[event charactersIgnoringModifiers] isEqualToString:@"q"]) {
            [NSApp terminate:nil];
            return YES;
        }
    }
    return [super performKeyEquivalent:event];
}
- (void)onWindowMoved:(NSNotification *)note {
    NSRect frame = [self frame];
    NSScreen *screen = [NSScreen mainScreen];
    CGFloat sHeight = screen ? screen.frame.size.height : 1080.0;
    double x = frame.origin.x;
    double y = sHeight - frame.origin.y - frame.size.height;
    onPanelMoved((char *)[self.panelId UTF8String], x, y);
}
- (void)onWindowResized:(NSNotification *)note {
    NSRect frame = [self frame];
    onPanelResized((char *)[self.panelId UTF8String], frame.size.width, frame.size.height);
}
@end

// Script message handler bridging JS to Go
@interface PostItMessageHandler : NSObject <WKScriptMessageHandler>
@property (nonatomic, weak) PostItPanel *panel;
@end

@implementation PostItMessageHandler
- (void)userContentController:(WKUserContentController *)userContentController didReceiveScriptMessage:(WKScriptMessage *)message {
    if ([message.body isKindOfClass:[NSString class]]) {
        onWebMessage((char *)[self.panel.panelId UTF8String], (char *)[(NSString *)message.body UTF8String]);
    }
}
@end

// App Delegate
@interface PostItAppDelegate : NSObject <NSApplicationDelegate>
@end

@implementation PostItAppDelegate
- (BOOL)applicationShouldHandleReopen:(NSApplication *)sender hasVisibleWindows:(BOOL)flag {
    onAppReopen();
    return YES;
}
- (void)applicationWillTerminate:(NSNotification *)notification {
    onAppWillTerminate();
}
@end

// Tray Target
@interface PostItTrayTarget : NSObject
@end

@implementation PostItTrayTarget
- (void)onNewNote:(id)sender {
    onTrayNewNote();
}
- (void)onToggleNotes:(id)sender {
    onTrayToggleNotes();
}
- (void)onOpenMenu:(id)sender {
    onTrayOpenMenu();
}
- (void)onQuit:(id)sender {
    [NSApp terminate:nil];
}
@end

static PostItAppDelegate *gAppDelegate = nil;
static PostItTrayTarget *gTrayTarget = nil;
static NSStatusItem *gStatusItem = nil;

static CGFloat getScreenHeight() {
    NSScreen *screen = [NSScreen mainScreen];
    return screen ? screen.frame.size.height : 1080.0;
}

void macosInitApp(void) {
    [NSApplication sharedApplication];
    // Run as background accessory app (no clutter in Dock by default)
    [NSApp setActivationPolicy:NSApplicationActivationPolicyAccessory];

    // Setup main menu with App (Quit Cmd+Q) and Edit menus
    NSMenu *mainMenu = [[NSMenu alloc] init];

    // Application Menu with Quit Post-it (Cmd+Q)
    NSMenuItem *appMenuItem = [[NSMenuItem alloc] init];
    NSMenu *appMenu = [[NSMenu alloc] initWithTitle:@"Post-it"];
    [appMenu addItemWithTitle:@"Encerrar Post-it" action:@selector(terminate:) keyEquivalent:@"q"];
    [appMenuItem setSubmenu:appMenu];
    [mainMenu addItem:appMenuItem];

    // Edit Menu for clipboard shortcuts inside textareas
    NSMenuItem *editMenuItem = [[NSMenuItem alloc] init];
    NSMenu *editMenu = [[NSMenu alloc] initWithTitle:@"Edit"];
    [editMenu addItemWithTitle:@"Undo" action:@selector(undo:) keyEquivalent:@"z"];
    [editMenu addItemWithTitle:@"Redo" action:@selector(redo:) keyEquivalent:@"Z"];
    [editMenu addItem:[NSMenuItem separatorItem]];
    [editMenu addItemWithTitle:@"Cut" action:@selector(cut:) keyEquivalent:@"x"];
    [editMenu addItemWithTitle:@"Copy" action:@selector(copy:) keyEquivalent:@"c"];
    [editMenu addItemWithTitle:@"Paste" action:@selector(paste:) keyEquivalent:@"v"];
    [editMenu addItemWithTitle:@"Select All" action:@selector(selectAll:) keyEquivalent:@"a"];
    [editMenuItem setSubmenu:editMenu];
    [mainMenu addItem:editMenuItem];
    [NSApp setMainMenu:mainMenu];

    // Global in-app key monitor for Cmd+Q
    [NSEvent addLocalMonitorForEventsMatchingMask:NSEventMaskKeyDown handler:^NSEvent *(NSEvent *event) {
        if (([event modifierFlags] & NSEventModifierFlagDeviceIndependentFlagsMask) == NSEventModifierFlagCommand) {
            if ([[event charactersIgnoringModifiers] isEqualToString:@"q"]) {
                [NSApp terminate:nil];
                return nil;
            }
        }
        return event;
    }];

    gAppDelegate = [[PostItAppDelegate alloc] init];
    [NSApp setDelegate:gAppDelegate];
    gTrayTarget = [[PostItTrayTarget alloc] init];
}

void macosRunApp(void) {
    [NSApp run];
}

void macosTerminateApp(void) {
    dispatch_async(dispatch_get_main_queue(), ^{
        [NSApp terminate:nil];
    });
}

void* macosCreatePanel(const char* panelId, double x, double y, double w, double h, const char* html, int isFloating) {
    __block PostItPanel *panel = nil;

    void (^createBlock)(void) = ^{
        CGFloat sHeight = getScreenHeight();
        CGFloat cocoaY = sHeight - y - h;

        NSRect frame = NSMakeRect(x, cocoaY, w, h);
        NSWindowStyleMask mask = NSWindowStyleMaskBorderless | NSWindowStyleMaskNonactivatingPanel;
        panel = [[PostItPanel alloc] initWithContentRect:frame
                                               styleMask:mask
                                                 backing:NSBackingStoreBuffered
                                                   defer:NO];

        panel.panelId = [NSString stringWithUTF8String:panelId];
        [panel setLevel:NSFloatingWindowLevel];
        [panel setOpaque:NO];
        [panel setBackgroundColor:[NSColor clearColor]];
        [panel setHasShadow:NO];
        [panel setCollectionBehavior:NSWindowCollectionBehaviorCanJoinAllSpaces | NSWindowCollectionBehaviorFullScreenAuxiliary];

        // WebKit Configuration
        WKWebViewConfiguration *config = [[WKWebViewConfiguration alloc] init];
        PostItMessageHandler *handler = [[PostItMessageHandler alloc] init];
        handler.panel = panel;
        [config.userContentController addScriptMessageHandler:handler name:@"postit"];

        WKWebView *wv = [[WKWebView alloc] initWithFrame:NSMakeRect(0, 0, w, h) configuration:config];
        [wv setValue:@NO forKey:@"drawsBackground"];
        if (@available(macOS 12.0, *)) {
            wv.underPageBackgroundColor = [NSColor clearColor];
        }
        [wv setAutoresizingMask:NSViewWidthSizable | NSViewHeightSizable];

        NSString *htmlStr = [NSString stringWithUTF8String:html];
        [wv loadHTMLString:htmlStr baseURL:nil];

        panel.webView = wv;
        [panel.contentView addSubview:wv];

        // Hover tracking area to activate panel on mouse enter
        NSTrackingArea *trackingArea = [[NSTrackingArea alloc]
            initWithRect:panel.contentView.bounds
            options:NSTrackingMouseEnteredAndExited | NSTrackingActiveAlways | NSTrackingInVisibleRect
            owner:panel
            userInfo:nil];
        [panel.contentView addTrackingArea:trackingArea];

        [[NSNotificationCenter defaultCenter] addObserver:panel
                                                 selector:@selector(onWindowMoved:)
                                                     name:NSWindowDidMoveNotification
                                                   object:panel];

        [[NSNotificationCenter defaultCenter] addObserver:panel
                                                 selector:@selector(onWindowResized:)
                                                     name:NSWindowDidResizeNotification
                                                   object:panel];

        [panel makeKeyAndOrderFront:nil];
    };

    if ([NSThread isMainThread]) {
        createBlock();
    } else {
        dispatch_sync(dispatch_get_main_queue(), createBlock);
    }

    return (__bridge_retained void*)panel;
}

void macosUpdatePanelHTML(void* panelPtr, const char* html) {
    if (!panelPtr) return;
    PostItPanel *panel = (__bridge PostItPanel*)panelPtr;
    NSString *htmlStr = [NSString stringWithUTF8String:html];

    dispatch_async(dispatch_get_main_queue(), ^{
        [panel.webView loadHTMLString:htmlStr baseURL:nil];
    });
}

void macosEvaluateJS(void* panelPtr, const char* script) {
    if (!panelPtr) return;
    PostItPanel *panel = (__bridge PostItPanel*)panelPtr;
    NSString *jsStr = [NSString stringWithUTF8String:script];

    dispatch_async(dispatch_get_main_queue(), ^{
        [panel.webView evaluateJavaScript:jsStr completionHandler:nil];
    });
}

void macosSetPanelVisible(void* panelPtr, int visible) {
    if (!panelPtr) return;
    PostItPanel *panel = (__bridge PostItPanel*)panelPtr;

    dispatch_async(dispatch_get_main_queue(), ^{
        if (visible) {
            [panel makeKeyAndOrderFront:nil];
        } else {
            [panel orderOut:nil];
        }
    });
}

void macosSetPanelFrame(void* panelPtr, double x, double y, double w, double h) {
    if (!panelPtr) return;
    PostItPanel *panel = (__bridge PostItPanel*)panelPtr;
    CGFloat sHeight = getScreenHeight();
    CGFloat cocoaY = sHeight - y - h;

    dispatch_async(dispatch_get_main_queue(), ^{
        [panel setFrame:NSMakeRect(x, cocoaY, w, h) display:YES];
    });
}

void macosGetPanelFrame(void* panelPtr, double* x, double* y, double* w, double* h) {
    if (!panelPtr) return;
    PostItPanel *panel = (__bridge PostItPanel*)panelPtr;

    void (^frameBlock)(void) = ^{
        NSRect frame = [panel frame];
        CGFloat sHeight = getScreenHeight();
        *x = frame.origin.x;
        *y = sHeight - frame.origin.y - frame.size.height;
        *w = frame.size.width;
        *h = frame.size.height;
    };

    if ([NSThread isMainThread]) {
        frameBlock();
    } else {
        dispatch_sync(dispatch_get_main_queue(), frameBlock);
    }
}

void macosClosePanel(void* panelPtr) {
    if (!panelPtr) return;
    PostItPanel *panel = (__bridge_transfer PostItPanel*)panelPtr;

    dispatch_async(dispatch_get_main_queue(), ^{
        [[NSNotificationCenter defaultCenter] removeObserver:panel];
        [panel.webView.configuration.userContentController removeScriptMessageHandlerForName:@"postit"];
        [panel close];
    });
}

void macosFocusPanel(void* panelPtr) {
    if (!panelPtr) return;
    PostItPanel *panel = (__bridge PostItPanel*)panelPtr;
    dispatch_async(dispatch_get_main_queue(), ^{
        [NSApp activateIgnoringOtherApps:YES];
        [panel orderFrontRegardless];
        [panel makeKeyWindow];
        if (panel.webView) {
            [panel makeFirstResponder:panel.webView];
            [panel.webView evaluateJavaScript:@"var el = document.getElementById('note-text'); if (el) { el.focus(); el.setSelectionRange(el.value.length, el.value.length); }" completionHandler:nil];
        }
    });
}

void macosStartDrag(void* panelPtr) {
    if (!panelPtr) return;
    PostItPanel *panel = (__bridge PostItPanel*)panelPtr;

    dispatch_async(dispatch_get_main_queue(), ^{
        [panel orderFrontRegardless];
        NSEvent *event = [NSApp currentEvent];
        if (event) {
            [panel performWindowDragWithEvent:event];
        }
    });
}

void macosMovePanel(void* panelPtr, double dx, double dy) {
    if (!panelPtr) return;
    PostItPanel *panel = (__bridge PostItPanel*)panelPtr;

    dispatch_async(dispatch_get_main_queue(), ^{
        NSRect frame = [panel frame];
        frame.origin.x += dx;
        frame.origin.y -= dy; // In Cocoa screen coordinates, Y is inverted relative to web
        [panel setFrameOrigin:frame.origin];
    });
}

void macosResizePanel(void* panelPtr, double dw, double dh) {
    if (!panelPtr) return;
    PostItPanel *panel = (__bridge PostItPanel*)panelPtr;

    dispatch_async(dispatch_get_main_queue(), ^{
        NSRect frame = [panel frame];
        CGFloat minW = 200.0;
        CGFloat minH = 180.0;

        CGFloat newW = frame.size.width + dw;
        if (newW < minW) newW = minW;

        CGFloat newH = frame.size.height + dh;
        if (newH < minH) newH = minH;

        CGFloat actualDh = newH - frame.size.height;
        frame.origin.y -= actualDh;
        frame.size.width = newW;
        frame.size.height = newH;

        [panel setFrame:frame display:YES];
    });
}

void macosSetupTray(const char* title) {
    dispatch_async(dispatch_get_main_queue(), ^{
        if (!gStatusItem) {
            gStatusItem = [[NSStatusBar systemStatusBar] statusItemWithLength:NSSquareStatusItemLength];
        }
        if (!gTrayTarget) {
            gTrayTarget = [[PostItTrayTarget alloc] init];
        }

        // Mini post-it white square icon
        NSSize iconSize = NSMakeSize(18, 18);
        NSImage *icon = [NSImage imageWithSize:iconSize flipped:NO drawingHandler:^BOOL(NSRect dstRect) {
            NSRect rect = NSMakeRect(2.5, 2.5, 13.0, 13.0);
            NSBezierPath *path = [NSBezierPath bezierPathWithRoundedRect:rect xRadius:1.5 yRadius:1.5];
            [[NSColor whiteColor] setFill];
            [path fill];
            [[NSColor colorWithWhite:0.2 alpha:0.35] setStroke];
            [path setLineWidth:1.0];
            [path stroke];
            return YES;
        }];
        [icon setTemplate:NO];
        gStatusItem.button.image = icon;
        gStatusItem.button.title = @"";

        NSMenu *menu = [[NSMenu alloc] initWithTitle:@"Post-it"];

        NSMenuItem *itemNew = [[NSMenuItem alloc] initWithTitle:@"Criar Nova Nota" action:@selector(onNewNote:) keyEquivalent:@"n"];
        [itemNew setKeyEquivalentModifierMask:NSEventModifierFlagCommand | NSEventModifierFlagShift];
        [itemNew setTarget:gTrayTarget];
        [menu addItem:itemNew];

        NSMenuItem *itemToggle = [[NSMenuItem alloc] initWithTitle:@"Ocultar / Exibir Notas" action:@selector(onToggleNotes:) keyEquivalent:@"p"];
        [itemToggle setKeyEquivalentModifierMask:NSEventModifierFlagCommand | NSEventModifierFlagShift];
        [itemToggle setTarget:gTrayTarget];
        [menu addItem:itemToggle];

        NSMenuItem *itemMenu = [[NSMenuItem alloc] initWithTitle:@"Ajustes" action:@selector(onOpenMenu:) keyEquivalent:@"a"];
        [itemMenu setKeyEquivalentModifierMask:NSEventModifierFlagCommand | NSEventModifierFlagShift];
        [itemMenu setTarget:gTrayTarget];
        [menu addItem:itemMenu];

        [menu addItem:[NSMenuItem separatorItem]];

        NSMenuItem *itemQuit = [[NSMenuItem alloc] initWithTitle:@"Sair" action:@selector(onQuit:) keyEquivalent:@"q"];
        [itemQuit setKeyEquivalentModifierMask:NSEventModifierFlagCommand];
        [itemQuit setTarget:gTrayTarget];
        [menu addItem:itemQuit];

        gStatusItem.menu = menu;
    });
}

void macosGetMousePos(double* x, double* y) {
    NSPoint p = [NSEvent mouseLocation];
    *x = p.x;
    *y = p.y;
}

static OSStatus hotKeyCarbonHandler(EventHandlerCallRef nextHandler, EventRef theEvent, void *userData) {
    EventHotKeyID hkID;
    GetEventParameter(theEvent, kEventParamDirectObject, typeEventHotKeyID, NULL, sizeof(hkID), NULL, &hkID);
    onHotKeyTriggered((int)hkID.id);
    return noErr;
}

void macosRegisterHotkeys(void) {
    EventTypeSpec eventType;
    eventType.eventClass = kEventClassKeyboard;
    eventType.eventKind = kEventHotKeyPressed;
    InstallApplicationEventHandler(&hotKeyCarbonHandler, 1, &eventType, NULL, NULL);

    EventHotKeyRef ref1, ref2, ref3, ref4, ref5, ref6;
    EventHotKeyID id1 = { 'POST', 1 };
    EventHotKeyID id2 = { 'POST', 2 };
    EventHotKeyID id3 = { 'POST', 3 };
    EventHotKeyID id4 = { 'POST', 4 };
    EventHotKeyID id5 = { 'POST', 5 };
    EventHotKeyID id6 = { 'POST', 6 };

    // Cmd+Shift+P (kVK_ANSI_P = 35): Toggle notes
    RegisterEventHotKey(35, cmdKey | shiftKey, id1, GetApplicationEventTarget(), 0, &ref1);
    // Cmd+Shift+N (kVK_ANSI_N = 45): New note
    RegisterEventHotKey(45, cmdKey | shiftKey, id2, GetApplicationEventTarget(), 0, &ref2);
    // Cmd+Shift+A (kVK_ANSI_A = 0): Toggle Ajustes
    RegisterEventHotKey(0, cmdKey | shiftKey, id3, GetApplicationEventTarget(), 0, &ref3);
    // Cmd+Shift+D (kVK_ANSI_D = 2): Delete note
    RegisterEventHotKey(2, cmdKey | shiftKey, id4, GetApplicationEventTarget(), 0, &ref4);
    // Cmd+Shift+U (kVK_ANSI_U = 32): Next note
    RegisterEventHotKey(32, cmdKey | shiftKey, id5, GetApplicationEventTarget(), 0, &ref5);
    // Cmd+Shift+R (kVK_ANSI_R = 15): Previous note
    RegisterEventHotKey(15, cmdKey | shiftKey, id6, GetApplicationEventTarget(), 0, &ref6);
}
