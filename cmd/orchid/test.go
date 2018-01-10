package main

import (
	"errors"
	"net/url"
	"sync"
	"time"

	"github.com/Gustav-Simonsson/orchid-lib/p2p"
	"github.com/ethereum/go-ethereum/log"
	webrtc "github.com/keroserene/go-webrtc"
)

type SimpleExit struct {
	Mutex sync.Mutex
	// TODO: generalize to multiple peers
	Source *p2p.WebRTCPeer
}

func simpleSource() error {
	log.Info("Starting simple source node...")

	ref, err := url.Parse("http://localhost:3200")
	if err != nil {
		return err
	}

	wPeer, err := p2p.NewWebRTCPeer(ref)
	if err != nil {
		return err
	}

	dcReadWriter := p2p.NewDCReadWriteCloser(wPeer.DCs[0])

	for {
		_, err := dcReadWriter.Write([]byte("foo"))
		if err != nil {
			log.Error("source: test write", "err", err)
		}
		time.Sleep(time.Millisecond * 2000)
	}

	return nil
}

func simpleExit() error {
	log.Info("Starting simple exit node...")

	exit := SimpleExit{
		sync.Mutex{},
		nil}

	var dcReadWriter *p2p.DCReadWriteCloser

	p2p.HTTPServer(3200, func(b []byte) ([]byte, error) {
		defer exit.Mutex.Unlock()
		exit.Mutex.Lock()
		if exit.Source != nil {
			return nil, errors.New("already have source peer")
		} else {
			dcReady := make(chan *webrtc.DataChannel, 1)
			resp, peer, err := p2p.NewExit(b, dcReady)
			if err != nil {
				return nil, err
			}
			exit.Source = peer

			go func() {
				dc, ok := <-dcReady
				if !ok {
					log.Error("dc ready chan not ok")
				} else {
					dcReadWriter = p2p.NewDCReadWriteCloser(dc)
					for {
						buf := make([]byte, 1)
						_, err := dcReadWriter.Read(buf)
						if err != nil {
							log.Error("exit: test read:", "err", err)
						}
						time.Sleep(time.Millisecond * 3000)
					}
				}

			}()

			return resp, err
		}

	})

	return nil
}
