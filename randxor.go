package obfu

import (
	"encoding/binary"
	"io"
	"math/rand"
	"net"

	"github.com/poohvpn/xor"
)

const ivSize = 4

var RandXOR Obfuscator = randXOR{}

type randXOR struct{}

func (randXOR) ObfuscatePacketConn(conn net.PacketConn) net.PacketConn {
	return &packetConn{
		PacketConn: conn,
	}
}

func (randXOR) ObfuscateStreamConn(conn net.Conn) net.Conn {
	return &streamConn{
		Conn: conn,
	}
}

func (randXOR) ObfuscateDatagramConn(conn net.Conn) net.Conn {
	return &datagramConn{
		Conn: conn,
	}
}

type streamConn struct {
	net.Conn
	localRng  *rand.Rand
	remoteRng *rand.Rand
}

func (c *streamConn) Read(b []byte) (int, error) {
	if c.remoteRng == nil {
		var iv [ivSize]byte
		n, err := c.Conn.Read(iv[:])
		if err != nil {
			return 0, err
		}
		if n != ivSize {
			return 0, io.ErrShortBuffer
		}
		c.remoteRng = rand.New(newPcgSource(int64(binary.BigEndian.Uint32(iv[:]))))
	}

	n, err := c.Conn.Read(b)
	if err != nil {
		return 0, err
	}

	rngBuf := make([]byte, n)
	c.remoteRng.Read(rngBuf)
	xor.DstBytes(b[:n], b[:n], rngBuf)

	return n, nil
}

func (c *streamConn) Write(b []byte) (int, error) {
	if c.localRng == nil {
		ivU32 := globalRand.Uint32()
		var iv [ivSize]byte
		binary.BigEndian.PutUint32(iv[:], ivU32)
		_, err := c.Conn.Write(iv[:])
		if err != nil {
			return 0, err
		}
		c.localRng = rand.New(newPcgSource(int64(ivU32)))
	}

	rngBuf := make([]byte, len(b))
	c.localRng.Read(rngBuf)
	xor.DstBytes(rngBuf, b, rngBuf)

	return c.Conn.Write(rngBuf)
}

type datagramConn struct {
	net.Conn
}

func (c *datagramConn) Read(b []byte) (int, error) {
	for {
		n, err := c.Conn.Read(b)
		if err != nil {
			return 0, err
		}
		if n < ivSize {
			continue
		}
		packetLen := n - ivSize
		rng := rand.New(newPcgSource(int64(binary.BigEndian.Uint32(b[packetLen:n]))))
		rngBuf := make([]byte, packetLen)
		rng.Read(rngBuf)
		return xor.DstBytes(b[:packetLen], b[:packetLen], rngBuf), nil
	}
}

func (c *datagramConn) Write(p []byte) (int, error) {
	pLen := len(p)

	ivU32 := globalRand.Uint32()
	rng := rand.New(newPcgSource(int64(ivU32)))
	rngBuf := make([]byte, ivSize+pLen)
	rng.Read(rngBuf[:pLen])
	binary.BigEndian.PutUint32(rngBuf[pLen:], ivU32)
	xor.DstBytes(rngBuf, p, rngBuf)

	_, err := c.Conn.Write(rngBuf)
	if err != nil {
		return 0, err
	}
	return pLen, nil
}

type packetConn struct {
	net.PacketConn
}

func (c *packetConn) ReadFrom(b []byte) (int, net.Addr, error) {
	for {
		n, addr, err := c.PacketConn.ReadFrom(b)
		if err != nil {
			return 0, nil, err
		}
		if n < ivSize {
			continue
		}
		packetLen := n - ivSize
		rng := rand.New(newPcgSource(int64(binary.BigEndian.Uint32(b[packetLen:n]))))
		rngBuf := make([]byte, packetLen)
		rng.Read(rngBuf)
		return xor.DstBytes(b[:packetLen], b[:packetLen], rngBuf), addr, nil
	}
}

func (c *packetConn) WriteTo(p []byte, addr net.Addr) (int, error) {
	pLen := len(p)

	ivU32 := globalRand.Uint32()
	rng := rand.New(newPcgSource(int64(ivU32)))
	rngBuf := make([]byte, ivSize+pLen)
	rng.Read(rngBuf[:pLen])
	binary.BigEndian.PutUint32(rngBuf[pLen:], ivU32)
	xor.DstBytes(rngBuf, p, rngBuf)

	_, err := c.PacketConn.WriteTo(rngBuf, addr)
	if err != nil {
		return 0, err
	}
	return pLen, nil
}
