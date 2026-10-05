package main

import (
	"context"
	"log"

	"quic-demo/bench"
	"quic-demo/internal/tlsutil"

	"github.com/quic-go/quic-go"
)

const addr = "localhost:5001"

func main() {
	ln, err := quic.ListenAddr(addr, tlsutil.GenerateTLSConfig(), nil)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("QUIC bench server on", addr)
	for {
		conn, err := ln.Accept(context.Background())
		if err != nil {
			continue
		}
		go func() {
			for {
				s, err := conn.AcceptStream(context.Background())
				if err != nil {
					return
				}
				go func() {
					msg, err := bench.ReadMessage(s)
					if err != nil {
						return
					}
					bench.WriteMessage(s, msg)
					s.Close()
				}()
			}
		}()
	}
}
