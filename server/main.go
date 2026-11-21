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

	go serveUDP(udp)
	serveTCP(tcp)
	return nil
}

func serveUDP(conn net.PacketConn) {
	buffer := make([]byte, 65507)
	for {
		_, addr, err := conn.ReadFrom(buffer)
		if err != nil {
			if errors.Is(err, net.ErrClosed) {
				return
			}
			log.Print("failed to read UDP request")
			continue
		}

		response, err := authResponse()
		if err != nil {
			log.Print("failed to generate auth response")
			continue
		}
		if _, err := conn.WriteTo(response, addr); err != nil {
			log.Print("failed to send UDP response")
		}
	}
}

func serveTCP(listener net.Listener) {
	for {
		conn, err := listener.Accept()
		if err != nil {
			if errors.Is(err, net.ErrClosed) {
				return
			}
			log.Print("failed to read TCP request")
			continue
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
			log.Print("failed to generate auth response")
			return
		}
		response = append(response, '\n')
		if _, err := conn.Write(response); err != nil {
			log.Print("failed to send TCP response")
			return
		}
	}
	if err := scanner.Err(); err != nil {
		log.Print("failed to read TCP request")
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
