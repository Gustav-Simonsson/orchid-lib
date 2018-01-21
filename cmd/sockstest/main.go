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

package main

import (
	"os"

	"github.com/Gustav-Simonsson/orchid-lib/p2p"
	"github.com/ethereum/go-ethereum/log"
)

func init() {
	log.Root().SetHandler(log.MultiHandler(
		log.StreamHandler(os.Stderr, log.TerminalFormat(true)),
		log.LvlFilterHandler(
			log.LvlDebug,
			log.Must.FileHandler("sockstest_errors.json", log.JsonFormat()))))

}

func main() {
	sp, err := p2p.NewSOCKSProxy()
	if err != nil {
		os.Exit(1)
	}

	err = sp.ListenAndServe(3206)
	if err != nil {
		os.Exit(1)
	} else {
		os.Exit(0)
	}

}
