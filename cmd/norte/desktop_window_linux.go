//go:build !headless && linux

package main

/*
#cgo pkg-config: gtk+-3.0
#include <gtk/gtk.h>
*/
import "C"

import (
	webview "github.com/webview/webview_go"
)

func maximizeDesktopWindow(w webview.WebView) {
	if ptr := w.Window(); ptr != nil {
		C.gtk_window_maximize((*C.GtkWindow)(ptr))
	}
}
