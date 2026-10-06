package tlsutil

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
)

// sinh chứng chỉ tự ký và trả về cấu hình TLS cho server
func GenerateTLSConfig() *tls.Config {
	// tạo khóa riêng RSA 2048 bit
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		panic(err)
	}

	// mẫu chứng chỉ, chỉ khai báo số serial và tổ chức
	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject:      pkix.Name{Organization: []string{"QUIC Demo"}},
	}

	// ký chứng chỉ. Tham số cha = chính mẫu => chứng chỉ tự ký
	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &key.PublicKey, key)
	if err != nil {
		panic(err)
	}

	// mã hóa sang PEM rồi ghép thành tls.Certificate
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)})
	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER})
	tlsCert, err := tls.X509KeyPair(certPEM, keyPEM)
	if err != nil {
		panic(err)
	}

	// cấu hình trả về, NextProtos là ALPN (client phải khai báo giống hệt)
	return &tls.Config{
		Certificates: []tls.Certificate{tlsCert},
		NextProtos:   []string{"quic-demo"},
	}
}