package http

import (
	"errors"
	"math/rand"
	"net"
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/pierrre/imageserver"
	imageserver_provider "github.com/pierrre/imageserver/provider"
	"github.com/pierrre/imageserver/testdata"
)

var (
	testSourceFileName = testdata.SmallFileName
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func TestInterface(t *testing.T) {
	var _ imageserver_provider.Provider = &Provider{}
}

func TestGet(t *testing.T) {
	listener := createTestHTTPServer(t)
	defer listener.Close()

	source := createTestURL(listener)

	provider := &Provider{}

	image, err := provider.Get(source, imageserver.Params{})
	if err != nil {
		t.Fatal(err)
	}
	if image == nil {
		t.Fatal("no image")
	}
	if len(image.Data) == 0 {
		t.Fatal("no data")
	}
	if image.Format == "" {
		t.Fatal("no format")
	}
}

func TestGetErrorNotFound(t *testing.T) {
	listener := createTestHTTPServer(t)
	defer listener.Close()

	source := createTestURL(listener)
	source.Path += "foobar"

	provider := &Provider{}

	_, err := provider.Get(source, imageserver.Params{})
	if err == nil {
		t.Fatal("no error")
	}
}

func TestGetErrorInvalidUrl(t *testing.T) {
	source := "foobar"

	provider := &Provider{}

	_, err := provider.Get(source, imageserver.Params{})
	if err == nil {
		t.Fatal("no error")
	}
}

func TestGetErrorInvalidUrlScheme(t *testing.T) {
	source := "custom://foobar"

	provider := &Provider{}

	_, err := provider.Get(source, imageserver.Params{})
	if err == nil {
		t.Fatal("no error")
	}
}

func TestGetErrorRequest(t *testing.T) {
	source := "http://localhost:123456"

	provider := &Provider{}

	_, err := provider.Get(source, imageserver.Params{})
	if err == nil {
		t.Fatal("no error")
	}
}

type errorReadCloser struct{}

func (erc *errorReadCloser) Read(p []byte) (n int, err error) {
	return 0, errors.New("error")
}

func (erc *errorReadCloser) Close() error {
	return errors.New("error")
}

func TestParseResponseErrorData(t *testing.T) {
	provider := &Provider{}

	response := &http.Response{
		StatusCode: http.StatusOK,
		Body:       &errorReadCloser{},
	}

	_, err := provider.parseResponse(response)
	if err == nil {
		t.Fatal("no error")
	}
}

func createTestHTTPServer(t *testing.T) *net.TCPListener {
	addr, err := net.ResolveTCPAddr("tcp", "")
	if err != nil {
		t.Fatal(err)
	}

	listener, err := net.ListenTCP("tcp", addr)
	if err != nil {
		t.Fatal(err)
	}

	server := &http.Server{
		Handler: http.FileServer(http.Dir(testdata.Dir)),
	}
	go server.Serve(listener)

	return listener
}

func createTestURL(listener *net.TCPListener) *url.URL {
	return &url.URL{
		Scheme: "http",
		Host:   listener.Addr().String(),
		Path:   testSourceFileName,
	}
}
