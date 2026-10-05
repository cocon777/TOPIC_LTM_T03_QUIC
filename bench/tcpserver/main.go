package main

import (
	"log"
	"net"

	"quic-demo/bench"
)

const addr = "localhost:5002"

func main() {
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("TCP bench server on", addr)
	for {
		conn, err := ln.Accept()
		if err != nil {
			continue
		}
		go func(c net.Conn) {
			defer c.Close()
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
