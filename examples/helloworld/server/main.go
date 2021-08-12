package main

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"log"
	"math/big"

	qnet "github.com/dpogorzelski/grpc-quic"
	pb "github.com/dpogorzelski/grpc-quic/examples/helloworld/helloworld"
	"github.com/lucas-clemente/quic-go"
	"google.golang.org/grpc"
)

const (
	addr        = "localhost:4242"
	defaultName = "world"
)

// server is used to implement hello.GreeterServer.
type server struct {
	pb.UnimplementedGreeterServer
}

// SayHello implements hello.GreeterServer
func (s *server) SayHello(ctx context.Context, in *pb.HelloRequest) (*pb.HelloReply, error) {
	log.Printf("Received: %v", in.GetName())
	return &pb.HelloReply{Message: "Hello " + in.GetName()}, nil
}

func main() {
	err := echoQuicServer()
	if err != nil {
		log.Printf("failed to Echo QUIC Server. %s", err.Error())
		return
	}
}

func echoQuicServer() error {
	log.Println("starting echo quic server")

	tlsConf, err := generateTLSConfig()
	if err != nil {
		log.Printf("QuicServer: failed to generateTLSConfig. %s", err.Error())
		return err
	}

	ql, err := quic.ListenAddr(addr, tlsConf, nil)
	if err != nil {
		log.Printf("QuicServer: failed to ListenAddr. %s", err.Error())
		return err
	}
	listener := qnet.Listen(ql)

	s := grpc.NewServer()
	pb.RegisterGreeterServer(s, &server{})
	log.Printf("QuicServer: listening at %v", listener.Addr())

	if err := s.Serve(listener); err != nil {
		log.Printf("QuicServer: failed to serve. %v", err)
		return err
	}

	log.Println("stopping echo quic server")
	return nil
}

// Setup a bare-bones TLS config for the server
func generateTLSConfig() (*tls.Config, error) {
	publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, err
	}

	template := x509.Certificate{SerialNumber: big.NewInt(1)}
	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, publicKey, privateKey)
	if err != nil {
		return nil, err
	}
	bytes, err := x509.MarshalPKCS8PrivateKey(privateKey)
	if err != nil {
		return nil, err
	}

	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: bytes})
	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER})

	tlsCert, err := tls.X509KeyPair(certPEM, keyPEM)
	if err != nil {
		return nil, err
	}
	return &tls.Config{
		MinVersion:       tls.VersionTLS13,
		CurvePreferences: []tls.CurveID{tls.X25519},
		CipherSuites:     []uint16{tls.TLS_CHACHA20_POLY1305_SHA256},
		Certificates:     []tls.Certificate{tlsCert},
		NextProtos:       []string{"quic-echo-example"},
	}, nil
}
