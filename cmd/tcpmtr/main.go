package main

import (
	"flag"
	"fmt"
	"io"
	"log/slog"
	"net"
	"os"
	"sync/atomic"
	"time"
)

var (
	localAddr  = flag.String("l", ":8080", "Local address to listen on")
	remoteAddr = flag.String("r", "", "Remote address to forward to")
)

var connId uint64

type proxyConn struct {
	bytesIn, bytesOut *tcpmeter
	id uint64
}

func main() {
	flag.Parse()

	// Configure slog to use JSON output
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))

	if *remoteAddr == "" {
		slog.Error("Remote address (-r) is required")
		os.Exit(1)
	}

	listener, err := net.Listen("tcp", *localAddr)
	if err != nil {
		slog.Error("Failed to listen", "error", err)
		os.Exit(1)
	}
	defer listener.Close()

	slog.Info("Listening on", "localAddr", *localAddr, "forwardingTo", *remoteAddr)

	for {
		conn, err := listener.Accept()
		if err != nil {
			slog.Error("Accept error", "error", err)
			continue
		}

		go handleConnection(conn)
	}
}

// tcpmeter is a struct that wraps a net.Conn and records the total number of bytes read.
type tcpmeter struct {
	conn    net.Conn
	total   uint64
}

// Write implements the io.Writer interface for tcpsrc.
func (t *tcpmeter) Write(p []byte) (n int, err error) {
	n, err = t.conn.Write(p)
	atomic.AddUint64(&t.total, uint64(n))
	return
}

func handleConnection(src net.Conn) {
	defer src.Close()

	dst, err := net.Dial("tcp", *remoteAddr)
	if err != nil {
		slog.Error("Remote connection failed", "error", err)
		return
	}
	defer dst.Close()

	inMeter := &tcpmeter{conn: src}
	outMeter := &tcpmeter{conn: dst}

	
	pc := &proxyConn{
		bytesIn:  inMeter,
		bytesOut: outMeter,
		id: atomic.AddUint64(&connId, 1),
	}

	// Start bandwidth monitoring
	stopMonitor := make(chan struct{})
	go pc.monitorBandwidth(stopMonitor)

	// Bridge connections
	go func() {
		io.Copy(inMeter, dst)
	}()
	io.Copy(outMeter, src)

	// Stop monitoring when connection closes
	close(stopMonitor)
}

func (pc *proxyConn) monitorBandwidth(stop <-chan struct{}) {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	var prevIn, prevOut uint64
	var calculateAndLog = func() {
		currentIn := atomic.LoadUint64(&pc.bytesIn.total)
		currentOut := atomic.LoadUint64(&pc.bytesOut.total)

		inRate := currentIn - prevIn
		outRate := currentOut - prevOut

		slog.Info("traffic", "id", pc.id, "in", formatRate(inRate), "out", formatRate(outRate), "bytesIn", pc.bytesIn.total, "bytesOut", pc.bytesOut.total)

		prevIn = currentIn
		prevOut = currentOut
	}
	for {
		select {
		case <-ticker.C:
			calculateAndLog()
		case <-stop:
			calculateAndLog()
			return
		}
	}
}

func formatRate(bytes uint64) string {
	switch {
	case bytes >= 1<<20:
		return fmt.Sprintf("%.2f MB/s", float64(bytes)/(1<<20))
	case bytes >= 1<<10:
		return fmt.Sprintf("%.2f KB/s", float64(bytes)/(1<<10))
	default:
		return fmt.Sprintf("%d B/s", bytes)
	}
}