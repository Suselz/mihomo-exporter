package mihomo

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestClientFetchEndpoints(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/connections":
			_, _ = w.Write([]byte(`{"uploadTotal":10,"downloadTotal":20,"connections":[]}`))
		case "/memory":
			_, _ = w.Write([]byte(`{"inuse":123,"oslimit":456}`))
		case "/traffic":
			_, _ = w.Write([]byte(`{"up":1,"down":2,"upTotal":3,"downTotal":4}`))
		case "/version":
			_, _ = w.Write([]byte(`{"meta":true,"version":"v1"}`))
		case "/proxies":
			_, _ = w.Write([]byte(`{"proxies":{"a":{"name":"a","type":"Selector","alive":true}}}`))
		case "/rules":
			_, _ = w.Write([]byte(`{"rules":[{"type":"InName","extra":{"disabled":false,"hitCount":7,"missCount":9}}]}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer ts.Close()

	c := NewClient(ts.URL, "/connections")
	ctx := context.Background()

	if s, err := c.Fetch(ctx); err != nil || s.UploadTotal != 10 || s.DownloadTotal != 20 {
		t.Fatalf("Fetch() err=%v snap=%+v", err, s)
	}
	if m, err := c.FetchMemory(ctx); err != nil || m.Inuse != 123 || m.OSLimit != 456 {
		t.Fatalf("FetchMemory() err=%v snap=%+v", err, m)
	}
	if tr, err := c.FetchTraffic(ctx); err != nil || tr.DownTotal != 4 {
		t.Fatalf("FetchTraffic() err=%v snap=%+v", err, tr)
	}
	if v, err := c.FetchVersion(ctx); err != nil || v.Version != "v1" || !v.Meta {
		t.Fatalf("FetchVersion() err=%v snap=%+v", err, v)
	}
	if p, err := c.FetchProxies(ctx); err != nil || len(p.Proxies) != 1 {
		t.Fatalf("FetchProxies() err=%v snap=%+v", err, p)
	}
	if r, err := c.FetchRules(ctx); err != nil || len(r.Rules) != 1 || r.Rules[0].Extra.HitCount != 7 {
		t.Fatalf("FetchRules() err=%v snap=%+v", err, r)
	}
}

func TestClientFetchDecodeError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"uploadTotal":`))
	}))
	defer ts.Close()

	c := NewClient(ts.URL, "/connections")
	_, err := c.Fetch(context.Background())
	if err == nil {
		t.Fatal("expected decode error")
	}
	if !errors.Is(err, ErrDecode) {
		t.Fatalf("expected ErrDecode, got: %v", err)
	}
}
