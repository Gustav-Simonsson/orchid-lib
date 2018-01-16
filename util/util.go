/*  orchid-lib - golang packages for the Orchid protocol.
    Copyright (C) 2018  Gustav Simonsson

    This file is part of orchid-lib.

    orchid-lib is free software: you can redistribute it and/or modify
    it under the terms of the GNU Affero General Public License as
    published by the Free Software Foundation, either version 3 of the
    License, or (at your option) any later version.

    orchid-lib is distributed in the hope that it will be useful,
    but WITHOUT ANY WARRANTY; without even the implied warranty of
    MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
    GNU Affero General Public License for more details.

    You should have received a copy of the GNU Affero General Public License
    along with this program.  If not, see <https://www.gnu.org/licenses/>.
*/

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
