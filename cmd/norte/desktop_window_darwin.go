//go:build !headless && darwin

package main

/*
#cgo CFLAGS: -x objective-c
#cgo LDFLAGS: -framework Cocoa
#import <Cocoa/Cocoa.h>

void maximizeWindow(void *window) {
    NSWindow *win = (__bridge NSWindow *)window;
    [win zoom:nil];
}
*/
import "C"

import (
	webview "github.com/webview/webview_go"
)

func maximizeDesktopWindow(w webview.WebView) {
	if ptr := w.Window(); ptr != nil {
		C.maximizeWindow(ptr)
	}
}
