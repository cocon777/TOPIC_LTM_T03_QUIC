package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"strings"
	"sync"
)

const addr = "localhost:4243"

func main() {
	messages := []string{"Message A", "Message B", "Message C", "Message D"}
	var wg sync.WaitGroup
	for _, msg := range messages {
		wg.Add(1)
		go func(msg string) {
			defer wg.Done()
			// TCP không có stream: mỗi message phải mở 1 connection riêng
			conn, err := net.Dial("tcp", addr)
			if err != nil {
				log.Println("dial:", err)
				return
			}
			defer conn.Close()
			fmt.Fprintf(conn, "%s\n", msg)
			reply, err := bufio.NewReader(conn).ReadString('\n')
			if err != nil {
				log.Println("read:", err)
				return
			}
			fmt.Printf("[conn %v] sent %q, reply: %s\n", conn.LocalAddr(), msg, strings.TrimSpace(reply))
		}(msg)
	}
	wg.Wait()
}
