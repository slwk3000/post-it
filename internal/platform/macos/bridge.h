#ifndef BRIDGE_H
#define BRIDGE_H

#include <stddef.h>

#ifdef __cplusplus
extern "C" {
#endif

void macosInitApp(void);
void macosRunApp(void);
void macosTerminateApp(void);

void* macosCreatePanel(const char* panelId, double x, double y, double w, double h, const char* html, int isFloating);
void macosUpdatePanelHTML(void* panelPtr, const char* html);
void macosEvaluateJS(void* panelPtr, const char* script);
void macosSetPanelVisible(void* panelPtr, int visible);
void macosSetPanelFrame(void* panelPtr, double x, double y, double w, double h);
void macosGetPanelFrame(void* panelPtr, double* x, double* y, double* w, double* h);
void macosClosePanel(void* panelPtr);

void macosStartDrag(void* panelPtr);
void macosMovePanel(void* panelPtr, double dx, double dy);
void macosResizePanel(void* panelPtr, double dw, double dh);

void macosSetupTray(const char* title);
void macosGetMousePos(double* x, double* y);
void macosRegisterHotkeys(void);

#ifdef __cplusplus
}
#endif

#endif // BRIDGE_H
