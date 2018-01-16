/*  orchid-lib - golang packages for the Orchid protocol
    Copyright (C) 2018 Gustav Simonsson

    This file is part of orchid-lib.

    orchid-lib is free software: you can redistribute it and/or modify
    it under the terms of the GNU General Public License as published by
    the Free Software Foundation, either version 3 of the License, or
    (at your option) any later version.

    orchid-lib is distributed in the hope that it will be useful,
    but WITHOUT ANY WARRANTY; without even the implied warranty of
    MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
    GNU General Public License for more details.

    You should have received a copy of the GNU General Public License
    along with orchid-lib.  If not, see <http://www.gnu.org/licenses/>.
*/

package test

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/Gustav-Simonsson/orchid-lib/node"
	"github.com/Gustav-Simonsson/socks"
	"github.com/ethereum/go-ethereum/log"
)

func TestMain(m *testing.M) {
	// setup logger
	log.Root().SetHandler(log.MultiHandler(
		log.StreamHandler(os.Stderr, log.TerminalFormat(true)),
		log.LvlFilterHandler(
			log.LvlDebug,
			log.Must.FileHandler("errors.json", log.JsonFormat()))))

	os.Exit(m.Run())

}

func TestNode(t *testing.T) {
	// Setup simple test source & exit 	omain.SimpleSource()
	go node.SimpleExit()
	go node.SimpleSource()
	time.Sleep(400 * time.Millisecond)
	log.Debug("Node Test after node setup")

	// setup test HTTP server to act as external website
	http.HandleFunc("/orchid-node-test/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "test resp %v", 1)
	})
	go http.ListenAndServe(":3300", nil)

	for i := 0; i < 4; i++ {
		go func(t *testing.T, i int) {
			log.Debug("TEST FUNC A", "i", i)
			// Configure SOCKS5 Dialer to proxy the test HTTP requests through
			dialSocksProxy := socks.DialSocksProxy(socks.SOCKS5, "127.0.0.1:"+strconv.Itoa(node.SourceTCPPort))

			tr := &http.Transport{Dial: dialSocksProxy}
			httpClient := &http.Client{Transport: tr}

			// TODO: verify req, then multiple concurrent ones
			resp, err := httpClient.Get("http://127.0.0.1:3300/orchid-node-test/")
			log.Debug("TEST FUNC B", "i", i)
			if err != nil {
				t.Fatal(err)
			}
			defer resp.Body.Close()
			if resp.StatusCode != http.StatusOK {
				t.Fatal(resp.StatusCode)
			}
			buf, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				t.Fatal(err)
			}
			if string(buf) != "test resp 1" {
				t.Fatal("buf mismatch, got: ", string(buf))
			}

			tr.CloseIdleConnections()
		}(t, i)
	}

	time.Sleep(1200 * time.Millisecond)
}
