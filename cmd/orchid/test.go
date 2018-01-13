package main

import (
	"errors"
	"io"
	"net"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"sync"
	"time"

	"github.com/Gustav-Simonsson/orchid-lib/p2p"
	"github.com/Gustav-Simonsson/orchid-lib/util"
	"github.com/ethereum/go-ethereum/log"
	homedir "github.com/mitchellh/go-homedir"
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
			dc, err := wPeer.NewDataChannel()
			if err != nil {
				log.Error("CreateDataChannel (TCP proxy callback)", "err", err)
				return nil, err
			}
			dcRWC := p2p.NewDCReadWriteCloser(dc)
			return dcRWC, nil
		})
	if err != nil {
		log.Error("p2p.NewTCPProxy", "err", err)
		return err
	}

	go proxy.ListenAndServe()
	time.Sleep(100 * time.Millisecond)

	home, err := homedir.Dir()
	if err != nil {
		panic(err)
	}
	orchidDir = filepath.Join(home, ".orchid")
	userChromeDir := filepath.Join(orchidDir, ".chrome")
	err = os.MkdirAll(userChromeDir, 0700)
	if err != nil {
		return err
	}

	chromeArgs := []string{
		"--no-first-run",
		"--user-data-dir=" + userChromeDir,
		"--proxy-server=socks5://127.0.0.1:3200",
		"--host-resolver-rules=MAP * ~NOTFOUND , EXCLUDE 127.0.0.1",
	}
	chromeBin := util.GetChromePath()

	log.Info("Source ready, launching chrome...")
	err = exec.Command(chromeBin, chromeArgs...).Run()
	if err != nil {
		log.Error("chrome", "err", err)
		return err
	}

	log.Info("chrome exited, stopping source node")
	return nil
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

	log.Info("Exit ready...")
	p2p.HTTPServer(ExitHTTPPort, func(b []byte) ([]byte, error) {
		exit.Mutex.Lock()

		// temp for testing. TODO: support multiple peers
		if exit.LocalPeer != nil {
			return nil, errors.New("already have source peer")
		}

		dcReady := make(chan *p2p.DCReadWriteCloser, 10)
		go func() {
			for {
				dcRWC := <-dcReady
				// stream (copyBuffer) from dcRWC to SOCKS5
				conn, err := net.Dial("tcp", "127.0.0.1:"+strconv.Itoa(ExitSOCKS5Port))
				if err != nil {
					log.Error("net.Dial (to SOCKS5 proxy)", "err", err)
					return
				}
				//_, _ = conn, dcRWC
				go p2p.ServeConn(conn, dcRWC)
				log.Info("p2p.NewDCReadWriteCloser (exit)", "ns", time.Now().UnixNano())
			}
		}()

		resp, peer, err := p2p.NewExit(b, dcReady)
		if err != nil {
			return nil, err
		}
		exit.LocalPeer = peer // this is ourself, not the remote peer
		exit.Mutex.Unlock()

		return resp, err

	})
	return nil
}
