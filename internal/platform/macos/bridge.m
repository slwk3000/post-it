#import <Cocoa/Cocoa.h>
#import <WebKit/WebKit.h>
#import "bridge.h"

// Forward declarations of Go exported callbacks
extern void onWebMessage(char* panelId, char* messageJson);
extern void onTrayNewNote(void);
extern void onTrayToggleNotes(void);
extern void onTrayOpenMenu(void);
extern void onAppReopen(void);

// Custom NSPanel that allows key/focus input while borderless
@interface PostItPanel : NSPanel
@property (nonatomic, strong) NSString *panelId;
@property (nonatomic, strong) WKWebView *webView;
@end

@implementation PostItPanel
- (BOOL)canBecomeKeyWindow {
    return YES;
}
- (BOOL)canBecomeMainWindow {
    return YES;
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
        if (isFloating) {
            [panel setLevel:NSFloatingWindowLevel];
        } else {
            [panel setLevel:NSFloatingWindowLevel + 1];
        }
        [panel setOpaque:NO];
        [panel setBackgroundColor:[NSColor clearColor]];
        [panel setHasShadow:YES];
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
        [panel.webView.configuration.userContentController removeScriptMessageHandlerForName:@"postit"];
        [panel close];
    });
}

void macosStartDrag(void* panelPtr) {
    if (!panelPtr) return;
    PostItPanel *panel = (__bridge PostItPanel*)panelPtr;

    dispatch_async(dispatch_get_main_queue(), ^{
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

void macosSetupTray(const char* title) {
    dispatch_async(dispatch_get_main_queue(), ^{
        if (!gStatusItem) {
            gStatusItem = [[NSStatusBar systemStatusBar] statusItemWithLength:NSVariableStatusItemLength];
        }
        gStatusItem.button.title = [NSString stringWithUTF8String:title];

        NSMenu *menu = [[NSMenu alloc] initWithTitle:@"Post-it"];

        NSMenuItem *itemNew = [[NSMenuItem alloc] initWithTitle:@"📝 Criar Nova Nota" action:@selector(onNewNote:) keyEquivalent:@"n"];
        [itemNew setKeyEquivalentModifierMask:NSEventModifierFlagCommand | NSEventModifierFlagShift];
        [itemNew setTarget:gTrayTarget];
        [menu addItem:itemNew];

        NSMenuItem *itemToggle = [[NSMenuItem alloc] initWithTitle:@"👁️ Ocultar / Exibir Notas" action:@selector(onToggleNotes:) keyEquivalent:@"p"];
        [itemToggle setKeyEquivalentModifierMask:NSEventModifierFlagCommand | NSEventModifierFlagShift];
        [itemToggle setTarget:gTrayTarget];
        [menu addItem:itemToggle];

        NSMenuItem *itemMenu = [[NSMenuItem alloc] initWithTitle:@"⚙️ Configurações / Menu" action:@selector(onOpenMenu:) keyEquivalent:@""];
        [itemMenu setTarget:gTrayTarget];
        [menu addItem:itemMenu];

        [menu addItem:[NSMenuItem separatorItem]];

        NSMenuItem *itemQuit = [[NSMenuItem alloc] initWithTitle:@"❌ Sair" action:@selector(onQuit:) keyEquivalent:@"q"];
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
