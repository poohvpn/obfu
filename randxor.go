package obfu

import (
	"io"
	"net"
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
		Conn:    conn,
		localIV: genIV(),
	}
}

func (randXOR) ObfuscateDatagramConn(conn net.Conn) net.Conn {
	return &datagramConn{
		Conn: conn,
	}
}

type streamConn struct {
	net.Conn
	localIVDone    bool
	localIV        []byte
	localIVOffset  int
	remoteIVDone   bool
	remoteIV       []byte
	remoteIVOffset int
}

func (c *streamConn) Read(b []byte) (int, error) {
	if !c.remoteIVDone {
		c.remoteIV = make([]byte, len(c.localIV))
		_, err := io.ReadFull(c.Conn, c.remoteIV)
		if err != nil {
			return 0, err
		}
		c.remoteIVDone = true
	}
	n, err := c.Conn.Read(b)
	if err != nil {
		return 0, err
	}
	c.remoteIVOffset = infXorBytes(c.remoteIV, b[:n], c.remoteIVOffset)
	return n, nil
}

func (c *streamConn) Write(b []byte) (int, error) {
	if !c.localIVDone {
		_, err := c.Conn.Write(c.localIV)
		if err != nil {
			return 0, err
		}
		c.localIVDone = true
	}
	data := duplicate(b)
	c.localIVOffset = infXorBytes(c.localIV, data, c.localIVOffset)
	return c.Conn.Write(data)
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
		iv := b[:ivSize]
		data := duplicate(b[ivSize:n])
		infXorBytes(iv, data, 0)
		return copy(b, data), nil
	}
}

func (c *datagramConn) Write(p []byte) (int, error) {
	_, err := c.Conn.Write(pack(p))
	if err != nil {
		return 0, err
	}
	return len(p), nil
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
		iv := b[:ivSize]
		data := duplicate(b[ivSize:n])
		infXorBytes(iv, data, 0)
		return copy(b, data), addr, nil
	}
}

func (c *packetConn) WriteTo(p []byte, addr net.Addr) (int, error) {
	_, err := c.PacketConn.WriteTo(pack(p), addr)
	if err != nil {
		return 0, err
	}
	return len(p), nil
}
