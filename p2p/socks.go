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

import (
	"strconv"

	socks5 "github.com/armon/go-socks5"
)

/* See comments in tcp.go

   This runs at exit nodes, streaming from WebRTC DataChannel(s)
   to a local SOCKS5 proxy.
*/

type SOCKSProxy struct {
	//Mutex sync.Mutex
	srv *socks5.Server
}

func NewSOCKSProxy() (*SOCKSProxy, error) {
	conf := &socks5.Config{
	// TODO: verify conf
	}
	server, err := socks5.New(conf)
	if err != nil {
		return nil, err
	}

	proxy := SOCKSProxy{
		//sync.Mutex{},
		server,
	}
	return &proxy, nil

}

func (s *SOCKSProxy) ListenAndServe(port int) error {
	// Starts SOCKS5 proxy on localhost
	return s.srv.ListenAndServe("tcp", "127.0.0.1:"+strconv.Itoa(port))
}
