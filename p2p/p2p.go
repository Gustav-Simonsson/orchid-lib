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

package p2p

const (
	// Default in io.Copy is 30 * 1024
	// On source nodes this is probably fine, on relays and exits
	// (especially if embedded devices) this can be important to tune
	// Used by TCPProxy and DCReadWriteCloser
	transferBufSize = 30 * 1024
)
