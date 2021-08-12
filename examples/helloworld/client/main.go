package main

import (
	"context"
	"crypto/tls"
	"log"
	"time"

	qnet "github.com/dpogorzelski/grpc-quic"
	pb "github.com/dpogorzelski/grpc-quic/examples/helloworld/helloworld"
	"google.golang.org/grpc"
)

const (
	addr        = "localhost:4242"
	defaultName = "world"
)

func main() {

	if err := echoQuicClient(); err != nil {
		log.Printf("failed to QUIC Client. %s", err.Error())
		return
	}

}

func echoQuicClient() error {
	tlsConf := &tls.Config{
		MinVersion:         tls.VersionTLS13,
		CurvePreferences:   []tls.CurveID{tls.X25519},
		CipherSuites:       []uint16{tls.TLS_CHACHA20_POLY1305_SHA256},
		InsecureSkipVerify: true,
		NextProtos:         []string{"quic-echo-example"},
	}

	creds := qnet.NewCredentials(tlsConf)

	dialer := qnet.NewQuicDialer(tlsConf)
	grpcOpts := []grpc.DialOption{
		grpc.WithContextDialer(dialer),
		grpc.WithTransportCredentials(creds),
	}

	conn, err := grpc.Dial(addr, grpcOpts...)
	if err != nil {
		log.Printf("QuicClient: failed to grpc.Dial. %s", err.Error())
		return err
	}
	defer func(conn *grpc.ClientConn) {
		err := conn.Close()
		if err != nil {
			log.Printf("QuicClient: failed to close - grpc.Dial. %s", err.Error())
		}
	}(conn)

	c := pb.NewGreeterClient(conn)
	name := defaultName
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	r, err := c.SayHello(ctx, &pb.HelloRequest{Name: name})
	if err != nil {
		log.Printf("QuicClient: could not greet. %v", err)
		return err
	}
	log.Printf("QuicClient: Greeting=%s", r.GetMessage())

	return nil
}
