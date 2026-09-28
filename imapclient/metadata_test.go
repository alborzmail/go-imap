package imapclient_test

import (
	"bufio"
	"net"
	"testing"

	"github.com/emersion/go-imap/v2/imapclient"
)

func TestGetMetadata_options(t *testing.T) {
	clientConn, serverConn := net.Pipe()
	defer serverConn.Close()

	sent := make(chan string, 1)
	go func() {
		serverConn.Write([]byte("* OK [CAPABILITY IMAP4rev1 METADATA] ready\r\n"))
		line, _ := bufio.NewReader(serverConn).ReadString('\n')
		sent <- line
		serverConn.Write([]byte(`* METADATA "" (/private/vendor/a "1")` + "\r\n"))
		serverConn.Write([]byte("T1 OK GETMETADATA completed\r\n"))
	}()

	client := imapclient.New(clientConn, nil)
	defer client.Close()

	maxSize := uint32(1024)
	data, err := client.GetMetadata("", []string{"/private/vendor"}, &imapclient.GetMetadataOptions{
		MaxSize: &maxSize,
		Depth:   imapclient.GetMetadataDepthInfinity,
	}).Wait()
	if err != nil {
		t.Fatalf("GetMetadata() = %v", err)
	}

	want := `T1 GETMETADATA (MAXSIZE 1024 DEPTH infinity) "" ("/private/vendor")` + "\r\n"
	if got := <-sent; got != want {
		t.Errorf("sent %q, want %q", got, want)
	}
	if v := data.Entries["/private/vendor/a"]; v == nil || string(*v) != "1" {
		t.Errorf("GetMetadata() = %+v, want /private/vendor/a = 1", data)
	}
}
