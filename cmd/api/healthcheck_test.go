package main

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

type testServer struct {
	*httptest.Server
}

func TestHealthcheck(t *testing.T) {
	app := newTestApp()
	ts := newTestServer(app.routes())
	defer ts.Close()

	code, _, body := ts.get(t, "/v1/healthcheck")

	if code != http.StatusOK {
		t.Errorf("want %d, got %d", http.StatusOK, code)
	}

	resp := `{
	"status": "available",
	"system_info": {
		"environment": "testing",
		"version": "1.0.0"
	}
}
`

	if string(body) != resp {
		t.Errorf("want body to equal %q,\n but got %q", resp, string(body))
	}
}

func newTestApp() *application {
	app := new(application)
	cfg := config{env: "testing"}
	app.config = cfg

	return app
}

func newTestServer(h http.Handler) *testServer {
	ts := httptest.NewServer(h)

	ts.Client().CheckRedirect = func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}
	return &testServer{ts}
}

func (ts *testServer) get(t *testing.T, urlPath string) (int, http.Header, []byte) {
	rs, err := ts.Client().Get(ts.URL + urlPath)
	if err != nil {
		t.Fatal(err)
	}

	defer func() {
		if err := rs.Body.Close(); err != nil {
			t.Fatal(err)
		}
	}()

	body, err := io.ReadAll(rs.Body)
	if err != nil {
		t.Fatal(err)
	}

	return rs.StatusCode, rs.Header, body
}
