// công cụ đo, nhận tham số -proto -mode -n -warmup -size

package main

import (
	"context"
	"crypto/rand"
	"crypto/tls"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"sort"
	"time"

	"quic-demo/bench"

	"github.com/quic-go/quic-go"
)

const (
	quicAddr = "127.0.0.1:5001"
	tcpAddr  = "127.0.0.1:5002"
	tlsAddr  = "127.0.0.1:5003"
)

// clientTLS: cấu hình dùng chung cho QUIC và TCP+TLS
func clientTLS() *tls.Config {
	return &tls.Config{
		InsecureSkipVerify: true, // demo local
		NextProtos:         []string{"quic-demo"},
		MinVersion:         tls.VersionTLS13,
	}
}

// thiết lập một connection QUIC, giới hạn 5 giây
func dialQUIC() (*quic.Conn, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return quic.DialAddr(ctx, quicAddr, clientTLS(), nil)
}

// dialTCP: thiết lập TCP hoặc TCP+TLS tùy proto
func dialTCP(proto string) (net.Conn, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if proto == "tcptls" {
		d := tls.Dialer{Config: clientTLS()}
		return d.DialContext(ctx, "tcp", tlsAddr)
	}
	var d net.Dialer
	return d.DialContext(ctx, "tcp", tcpAddr)
}

// tryHandshake: một lần thiết lập kết nối rồi đóng - được đo ra kết quả, hạn 5 giây 
func tryHandshake(proto string) error {
	if proto == "quic" {
		conn, err := dialQUIC()
		if err != nil {
			return err
		}
		return conn.CloseWithError(0, "")
	}
	conn, err := dialTCP(proto)
	if err != nil {
		return err
	}
	return conn.Close()
}

// measure dùng cho mode=transfer: bỏ warmup, đo n lần, dừng nếu lỗi
func measure(n, warmup int, op func() error) ([]time.Duration, time.Duration) {
	for i := 0; i < warmup; i++ {
		if err := op(); err != nil {
			log.Fatal(err)
		}
	}
	d := make([]time.Duration, 0, n)
	start := time.Now()
	for i := 0; i < n; i++ {
		t0 := time.Now()
		if err := op(); err != nil {
			log.Fatal(err)
		}
		d = append(d, time.Since(t0))
	}
	return d, time.Since(start)
}

func main() {
	proto := flag.String("proto", "quic", "quic | tcp | tcptls")
	mode := flag.String("mode", "transfer", "handshake | transfer")
	n := flag.Int("n", 100, "số lần đo")
	warmup := flag.Int("warmup", 5, "số lần chạy khởi động (không tính)")
	size := flag.Int("size", 65536, "kích thước payload (bytes), mode=transfer")
	flag.Parse()

	if *proto != "quic" && *proto != "tcp" && *proto != "tcptls" {
		log.Fatal("proto phải là quic | tcp | tcptls")
	}

	payload := make([]byte, *size)
	rand.Read(payload)

	var durations []time.Duration
	var wall time.Duration
	fails := 0

	switch *mode {
	case "handshake":
		for i := 0; i < *warmup; i++ {
			tryHandshake(*proto) // khởi động, bỏ qua kết quả
		}
		for i := 0; i < *n; i++ {
			t0 := time.Now()
			err := tryHandshake(*proto)
			d := time.Since(t0)
			if err != nil {
				fails++       // thất bại: đếm riêng, không đưa vào kết quả avg
				log.Printf("lần %d thất bại sau %v: %v", i+1, d, err)
				continue
			}
			durations = append(durations, d)
			wall += d
		}

	case "transfer":
		if *proto == "quic" {
			conn, err := dialQUIC()
			if err != nil {
				log.Fatal(err)
			}
			defer conn.CloseWithError(0, "")
			durations, wall = measure(*n, *warmup, func() error {
				ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
				defer cancel()
				s, err := conn.OpenStreamSync(ctx)
				if err != nil {
					return fmt.Errorf("open stream: %w", err)
				}
				if err := bench.WriteMessage(s, payload); err != nil {
					return fmt.Errorf("write: %w", err)
				}
				if _, err := bench.ReadMessage(s); err != nil {
					return fmt.Errorf("read: %w", err)
				}
				if err := s.Close(); err != nil {
					return err
				}
				_, err = io.Copy(io.Discard, s)
				return err
			})
		} else {
			conn, err := dialTCP(*proto)
			if err != nil {
				log.Fatal(err)
			}
			defer conn.Close()
			durations, wall = measure(*n, *warmup, func() error {
				if err := bench.WriteMessage(conn, payload); err != nil {
					return err
				}
				_, err := bench.ReadMessage(conn)
				return err
			})
		}
	default:
		log.Fatal("mode phải là handshake | transfer")
	}

	fmt.Printf("proto=%s mode=%s n=%d warmup=%d", *proto, *mode, *n, *warmup)
	if *mode == "transfer" {
		fmt.Printf(" size=%dB", *size)
	}
	fmt.Println()
	if *mode == "handshake" {
		fmt.Printf("thành công=%d thất bại=%d\n", len(durations), fails)
	}
	if len(durations) == 0 {
		log.Fatal("không có lần đo nào thành công")
	}

	// tính và in kết quả
	avg := wall / time.Duration(len(durations))			// trung bình
	sorted := append([]time.Duration(nil), durations...)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i] < sorted[j] })		// sắp xếp để lấy min/median/max
	fmt.Printf("avg=%v | từng lần: min=%v median=%v max=%v\n", avg, sorted[0], sorted[len(sorted)/2], sorted[len(sorted)-1])
	if *mode == "transfer" {
		mib := float64(2**size) / avg.Seconds() / 1024 / 1024		// 2 : hai chiều
		fmt.Printf("throughput (2 chiều, theo avg) = %.2f MiB/s\n", mib)
	}
}
