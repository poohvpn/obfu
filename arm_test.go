package obfu

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func dupInfXorBytes(iv, data []byte, offset int) ([]byte, []byte, int) {
	iv, data = duplicate(iv), duplicate(data)
	offset = infXorBytes(iv, data, offset)
	return iv, data, offset
}

func TestInfXorBytes(tt *testing.T) {
	t := assert.New(tt)

	iv := []byte{0x12, 0x34, 0x56, 0x78}
	data := []byte{
		0x04, 0x00, 0x00, 0x00,
		0x11, 0x11, 0x11, 0x11,
	}
	offset := 0

	obfuIV, obfuData, obfuOffset := dupInfXorBytes(iv, data, offset)
	t.Equal(obfuIV, []byte{
		0x3b, 0x9a, 0x64, 0x38,
	})
	t.Equal([]byte{
		0x8a, 0x57, 0x9a, 0x48,
		0x2a, 0x8b, 0x75, 0x29,
	}, obfuData)
	t.Zero(obfuOffset)

	deobfuIV, deobfuData, deobfuOffset := dupInfXorBytes(iv, obfuData, obfuOffset)
	t.Equal(deobfuIV, []byte{
		0x3b, 0x9a, 0x64, 0x38,
	})
	t.Equal(data, deobfuData)
	t.Zero(deobfuOffset)
}
