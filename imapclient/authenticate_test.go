package imapclient_test

import (
	"bufio"
	"net"
	"strings"
	"testing"

	"github.com/emersion/go-sasl"

	"github.com/emersion/go-imap/v2"
	"github.com/emersion/go-imap/v2/imapclient"
	"github.com/emersion/go-imap/v2/internal"
)

func TestClient_Authenticate(t *testing.T) {
	client, server := newClientServerPair(t, imap.ConnStateNotAuthenticated)
	defer client.Close()
	defer server.Close()

	saslClient := sasl.NewPlainClient("", testUsername, testPassword)
	if err := client.Authenticate(saslClient); err != nil {
		t.Fatalf("Authenticate() = %v", err)
	}

	if state := client.State(); state != imap.ConnStateAuthenticated {
		t.Errorf("State() = %v, want %v", state, imap.ConnStateAuthenticated)
	}
}

func TestClient_Authenticate_debugRedacted(t *testing.T) {
	secret := internal.EncodeSASL([]byte("\x00" + testUsername + "\x00" + testPassword))

	t.Run("initial response", func(t *testing.T) {
		conn, server := newMemClientServerPair(t)
		defer server.Close()

		var debug lockedBuffer
		client := imapclient.New(conn, &imapclient.Options{DebugWriter: &debug})
		defer client.Close()

		if err := client.Authenticate(sasl.NewPlainClient("", testUsername, testPassword)); err != nil {
			t.Fatalf("Authenticate() = %v", err)
		}
		trace := debug.String()
		if !strings.Contains(trace, "<redacted>") || strings.Contains(trace, secret) {
			t.Errorf("debug output does not redact the initial response:\n%v", trace)
		}
	})

	t.Run("continuation", func(t *testing.T) {
		clientConn, serverConn := net.Pipe()
		defer serverConn.Close()
		go func() {
			r := bufio.NewReader(serverConn)
			serverConn.Write([]byte("* OK [CAPABILITY IMAP4rev1 AUTH=PLAIN] ready\r\n"))
			r.ReadString('\n')
			serverConn.Write([]byte("+ \r\n"))
			r.ReadString('\n')
			serverConn.Write([]byte("T1 OK authenticated\r\n"))
		}()

		var debug lockedBuffer
		client := imapclient.New(clientConn, &imapclient.Options{DebugWriter: &debug})
		defer client.Close()

		if err := client.Authenticate(sasl.NewPlainClient("", testUsername, testPassword)); err != nil {
			t.Fatalf("Authenticate() = %v", err)
		}
		trace := debug.String()
		if !strings.Contains(trace, "<redacted>") || strings.Contains(trace, secret) {
			t.Errorf("debug output does not redact the response:\n%v", trace)
		}
	})
}
