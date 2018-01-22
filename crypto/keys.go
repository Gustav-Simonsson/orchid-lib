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
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"

	crand "crypto/rand"

	nacl "github.com/kevinburke/nacl"
	naclbox "github.com/kevinburke/nacl/box"
)

type NodeKey struct {
	Pub  nacl.Key
	Priv nacl.Key
}

type nodeKeyJSON struct {
	Pub  string `json:"pub"`
	Priv string `json:"priv"`
}

func NewNodeKey() (*NodeKey, error) {
	pub, priv, err := naclbox.GenerateKey(crand.Reader)
	if err != nil {
		return nil, err
	}
	return &NodeKey{pub, priv}, nil
}

func (k *NodeKey) MarshalJSON() (j []byte, err error) {
	pub32 := [32]byte(*k.Pub)
	priv32 := [32]byte(*k.Priv)
	pubBytes := pub32[:]
	privBytes := priv32[:]

	jStruct := nodeKeyJSON{
		hex.EncodeToString(pubBytes),
		hex.EncodeToString(privBytes),
	}
	j, err = json.Marshal(jStruct)
	return j, err
}

func (k *NodeKey) UnmarshalJSON(j []byte) (err error) {
	kJSON := new(nodeKeyJSON)
	err = json.Unmarshal(j, &kJSON)
	if err != nil {
		return err
	}

	pub, err := nacl.Load(kJSON.Pub)
	priv, err := nacl.Load(kJSON.Priv)
	if err != nil {
		return err
	}
	k.Pub = pub
	k.Priv = priv

	return nil
}

func (k *NodeKey) PubBytes() []byte {
	pub32 := [32]byte(*k.Pub)
	return pub32[:]
}

// URL safe base64: '+' and '/' are replaced and '=' is omitted
// See https://git.saurik.com/schoinion.git/blob/HEAD:/src/index.ts#l58
// and https://stackoverflow.com/questions/26353710/how-to-achieve-base64-url-safe-encoding-in-c
func (k *NodeKey) URLBase64() string {
	return NACLKeyToURLBase64(k.Pub)
}

func NACLKeyToURLBase64(key nacl.Key) string {
	b := [32]byte(*key)
	s0 := base64.StdEncoding.EncodeToString(b[:])
	s1 := strings.Replace(s0, "+", "-", -1)
	s2 := strings.Replace(s1, "/", "_", -1)
	return strings.Replace(s2, "=", "", -1)
}

func URLBase64ToNACLKey(s0 string) (nacl.Key, error) {
	s1 := strings.Replace(s0, "-", "+", -1)
	s2 := strings.Replace(s1, "_", "/", -1)
	s3 := s2
	if len(s2)%2 != 0 {
		s3 = s2 + "="
	}
	b, err := base64.StdEncoding.DecodeString(s3)
	if err != nil {
		return nil, err
	}
	if len(b) != nacl.KeySize {
		return nil, fmt.Errorf("key len mismatch, have: %d, expected: %d", len(b), nacl.KeySize)
	}

	key := new([nacl.KeySize]byte) // nacl.Key is *[nacl.KeySize]byte
	copy(key[:], b)
	return key, nil
}
