package obfu

import "net"

type Nop struct{}

var _ Obfuscator = Nop{}

func (n Nop) ObfuscatePacketConn(conn net.PacketConn) net.PacketConn {
	return conn
}

func (n Nop) ObfuscateDatagramConn(conn net.Conn) net.Conn {
	return conn
}

func (n Nop) ObfuscateStreamConn(conn net.Conn) net.Conn {
	return conn
}
