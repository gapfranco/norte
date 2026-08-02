//go:build !headless && windows

package main

import (
	"syscall"

	webview "github.com/webview/webview_go"
)

const swMaximize = 3

var (
	user32     = syscall.NewLazyDLL("user32.dll")
	showWindow = user32.NewProc("ShowWindow")
)

func maximizeDesktopWindow(w webview.WebView) {
	if ptr := w.Window(); ptr != nil {
		_, _, _ = showWindow.Call(uintptr(ptr), swMaximize)
	}
}
