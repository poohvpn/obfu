package obfu

import (
	"io"
	"net"
)

const ivSize = 4

var RandXOR Obfuscator = &randXOR{}

type randXOR struct{}

var _ Obfuscator = &randXOR{}

func (x *randXOR) ObfuscatePacketConn(conn net.PacketConn) net.PacketConn {
	return &packetConn{
		PacketConn: conn,
	}
}

func (x *randXOR) ObfuscateStreamConn(conn net.Conn) net.Conn {
	return &streamConn{
		Conn: conn,
		lIV:  genIV(),
	}
}

func (x *randXOR) ObfuscateDatagramConn(conn net.Conn) net.Conn {
	return &datagramConn{
		Conn: conn,
	}
}

type streamConn struct {
	net.Conn
	lIVdone   bool
	lIV       []byte
	lIVOffset int
	rIVdone   bool
	rIV       []byte
	rIVOffset int
}

func (c *streamConn) Read(b []byte) (int, error) {
	if !c.rIVdone {
		c.rIV = make([]byte, len(c.lIV))
		_, err := io.ReadFull(c.Conn, c.rIV)
		if err != nil {
			return 0, err
		}
		c.rIVdone = true
	}
	n, err := c.Conn.Read(b)
	if err != nil {
		return 0, err
	}
	c.rIVOffset = infXorBytes(c.rIV, b[:n], c.rIVOffset)
	return n, nil
}

func (c *streamConn) Write(b []byte) (int, error) {
	if !c.lIVdone {
		_, err := c.Conn.Write(c.lIV)
		if err != nil {
			return 0, err
		}
		c.lIVdone = true
	}
	data := duplicate(b)
	c.lIVOffset = infXorBytes(c.lIV, data, c.lIVOffset)
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
