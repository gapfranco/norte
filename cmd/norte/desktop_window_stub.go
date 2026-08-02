//go:build !headless && !linux && !windows && !darwin

package main

import webview "github.com/webview/webview_go"

func maximizeDesktopWindow(w webview.WebView) {}
