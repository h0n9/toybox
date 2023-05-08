package cluster

import (
	"bufio"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"math/big"
	"sync"

	"github.com/quic-go/quic-go"
)

type Manager struct {
	listenAddr string
}

func NewManager(listenAddr string) *Manager {
	return &Manager{
		listenAddr: listenAddr,
	}
}

func (manager *Manager) Run(ctx context.Context) error {
	tlsCfg := generateTLSConfig()
	listener, err := quic.ListenAddr(manager.listenAddr, tlsCfg, nil)
	if err != nil {
		return err
	}
	defer listener.Close()

	wg := sync.WaitGroup{}
	defer func() {
		fmt.Println("gotcha")
		wg.Wait()
	}()

	for {
		conn, err := listener.Accept(ctx)
		if err != nil {
			return err
		}
		defer conn.CloseWithError(0, "bye")

		stream, err := conn.AcceptStream(ctx)
		if err != nil {
			fmt.Println(err)
			continue
		}

		wg.Add(1)
		go func(ctx context.Context, stream quic.Stream) {
			defer wg.Done()
			handleStream(ctx, stream)
		}(ctx, stream)
	}
}

func handleStream(ctx context.Context, stream quic.Stream) {
	defer stream.Close()

	for {
		select {
		case <-ctx.Done():
			return
		default:
			reader := bufio.NewReader(stream)
			data, err := reader.ReadBytes('\n')
			if err != nil {
				fmt.Println(err)
				return
			}
			fmt.Printf("%s", data)
		}
	}
}

// TODO: replace or re-implement function
func generateTLSConfig() *tls.Config {
	key, err := rsa.GenerateKey(rand.Reader, 1024)
	if err != nil {
		panic(err)
	}
	template := x509.Certificate{SerialNumber: big.NewInt(1)}
	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &key.PublicKey, key)
	if err != nil {
		panic(err)
	}
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)})
	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER})

	tlsCert, err := tls.X509KeyPair(certPEM, keyPEM)
	if err != nil {
		panic(err)
	}
	return &tls.Config{
		Certificates: []tls.Certificate{tlsCert},
		NextProtos:   []string{"msg-lake"},
	}
}
