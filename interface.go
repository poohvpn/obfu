package obfu

import (
	"math/rand"
	"net"
	"time"
)

var globalRand = rand.New(newPcgSource(time.Now().UnixNano()))

type Obfuscator interface {
	ObfuscatePacketConn(conn net.PacketConn) net.PacketConn
	ObfuscateStreamConn(conn net.Conn) net.Conn
	ObfuscateDatagramConn(conn net.Conn) net.Conn
}
