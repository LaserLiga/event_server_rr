package eventserver

import (
	"context"
	"net"
	"net/http"
	"testing"
	"time"

	"go.uber.org/zap"
)

func newTestPlugin(address string) *Plugin {
	return &Plugin{
		cfg:     &Config{Address: address},
		log:     zap.NewNop(),
		metrics: newStatsExporter(),
		state:   stateStopped,
	}
}

func TestTriggerEventBeforeServeReturnsError(t *testing.T) {
	p := newTestPlugin("127.0.0.1:0")
	r := p.RPC().(*rpc)

	var out int
	if err := r.TriggerEvent(EventMessage{Type: "created", Message: "payload"}, &out); err == nil {
		t.Fatal("expected error when triggering before server is running")
	}
}

func TestServeAddressInUseReportsUnavailable(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()

	p := newTestPlugin(ln.Addr().String())
	errCh := p.Serve()

	select {
	case err := <-errCh:
		if err == nil {
			t.Fatal("expected bind error")
		}
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for bind error")
	}

	st, err := p.Ready()
	if err != nil {
		t.Fatal(err)
	}
	if st.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected unavailable status, got %d", st.Code)
	}
}

func TestServeTriggerStopLifecycle(t *testing.T) {
	p := newTestPlugin("127.0.0.1:0")
	p.Serve()

	st, err := p.Ready()
	if err != nil {
		t.Fatal(err)
	}
	if st.Code != http.StatusOK {
		t.Fatalf("expected ready status, got %d", st.Code)
	}

	r := p.RPC().(*rpc)
	var out int
	if err := r.TriggerEvent(EventMessage{Type: "created", Message: "payload"}, &out); err != nil {
		t.Fatal(err)
	}
	if out != 2 {
		t.Fatalf("expected event id 2, got %d", out)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err := p.Stop(ctx); err != nil {
		t.Fatal(err)
	}

	st, err = p.Ready()
	if err != nil {
		t.Fatal(err)
	}
	if st.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected unavailable after stop, got %d", st.Code)
	}

	if err := r.TriggerEvent(EventMessage{Type: "created"}, &out); err == nil {
		t.Fatal("expected error after stop")
	}
}

func TestResetBeforeServeStartsServer(t *testing.T) {
	p := newTestPlugin("127.0.0.1:0")

	if err := p.Reset(); err != nil {
		t.Fatal(err)
	}

	st, err := p.Ready()
	if err != nil {
		t.Fatal(err)
	}
	if st.Code != http.StatusOK {
		t.Fatalf("expected ready after reset, got %d", st.Code)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err := p.Stop(ctx); err != nil {
		t.Fatal(err)
	}
}

func TestServeDuringResetReturnsError(t *testing.T) {
	p := newTestPlugin("127.0.0.1:0")
	p.state = stateResetting

	errCh := p.Serve()

	select {
	case err := <-errCh:
		if err == nil {
			t.Fatal("expected reset-in-progress error")
		}
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for reset-in-progress error")
	}

	st, err := p.Ready()
	if err != nil {
		t.Fatal(err)
	}
	if st.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected unavailable while resetting, got %d", st.Code)
	}
}

func TestMetricsCollectorBeforeInitIsSafe(t *testing.T) {
	var p Plugin
	if collectors := p.MetricsCollector(); collectors != nil {
		t.Fatalf("expected nil collectors before init, got %d", len(collectors))
	}
}
