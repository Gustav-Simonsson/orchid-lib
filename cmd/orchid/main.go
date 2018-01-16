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

package main

import (
	"os"
	"path/filepath"

	"github.com/Gustav-Simonsson/orchid-lib/node"
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
		err = node.SimpleSource()
	} else {
		err = node.SimpleExit()
	}
	if err != nil {
		log.Error("node exit:", "source", os.Args[1] == "source", "err", err)
	}

	os.Exit(1)
}
