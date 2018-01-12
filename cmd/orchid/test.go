package main

import (
	"errors"
	"io"
	"net"
	"net/url"
	"strconv"
	"sync"

	webrtc "github.com/Gustav-Simonsson/go-webrtc"
	"github.com/Gustav-Simonsson/orchid-lib/p2p"
	"github.com/ethereum/go-ethereum/log"
)

const (
	SourceTCPPort  = 3200
	ExitHTTPPort   = 3201
	ExitSOCKS5Port = 3202
)

func simpleSource() error {
	log.Info("Starting simple source node...")

	ref, err := url.Parse("http://localhost:" + strconv.Itoa(ExitHTTPPort))
	if err != nil {
		return err
	}

	wPeer, err := p2p.NewWebRTCPeer(ref)
	if err != nil {
		return err
	}

	proxy, err := p2p.NewTCPProxy(SourceTCPPort,
		func() (io.ReadWriteCloser, error) {
			log.Debug("TCPProxy DstGen:")
			dc, err := wPeer.NewDataChannel()
			log.Debug("TCPProxy DstGen: after wPeer.NewDataChannel")
			if err != nil {
				log.Error("CreateDataChannel (TCP proxy callback)", "err", err)
				return nil, err
			}
			dcRWC := p2p.NewDCReadWriteCloser(dc)
			log.Debug("TCPProxy DstGen: after p2p.NewDCReadWriteCloser")
			return dcRWC, nil
		})
	if err != nil {
		log.Error("p2p.NewTCPProxy", "err", err)
		return err
	}

	return proxy.ListenAndServe()
}

type SimpleExit struct {
	Mutex sync.Mutex
	// TODO: generalize to multiple peers
	LocalPeer *p2p.WebRTCPeer
}

func simpleExit() error {
	log.Info("Starting simple exit node...")

	exit := SimpleExit{
		sync.Mutex{},
		nil}

	proxy, err := p2p.NewSOCKSProxy()
	if err != nil {
		return err
	}

	go func() {
		err := proxy.ListenAndServe(ExitSOCKS5Port)
		if err != nil {
			log.Error("SOCKS5 proxy ListenAndServe", "err", err)
		}
	}()

	p2p.HTTPServer(ExitHTTPPort, func(b []byte) ([]byte, error) {
		defer exit.Mutex.Unlock()
		exit.Mutex.Lock()

		// temp for testing. TODO: support multiple peers
		if exit.LocalPeer != nil {
			return nil, errors.New("already have source peer")
		} else {
			dcReady := make(chan *webrtc.DataChannel, 10)
			resp, peer, err := p2p.NewExit(b, dcReady)
			if err != nil {
				return nil, err
			}
			exit.LocalPeer = peer // this is ourself, not the remote peer

			go func() {
				for {
					dc, ok := <-dcReady
					if !ok {
						log.Error("dcReady chan not ok")
						return
					}
					dcRWC := p2p.NewDCReadWriteCloser(dc)

					// stream (copyBuffer) from dcRWC to SOCKS5
					conn, err := net.Dial("tcp", "127.0.0.1:"+strconv.Itoa(ExitSOCKS5Port))
					if err != nil {
						log.Error("net.Dial (to SOCKS5 proxy)", "err", err)
						return
					}
					_ = conn
					_ = dcRWC
					//p2p.ServeConn(conn, dcRWC)
				}
			}()

			return resp, err
		}
	})
	return nil
}
