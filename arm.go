package obfu

import (
	"encoding/binary"
	"math/rand"
	"time"

	"github.com/poohvpn/xor"
	"github.com/zeebo/xxh3"
)

var globalRand = rand.New(rand.NewSource(time.Now().UnixNano()))

func genIV() []byte {
	iv := make([]byte, ivSize)
	_, err := globalRand.Read(iv)
	if err != nil {
		panic(err)
	}
	return iv
}

// infXorBytes will modify iv and data
func infXorBytes(iv, data []byte, offset int) int {
	dataLen := len(data)
	ivLen := len(iv)
	for i := 0; i < dataLen; {
		if offset == 0 {
			binary.BigEndian.PutUint32(iv, uint32(xxh3.Hash(iv)))
		}
		n := xor.DstBytes(data[i:], iv[offset:], data[i:])
		offset = (offset + n) % ivLen
		i += n
	}
	return offset
}

func duplicate(bs []byte) []byte {
	if bs == nil {
		return nil
	}
	bsDup := make([]byte, len(bs))
	if len(bs) > 0 {
		copy(bsDup, bs)
	}
	return bsDup
}

func pack(data []byte) []byte {
	iv := genIV()
	packet := append(iv, data...)
	infXorBytes(packet[:len(iv)], packet[len(iv):], 0)
	copy(packet, iv)
	return packet
}
