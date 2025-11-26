package agfa

import (
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func buildLoginHTML(action string) string {
	return `
	<html>
	    <body>
	        <form action="` + action + `" method="post">
	            <input type="hidden" name="session_code" value="abc">
	            <input type="hidden" name="execution" value="xyz">
	            <input type="hidden" name="client_id" value="my-client">
	            <input type="text" name="username">
	            <input type="password" name="password">
	        </form>
	    </body>
	</html>`
}

func TestGetRedirectUrl(t *testing.T) {
	var capturedReq *http.Request

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedReq = r
		w.Header().Set("Location", "/login")
		w.WriteHeader(http.StatusFound)
	}))
	defer ts.Close()

	c := NewClient(ts.URL)

	url, err := c.getRedirectUrl(strings.TrimRight(ts.URL, "/"))
	require.NoError(t, err)
	require.Equal(t, "/login", url)
	require.Equal(t, "/List", capturedReq.URL.Path)
}

func TestGetRedirectUrl_NoRedirect(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	c := NewClient(ts.URL)

	_, err := c.getRedirectUrl(strings.TrimRight(ts.URL, "/"))
	require.Error(t, err)
	require.Contains(t, err.Error(), "expected 302")
}

func TestNewFormRequest(t *testing.T) {
	ep := "/submit-login"
	loginHTML := buildLoginHTML(ep)
	resp := makeResp(t, loginHTML)
	resp.Request = &http.Request{
		Method: http.MethodGet,
		URL: &url.URL{
			Scheme: "https",
			Host:   "example.com",
			Path:   "/login",
		},
	}

	req, err := newFormRequest("testuser", "testpass", resp)
	require.NoError(t, err)
	require.Equal(t, http.MethodPost, req.Method)
	require.Equal(t, ep, req.URL.Path)

	require.Equal(t, "application/x-www-form-urlencoded", req.Header.Get("Content-Type"))
	require.Equal(t, "https://example.com/login", req.Header.Get("Referer"))
	require.Equal(t, "example.com", req.Header.Get("Origin")[8:])
	require.Equal(t, "Mozilla/5.0", req.Header.Get("User-Agent"))

	body, _ := io.ReadAll(req.Body)
	values, _ := url.ParseQuery(string(body))
	require.Equal(t, "testuser", values.Get("username"))
	require.Equal(t, "testpass", values.Get("password"))
	require.Equal(t, "Sign In", values.Get("login"))
}

func TestGetAuthRedirect(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Location", "/final")
		w.WriteHeader(http.StatusFound)
	}))
	defer ts.Close()

	c := NewClient(ts.URL)

	req, _ := http.NewRequest(http.MethodPost, ts.URL, nil)
	redirect, err := c.getAuthRedirect(req)
	require.NoError(t, err)
	require.Equal(t, "/final", redirect)
}

func TestGetAuthRedirect_NoLocation(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer ts.Close()

	c := NewClient(ts.URL)

	req, _ := http.NewRequest(http.MethodPost, ts.URL, nil)
	_, err := c.getAuthRedirect(req)
	require.Error(t, err)
}
