package main

import (
	"bufio"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"log"
	"net"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	udp, err := net.ListenPacket("udp", ":6745")
	if err != nil {
		return fmt.Errorf("listen on UDP port 6745: %w", err)
	}
	defer udp.Close()

	tcp, err := net.Listen("tcp", ":6746")
	if err != nil {
		return fmt.Errorf("listen on TCP port 6746: %w", err)
	}
	defer tcp.Close()

	go serve(udp)
	return serveTCP(tcp)
}

func serve(conn net.PacketConn) {
	buffer := make([]byte, 65507)
	for {
		_, addr, err := conn.ReadFrom(buffer)
		if err != nil {
			if errors.Is(err, net.ErrClosed) {
				return
			}
			log.Printf("read UDP request: %v", err)
			continue
		}

		response, err := authResponse()
		if err != nil {
			log.Printf("generate auth response: %v", err)
			continue
		}
		if _, err := conn.WriteTo(response, addr); err != nil {
			log.Printf("send UDP response to %s: %v", addr, err)
		}
	}
}

func serveTCP(listener net.Listener) error {
	for {
		conn, err := listener.Accept()
		if err != nil {
			return fmt.Errorf("accept TCP connection: %w", err)
		}
		go handleTCP(conn)
	}
}

func handleTCP(conn net.Conn) {
	defer conn.Close()

	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		response, err := authResponse()
		if err != nil {
			log.Printf("generate auth response: %v", err)
			return
		}
		response = append(response, '\n')
		if _, err := conn.Write(response); err != nil {
			log.Printf("send TCP response to %s: %v", conn.RemoteAddr(), err)
			return
		}
	}
	if err := scanner.Err(); err != nil {
		log.Printf("read TCP request from %s: %v", conn.RemoteAddr(), err)
	}
}

func authResponse() ([]byte, error) {
	auth := make([]byte, 32)
	_, err := rand.Read(auth)
	if err != nil {
		return nil, err
	}
	return []byte(`{"auth":"` + base64.StdEncoding.EncodeToString(auth) + `"}`), nil
}
