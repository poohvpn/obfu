package obfu

import "net"

type Obfuscator interface {
	ObfuscatePacketConn(conn net.PacketConn) net.PacketConn
	ObfuscateStreamConn(conn net.Conn) net.Conn
	ObfuscateDatagramConn(conn net.Conn) net.Conn
}
