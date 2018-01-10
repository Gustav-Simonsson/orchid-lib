package main

import (
	"os"
	"path/filepath"

	"github.com/ethereum/go-ethereum/log"
	homedir "github.com/mitchellh/go-homedir"
)

const ()

var (
	orchidDir string
)

func init() {
	log.Root().SetHandler(log.MultiHandler(
		log.StreamHandler(os.Stderr, log.TerminalFormat(true)),
		log.LvlFilterHandler(
			log.LvlDebug,
			log.Must.FileHandler("orchid_errors.json", log.JsonFormat()))))

	home, err := homedir.Dir()
	if err != nil {
		panic(err)
	}
	orchidDir = filepath.Join(home, ".orchid")
	err = os.MkdirAll(orchidDir, 0700)
	if err != nil {
		panic(err)
	}
}

func main() {
	// TODO: replace with urfave/cli app
	// simple session with one source and one exit node, launch as either:
	if len(os.Args) != 2 || (os.Args[1] != "source" && os.Args[1] != "exit") {
		log.Error("dev testing: run as 'orchid source' or 'orchid exit'")
	}

	var err error
	if os.Args[1] == "source" {
		err = simpleSource()
	} else {
		err = simpleExit()
	}
	if err != nil {
		log.Error("node exit:", "source", os.Args[1] == "source", "err", err)
	}

	os.Exit(1)
}
