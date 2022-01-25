package collector

import (
	"context"
	"net"
	"sync"
	"testing"
	"time"

	"github.com/devon-mar/tacacs-exporter/config"
	"github.com/nwaples/tacplus"
)

const (
	testServerAddress = "127.0.0.1:4949"
)

type testHandler struct {
	mtx               sync.Mutex
	authenReply       *tacplus.AuthenReply
	authStartCount    int
	authzRequestCount int
	acctRequestCount  int
}

// HandleAuthenStart implements tacplus.RequestHandler
func (th *testHandler) HandleAuthenStart(ctx context.Context, a *tacplus.AuthenStart, s *tacplus.ServerSession) *tacplus.AuthenReply {
	th.mtx.Lock()
	th.authStartCount++
	th.mtx.Unlock()
	return th.authenReply
}

// HandleAuthenRequest implements tacplus.RequestHandler
func (th *testHandler) HandleAuthorRequest(ctx context.Context, a *tacplus.AuthorRequest, s *tacplus.ServerSession) *tacplus.AuthorResponse {
	th.mtx.Lock()
	th.authzRequestCount++
	th.mtx.Unlock()
	return nil
}

// HandleAuthenStart implements tacplus.RequestHandler
func (th *testHandler) HandleAcctRequest(ctx context.Context, a *tacplus.AcctRequest, s *tacplus.ServerSession) *tacplus.AcctReply {
	th.mtx.Lock()
	th.acctRequestCount++
	th.mtx.Unlock()
	return nil
}

func newTestServer(secret string, authenReplyStatus uint8) (*tacplus.Server, *testHandler, error) {
	th := &testHandler{authenReply: &tacplus.AuthenReply{Status: authenReplyStatus}}

	sch := tacplus.ServerConnHandler{
		Handler: th,
		ConnConfig: tacplus.ConnConfig{
			Mux:          false,
			LegacyMux:    false,
			Secret:       []byte(secret),
			IdleTimeout:  time.Second * 5,
			ReadTimeout:  time.Second * 5,
			WriteTimeout: time.Second * 5,
		},
	}

	s := &tacplus.Server{ServeConn: sch.Serve}
	return s, th, nil
}

func getTestCollector(address string, secret string) Collector {
	m := config.Module{
		Username: "test",
		Password: "password",
		Secret:   []byte(secret),
		Timeout:  time.Second * 5,
		Port:     "test",
	}
	return NewCollector(address, "127.0.0.2", &m)
}

func TestProbe(t *testing.T) {
	c := getTestCollector(testServerAddress, "5ecr3t")
	server, handler, err := newTestServer("5ecr3t", tacplus.AuthenStatusPass)
	if err != nil {
		t.Fatalf("error starting server: %v", err)
	}

	l, err := net.Listen("tcp", testServerAddress)
	if err != nil {
		t.Fatalf("error creating listener server on %s: %v", testServerAddress, err)
	}
	defer l.Close()

	go func() { server.Serve(l) }()

	err = c.probe()
	l.Close()
	if err != nil {
		t.Errorf("expected no error but got: %v", err)
	}

	handler.mtx.Lock()
	defer handler.mtx.Unlock()
	if handler.authStartCount != 1 {
		t.Errorf("Expected 1 auth start requests but got %d", handler.authStartCount)
	}
	if handler.acctRequestCount != 0 {
		t.Errorf("Expected 0 acct requests but got %d", handler.acctRequestCount)
	}
	if handler.authzRequestCount != 0 {
		t.Errorf("Expected 0 authz requests but got %d", handler.authzRequestCount)
	}
}

func TestProbeInvalidSecret(t *testing.T) {
	c := getTestCollector(testServerAddress, "invalid secret set on the collector")
	server, _, err := newTestServer("5ecr3t", tacplus.AuthenStatusPass)
	if err != nil {
		t.Fatalf("error starting server: %v", err)
	}

	l, err := net.Listen("tcp", testServerAddress)
	if err != nil {
		t.Fatalf("error creating listener server on %s: %v", testServerAddress, err)
	}
	defer l.Close()

	go func() { server.Serve(l) }()

	err = c.probe()
	l.Close()
	if err == nil {
		t.Error("expected an error")
	}
}

func TestProbeInvalidHost(t *testing.T) {
	c := getTestCollector("127.0.0.1:4950", "invalid secret set on the collector")
	c.Module.Timeout = time.Second
	server, _, err := newTestServer("5ecr3t", tacplus.AuthenStatusPass)
	if err != nil {
		t.Fatalf("error starting server: %v", err)
	}

	l, err := net.Listen("tcp", testServerAddress)
	if err != nil {
		t.Fatalf("error creating listener server on %s: %v", testServerAddress, err)
	}
	defer l.Close()

	go func() { server.Serve(l) }()

	err = c.probe()
	l.Close()
	if err == nil {
		t.Error("expected an error")
	}
}
