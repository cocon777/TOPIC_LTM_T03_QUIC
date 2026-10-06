// server QUIC : Chấp nhận nhiều connection, mỗi connection chấp nhận nhiều stream.
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

// mở listener QUIC và chấp nhận connection
func main() {
	// lắng nghe UDP:4242 bằng QUIC, dùng TLS config tự sinh
	ln, err := quic.ListenAddr(addr, tlsutil.GenerateTLSConfig(), nil)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("QUIC server listening on", addr)

	// vòng lặp nhận connection, mỗi connection một goroutine
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

// nhận các stream mà client mở trên connection này
func handleConnection(conn *quic.Conn) {
	for {
		// AcceptStream chờ cho tới khi client mở stream mới
		stream, err := conn.AcceptStream(context.Background())
		if err != nil {
			log.Println("connection closed:", err)
			return
		}
		log.Printf("stream %d opened", stream.StreamID())
		go handleStream(stream)
	}
}

// đọc toàn bộ dữ liệu của 1 stream rồi trả về
func handleStream(stream *quic.Stream) {
	// ReadAll chỉ trả về khi client đóng chiều gửi
	buf, err := io.ReadAll(stream)
	if err != nil {
		log.Println("read error:", err)
		return
	}
	log.Printf("stream %d received: %s", stream.StreamID(), buf)
	fmt.Fprintf(stream, "echo: %s", buf)
	// đóng chiều gửi của server, báo cho client là đã trả lời xong
	stream.Close()
}
