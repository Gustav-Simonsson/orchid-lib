/*  orchid-lib  golang packages for the Orchid protocol.
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

/* In Orchid source nodes, we want a web browser connecting to a SOCKS5 proxy
   that abstracts the Orchid proxying of data - the forwarding of all data
   into the webRTC DataChannel of the connected relay node.
   To simplify and generalize this proxying, we model this as a generic
   TCP streamer that can be presented as a SOCKS5 proxy, even though the SOCKS5
   protocol handling actually takes place at the exit node.

   Example: source node (s), two relays (r1, r2), exit (e) and website(w):

   Request:  s -> r1 -> r2 -> e -> w
   Response: w -> e -> r2 -> r1 -> s

   If we ignore other Orchid protocol details such as connection setup,
   payment tickets, commands tags etc, we can view the request flow as:

   Web Browser (s) -> SOCKS5 (s) -> webRTC DC (s) ->
   webRTC DC (r1) ->
   webRTC DC (r2) ->
   webRTC DC (e) -> SOCKS5(e) ->
   HTTP Req (w)

   And the response flow:

   HTTP Resp (w) ->
   SOCKS5 (e) -> webRTC DC (e) ->
   webRTC DC (r2) ->
   webRTC DC (r1) ->
   webRTC DC (s) -> SOCKS5 (s) -> Web Browser (s)

   Note that the SOCKS5 endpoint at the source is locally just a TCP listener,
   and the SOCKS5 protocol handling takes place at the exit node.

*/

package p2p

import (
	"io"
	"net"
	"strconv"

	"github.com/ethereum/go-ethereum/log"
)

type TCPProxy struct {
	Host   string
	DstGen func() (io.ReadWriteCloser, error)
}

func NewTCPProxy(port int, dstGen func() (io.ReadWriteCloser, error)) (*TCPProxy, error) {
	host := "127.0.0.1:" + strconv.Itoa(port)
	ts := &TCPProxy{
		host,
		dstGen,
	}
	return ts, nil
}

func (ts *TCPProxy) ListenAndServe() error {
	l, err := net.Listen("tcp", ts.Host)
	if err != nil {
		return err
	}

	for {
		conn, err := l.Accept()
		if err != nil {
			return err
		}
		dst, err := ts.DstGen()
		if err != nil {
			return err
		}
		go func(c net.Conn, d io.ReadWriteCloser) {
			log.Debug("ServeConn (source)")
			ServeConn(c, d)
		}(conn, dst)
	}
	return nil
}

func ServeConn(src net.Conn, dst io.ReadWriteCloser) {
	srcDone := make(chan struct{}, 1)
	dstDone := make(chan struct{}, 1)

	buf0 := make([]byte, transferBufSize)
	buf1 := make([]byte, transferBufSize)
	// TODO: add timeout channels
	go copyBuffer(src, dst, buf0, srcDone)
	go copyBuffer(dst, src, buf1, dstDone)

	closeConns := func() {
		// TODO: consider setting (needs cast to net.TCPConn):
		// src.SetLinger(0)
		err := src.Close()
		if err != nil {
			log.Error("src.Close", "err", err)
		}
		err = dst.Close()
		if err != nil {
			log.Error("dst.Close", "err", err)
		}

	}

	select {
	case <-srcDone:
		closeConns()
	case <-dstDone:
		closeConns()
	}

	select {
	case <-srcDone:
	case <-dstDone:
	}

	return
}

func copyBuffer(dst io.Writer, src io.Reader, buf []byte, done chan struct{}) {
	n, err := io.CopyBuffer(dst, src, buf)
	done <- struct{}{}

	if err == nil {
		log.Debug("io.CopyBuffer closed with no error", "streamed", n)
	} else if err == io.EOF {
		log.Debug("io.CopyBuffer closed with EOF", "streamed", n)
	} else { // TODO: handle other specific io errors
		log.Info("io.CopyBuffer", "streamed", n, "err", err)
	}
}
