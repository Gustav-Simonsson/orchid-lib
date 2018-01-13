package util

import "runtime"

// TODO: refactor to use https://golang.org/pkg/go/build/#hdr-Build_Constraints
func GetChromePath() string {
	switch runtime.GOOS {
	case "darwin":
		return "/Applications/Google Chrome.app/Contents/MacOS/Google Chrome"
	case "win32":
		// windows 10, TODO: older versions
		return "C:\\Program Files (x86)\\Google\\Chrome\\Application\\chrome.exe"
	case "linux":
		return "/usr/bin/google-chrome"
	default:
		panic("unsupported platform")
	}
}
