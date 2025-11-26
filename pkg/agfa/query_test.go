package agfa

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

type objStruct struct {
	Foo string `json:"foo"`
}

func newClientWithServer(ts *httptest.Server) *Client {
	return &Client{
		BaseUrl: ts.URL,
		hc:      ts.Client(),
		authHeaders: map[string]string{
			"Authorization": "Bearer token",
		},
	}
}

func TestReqUrl(t *testing.T) {
	client := &Client{BaseUrl: "https://example.com/api"}
	u := client.reqUrl("test/123")
	require.Equal(t, "https://example.com/api/test/123", u.String())
}

func TestClientGet(t *testing.T) {
	want := objStruct{Foo: "bar"}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "application/fhir+json", r.Header.Get("Accept"))
		require.Equal(t, "Bearer token", r.Header.Get("Authorization"))
		require.Equal(t, "/test/123", r.URL.Path)

		params := strings.Split(r.URL.RawQuery, "&")
		require.Len(t, params, 2)
		require.Contains(t, params, "q=xyz")
		require.Contains(t, params, "limit=10")

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(want)
	}))
	defer ts.Close()

	client := newClientWithServer(ts)
	params := map[string]string{
		"q":     "xyz",
		"limit": "10",
	}

	var got objStruct
	err := client.Get("test/123", params, &got)
	require.NoError(t, err)
	require.Equal(t, want, got)
}

func TestClientGet_ErrorWithBody(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		io.WriteString(w, "your bad")
	}))
	defer ts.Close()

	client := newClientWithServer(ts)

	var got objStruct
	err := client.Get("bad/request", nil, &got)
	require.Error(t, err)
	require.Contains(t, err.Error(), "your bad")
	require.Contains(t, err.Error(), "400")
}

func TestReqUrl_Panic(t *testing.T) {
	client := &Client{BaseUrl: ":"}
	require.Panics(t, func() {
		_ = client.reqUrl("List", "123")
	})
}

func TestClientGet_NetworkError(t *testing.T) {
	ts := httptest.NewServer(nil)
	ts.Close()

	client := &Client{
		BaseUrl: ts.URL,
		hc:      &http.Client{},
	}

	var got objStruct
	err := client.Get("server/down", nil, &got)
	require.Error(t, err)
}
