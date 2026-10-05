package main

import (
	"context"
	"fmt"
	"io"
	"log"

	"quic-demo/internal/tlsutil"

	"github.com/quic-go/quic-go"
)

const addr = "localhost:4242"

func main() {
	ln, err := quic.ListenAddr(addr, tlsutil.GenerateTLSConfig(), nil)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("QUIC server listening on", addr)

	for {
		conn, err := ln.Accept(context.Background())
		if err != nil {
			log.Println("accept error:", err)
			continue
		}
		log.Println("new connection from", conn.RemoteAddr())
		go handleConnection(conn)
	}
}

func handleConnection(conn *quic.Conn) {
	for {
		stream, err := conn.AcceptStream(context.Background())
		if err != nil {
			log.Println("connection closed:", err)
			return
		}
		log.Printf("stream %d opened", stream.StreamID())
		go handleStream(stream)
	}
}

func handleStream(stream *quic.Stream) {
	buf, err := io.ReadAll(stream)
	if err != nil {
		log.Println("read error:", err)
		return
	}
	log.Printf("stream %d received: %s", stream.StreamID(), buf)
	fmt.Fprintf(stream, "echo: %s", buf)
	stream.Close()
}
