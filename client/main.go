// mở một connection, gửi 4 message song song, mỗi message một stream.
package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"log"
	"sync"
	"time"

	"github.com/quic-go/quic-go"
)

const addr = "localhost:4242"

func main() {
	// Cấu hình TLS phía client
	tlsConf := &tls.Config{
		InsecureSkipVerify: true, // demo local: cert self-signed
		NextProtos:         []string{"quic-demo"},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 1 handshake = 1 QUIC connection
	conn, err := quic.DialAddr(ctx, addr, tlsConf, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.CloseWithError(0, "done")
	fmt.Printf("1 connection: %v -> %v\n", conn.LocalAddr(), conn.RemoteAddr())

	messages := []string{"Message A", "Message B", "Message C", "Message D"}

	// mỗi message chạy trong goroutine riêng, WaitGroup chờ tất cả xong
	var wg sync.WaitGroup
	for _, msg := range messages {
		wg.Add(1)
		go func(msg string) {
			defer wg.Done()
			// Mỗi message đi trên 1 stream riêng, nhưng cùng connection
			stream, err := conn.OpenStreamSync(ctx)
			if err != nil {
				log.Println("open stream:", err)
				return
			}
			stream.Write([]byte(msg))
			stream.Close() // đóng chiều gửi để server đọc được

			// đọc phản hồi của riêng stream này
			buf, err := io.ReadAll(stream)
			if err != nil {
				log.Println("read:", err)
				return
			}
			fmt.Printf("[stream %d] sent %q, reply: %s\n", stream.StreamID(), msg, buf)
		}(msg)
	}
	wg.Wait()
}
