package obfu

import (
	"bytes"
	"fmt"
	"io"
	"net"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func genBytes(n int) []byte {
	buf := make([]byte, n)
	for j := 0; j < n; j++ {
		buf[j] = byte(j)
	}
	return buf
}

func TestStreamConn(tt *testing.T) {
	t := require.New(tt)
	a, b := net.Pipe()
	a = RandXOR.ObfuscateStreamConn(a)
	b = RandXOR.ObfuscateStreamConn(b)
	all := make([]byte, 0, 100000)

	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		for i := 0; i < 2048; i++ {
			buf := genBytes(i)
			all = append(all, buf...)
			n, err := a.Write(buf)
			t.NoError(err)
			t.Equal(i, n)
		}
		t.NoError(a.Close())
		wg.Done()
	}()
	go func() {
		buf := make([]byte, 100)
		cur := 0
		for {
			n, err := b.Read(buf)
			switch err {
			case io.EOF:
			default:
				t.NoError(err)
			}
			if n > 0 {
				t.Equal(all[cur:cur+n], buf[:n])
				cur += n
			}
			if err == io.EOF {
				break
			}
		}
		wg.Done()
	}()
	wg.Wait()
}

func TestDatagramConn(tt *testing.T) {
	t := require.New(tt)
	l, err := net.ListenPacket("udp4", "")
	t.NoError(err)
	l = RandXOR.ObfuscatePacketConn(l)

	var wg sync.WaitGroup
	wg.Add(1)

	start := 0
	totalSize := 65536 - ivSize - 20 - 8
	step := 25

	go func() {
		conn, err := net.Dial("udp4", l.LocalAddr().String())
		t.NoError(err)
		conn = RandXOR.ObfuscateDatagramConn(conn)

		buf := make([]byte, 65536)
		for i := start; i < totalSize; i += step {
			data := genBytes(i)
			nw, err := conn.Write(data)
			if err != nil {
				fmt.Println(len(data))
			}
			t.NoError(err)
			t.Equal(len(data), nw)
			n, err := conn.Read(buf)
			t.NoError(err)
			bytes.Equal(data, buf[:n])
			time.Sleep(time.Microsecond)
		}
		wg.Done()
	}()

	buf := make([]byte, 65536)
	for i := start; i < totalSize; i += step {
		n, addr, err := l.ReadFrom(buf)
		t.NoError(err)
		t.True(strings.HasPrefix(addr.String(), "127.0.0.1:"))
		t.Equal(genBytes(i), buf[:n])
		nw, err := l.WriteTo(buf[:n], addr)
		t.NoError(err)
		t.Equal(nw, n)
	}
	wg.Wait()
}
