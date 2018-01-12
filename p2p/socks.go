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
