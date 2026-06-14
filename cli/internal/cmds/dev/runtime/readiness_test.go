package runtime

import (
	"context"
	"net"
	"testing"
	"time"
)

func TestPortReadyProbeReturnsWhenPortAcceptsConnections(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()

	done := make(chan struct{})
	go func() {
		conn, err := ln.Accept()
		if err == nil {
			_ = conn.Close()
		}
		close(done)
	}()

	port := ln.Addr().(*net.TCPAddr).Port
	probe := PortReadyProbe(port, 10*time.Millisecond, time.Second)
	if err := probe(context.Background()); err != nil {
		t.Fatalf("expected ready port, got %v", err)
	}
	<-done
}

func TestPortReadyProbeHonorsContext(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()

	probe := PortReadyProbe(1, 5*time.Millisecond, time.Second)
	if err := probe(ctx); err == nil {
		t.Fatal("expected context error")
	}
}
