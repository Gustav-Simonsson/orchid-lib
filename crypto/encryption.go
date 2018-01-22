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

package crypto

import (
	nacl "github.com/kevinburke/nacl"
	naclbox "github.com/kevinburke/nacl/box"
)

const (
	/* https://github.com/kevinburke/nacl/blob/master/secretbox/secretbox.go#L31
	   secretbox.sliceForAppend (called from secretbox.Seal) performs
	   no allocation if the original slice has sufficient capacity.

	   TODO: re-design resource allocation for encryption /
	         consider forking nacl lib to optimize resource allocation.

	   TODO: analyze resource allocation and other steps performed here outside
	         the nacl lib for potential side channel attacks
	         (e.g. leaking key bits or other info from how we do resource alloc)
	*/
	sealPreAllocSize = 64
)

var ()

type Box struct {
	sharedKey nacl.Key
	nonce     nacl.Nonce
}

func NewBox(peerPub, priv nacl.Key) (*Box, error) {
	shared := naclbox.Precompute(peerPub, priv)
	nonce := nacl.NewNonce()
	return &Box{shared, nonce}, nil
}

func (b *Box) Seal(msg []byte) []byte {
	out := make([]byte, sealPreAllocSize)
	return naclbox.SealAfterPrecomputation(out, msg, b.nonce, b.sharedKey)
}

func (b *Box) Open(ciphertext []byte, nonce nacl.Nonce) ([]byte, bool) {
	out := make([]byte, sealPreAllocSize)
	return naclbox.OpenAfterPrecomputation(out, ciphertext, nonce, b.sharedKey)
}
