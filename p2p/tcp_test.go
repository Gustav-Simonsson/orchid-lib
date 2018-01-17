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

package p2p

import (
	"io"
	"net"
	"os"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/log"
)

type TestDest struct {
	data   []byte
	Closed bool
}

func (td *TestDest) Read(p []byte) (n int, err error) {
	n = copy(p, td.data)
	if len(p) >= len(td.data) {
		td.data = []byte{}
	} else {
		td.data = append(td.data[:len(p)], td.data[len(p)+1:]...)
	}
	return n, err
}

func (td *TestDest) Write(p []byte) (n int, err error) {
	td.data = append(td.data, p...)
	return len(p), nil
}

func (td *TestDest) Close() error {
	td.data = []byte{}
	td.Closed = true
	return nil
}

func TestMain(m *testing.M) {
	// setup logger
	log.Root().SetHandler(log.MultiHandler(
		log.StreamHandler(os.Stderr, log.TerminalFormat(true)),
		log.LvlFilterHandler(
			log.LvlDebug,
			log.Must.FileHandler("errors.json", log.JsonFormat()))))

	os.Exit(m.Run())
}

func TestTCProxy(t *testing.T) {
	td := &TestDest{[]byte("foobar"), false}

	dstGen := func() (io.ReadWriteCloser, error) {
		return td, nil
	}

	proxy, err := NewTCPProxy(8080, dstGen)
	if err != nil {
		t.Fail()
	}

	go func() {
		err = proxy.ListenAndServe()
		if err != nil {
			t.Fail()
		}
	}()

	conn, err := net.Dial("tcp", "localhost:8080")
	if err != nil {
		t.Fail()
	}

	nw, err := conn.Write([]byte("42"))
	if err != nil {
		t.Fatalf("conn.Write err: %v", err)
	}
	if nw != 2 {
		t.Fatalf("unexpected conn.Write n: %v", nw)
	}

	buf := make([]byte, 3)
	nr, err := conn.Read(buf)
	if err != nil {
		t.Fatalf("conn.Read err: %v", err)
	}
	if nr != 3 {
		t.Fatalf("unexpected conn.Read n: %v", nr)
	}

	if string(buf) != "foo" {
		t.Fatalf("unexpected conn.Read buf: %v", buf)
	}

	buf = make([]byte, 2)
	nr, err = conn.Read(buf)
	if err != nil {
		t.Fatalf("conn.Read err: %v", err)
	}
	if nr != 2 {
		t.Fatalf("unexpected conn.Read n: %v", nr)
	}

	if string(buf) != "ba" {
		t.Fatalf("unexpected conn.Read buf: %v", buf)
	}

	err = conn.Close()
	if err != nil {
		t.Fatalf("td.Close err: %v", err)
	}

	err = td.Close()
	if err != nil {
		t.Fatalf("td.Close err: %v", err)
	}
	if td.Closed == false {
		t.Fatalf("td.Closed %v", td.Closed)
	}

	time.Sleep(100 * time.Millisecond)
}
