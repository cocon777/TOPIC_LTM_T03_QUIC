package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
)

const addr = "localhost:4243"

func main() {
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("TCP server listening on", addr)

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Println("accept error:", err)
			continue
		}
		log.Println("new connection from", conn.RemoteAddr())
		go handle(conn)
	}
}

func handle(conn net.Conn) {
	defer conn.Close()
	msg, err := bufio.NewReader(conn).ReadString('\n')
	if err != nil {
		log.Println("read error:", err)
		return
	}
	log.Printf("%v received: %q", conn.RemoteAddr(), msg)
	fmt.Fprintf(conn, "echo: %s", msg)
}
