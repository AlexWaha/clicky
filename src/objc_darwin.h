#ifndef OBJC_DARWIN_H
#define OBJC_DARWIN_H

void setIconData(const void *data, int length);
void createAndRunGUI(void);
void macSetCursorPos(int x, int y);
void macGetCursorPos(int *outX, int *outY);
void macClick(void);
void macPreventSleep(void);
void macAllowSleep(void);
void macMoveButton(int x, int y);
void macClientToScreen(int cx, int cy, int *outX, int *outY);
void macSetButtonActive(int isActive);
void macReinforceTopmost(void);
void macQuit(void);

#endif
