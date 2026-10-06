// server echo cho TCP+TLS (cổng 5003)
package main

import (
	"crypto/tls"
	"log"
	"net"

	"quic-demo/bench"
	"quic-demo/internal/tlsutil"
)

const addr = "localhost:5003"

func main() {
	ln, err := tls.Listen("tcp", addr, tlsutil.GenerateTLSConfig())
	if err != nil {
		log.Fatal(err)
	}
	log.Println("TCP+TLS bench server on", addr)
	for {
		conn, err := ln.Accept()
		if err != nil {
			continue
		}
		go func(c net.Conn) {
			defer c.Close()
			// một connection phục vụ NHIỀU request liên tiếp
			for {
				msg, err := bench.ReadMessage(c)
				if err != nil {
					return
				}
				bench.WriteMessage(c, msg)
			}
		}(conn)
	}
}
