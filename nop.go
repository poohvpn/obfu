package obfu

import "net"

var Nop Obfuscator = nop{}

type nop struct{}

func (nop) ObfuscatePacketConn(conn net.PacketConn) net.PacketConn {
	return conn
}

func (nop) ObfuscateDatagramConn(conn net.Conn) net.Conn {
	return conn
}

func (nop) ObfuscateStreamConn(conn net.Conn) net.Conn {
	return conn
}
