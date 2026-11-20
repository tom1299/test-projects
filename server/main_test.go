package main

import (
	"bufio"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net"
	"testing"
	"time"
)

func TestServeUDP(t *testing.T) {
	listener, err := net.ListenPacket("udp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	done := make(chan struct{})
	go func() {
		defer close(done)
		serve(listener)
	}()
	defer func() {
		listener.Close()
		<-done
	}()

	client, err := net.Dial("udp", listener.LocalAddr().String())
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()

	var previous string
	for range 2 {
		if err := client.SetDeadline(time.Now().Add(time.Second)); err != nil {
			t.Fatal(err)
		}
		if _, err := client.Write([]byte("request")); err != nil {
			t.Fatal(err)
		}
		buffer := make([]byte, 1024)
		n, err := client.Read(buffer)
		if err != nil {
			t.Fatal(err)
		}
		auth := checkResponse(t, buffer[:n])
		if auth == previous {
			t.Fatal("auth values are identical")
		}
		previous = auth
	}
}

func TestServeTCP(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	done := make(chan error, 1)
	go func() { done <- serveTCP(listener) }()
	defer func() {
		listener.Close()
		if err := <-done; !errors.Is(err, net.ErrClosed) {
			t.Errorf("serveTCP returned %v, want closed listener error", err)
		}
	}()

	client, err := net.Dial("tcp", listener.Addr().String())
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()
	if err := client.SetDeadline(time.Now().Add(time.Second)); err != nil {
		t.Fatal(err)
	}

	reader := bufio.NewReader(client)
	var previous string
	for range 2 {
		if _, err := client.Write([]byte("request\n")); err != nil {
			t.Fatal(err)
		}
		line, err := reader.ReadBytes('\n')
		if err != nil {
			t.Fatal(err)
		}
		auth := checkResponse(t, line)
		if auth == previous {
			t.Fatal("auth values are identical")
		}
		previous = auth
	}
}

func checkResponse(t *testing.T, data []byte) string {
	t.Helper()
	var response struct {
		Auth string `json:"auth"`
	}
	if err := json.Unmarshal(data, &response); err != nil {
		t.Fatal(err)
	}
	auth, err := base64.StdEncoding.DecodeString(response.Auth)
	if err != nil || len(auth) != 32 {
		t.Fatalf("invalid auth value %q: %v", response.Auth, err)
	}
	return response.Auth
}
