// Code taken from: https://github.com/henrycg/prio/blob/master/utils/rand.go
/*

Copyright (c) 2016, Henry Corrigan-Gibbs

Permission to use, copy, modify, and/or distribute this software for any
purpose with or without fee is hereby granted, provided that the above
copyright notice and this permission notice appear in all copies.

THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES
WITH REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF
MERCHANTABILITY AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR
ANY SPECIAL, DIRECT, INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES
WHATSOEVER RESULTING FROM LOSS OF USE, DATA OR PROFITS, WHETHER IN AN
ACTION OF CONTRACT, NEGLIGENCE OR OTHER TORTIOUS ACTION, ARISING OUT OF
OR IN CONNECTION WITH THE USE OR PERFORMANCE OF THIS SOFTWARE.

*/

package pir

import (
	"bufio"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/binary"
	"io"
	"math/big"
	mrand "math/rand"
	"sync"
)

type PRGKey [aes.BlockSize]byte

var prgMutex sync.Mutex
var bufPrgReader *BufPRGReader

const bufSize = 8192

// Produce a random integer in Z_p where mod is the value p.
func RandInt(mod *big.Int) *big.Int {
	prgMutex.Lock()
	out := bufPrgReader.RandInt(mod)
	prgMutex.Unlock()
	return out
}

func MathRand() *mrand.Rand {
	return mrand.New(bufPrgReader)
}

// We use the AES-CTR to generate pseudo-random  numbers using a
// stream cipher. Go's native rand.Reader is extremely slow because
// it makes tons of system calls to generate a small number of
// pseudo-random bytes.
//
// We pay the overhead of using a sync.Mutex to synchronize calls
// to AES-CTR, but this is relatively cheap.
type PRGReader struct {
	Key    PRGKey
	stream cipher.Stream
}

type BufPRGReader struct {
	mrand.Source64
	Key    PRGKey
	stream *bufio.Reader
}

func NewPRG(key *PRGKey) *PRGReader {
	out := new(PRGReader)
	out.Key = *key

	var err error
	var iv [aes.BlockSize]byte

	block, err := aes.NewCipher(key[:])
	if err != nil {
		panic(err)
	}

	out.stream = cipher.NewCTR(block, iv[:])
	return out
}

func RandomPRGKey() *PRGKey {
	var key PRGKey
	_, err := io.ReadFull(rand.Reader, key[:])
	if err != nil {
		panic(err)
	}

	return &key
}

func RandomPRG() *PRGReader {
	return NewPRG(RandomPRGKey())
}

func (s *PRGReader) Read(p []byte) (int, error) {
	if len(p) < aes.BlockSize {
		var buf [aes.BlockSize]byte
		s.stream.XORKeyStream(buf[:], buf[:])
		copy(p[:], buf[:])
	} else {
		s.stream.XORKeyStream(p, p)
	}
	return len(p), nil
}

func NewBufPRG(prg *PRGReader) *BufPRGReader {
	out := new(BufPRGReader)
	out.Key = prg.Key
	out.stream = bufio.NewReaderSize(prg, bufSize)
	return out
}

func (b *BufPRGReader) RandInt(mod *big.Int) *big.Int {
	out, err := rand.Int(b.stream, mod)
	if err != nil {
		// TODO: Replace this with non-absurd error handling.
		panic("Catastrophic randomness failure!")
	}
	return out
}

func (b *BufPRGReader) Int63() int64 {
	uout := b.Uint64()
	uout = uout % (1 << 63)
	return int64(uout)
}

func (b *BufPRGReader) Uint64() uint64 {
	var buf [8]byte

	prgMutex.Lock()
	read := 0
	for read < 8 {
		n, err := b.stream.Read(buf[read:8])
		if err != nil {
			panic("Should never get here")
		}
		read += n
	}
	prgMutex.Unlock()

	return binary.LittleEndian.Uint64(buf[:])
}

func (b *BufPRGReader) Seed(int64) {
	panic("Should never call seed")
}

func init() {
	bufPrgReader = NewBufPRG(RandomPRG())
}
