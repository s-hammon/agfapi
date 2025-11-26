package agfa

import (
	"net/http"
	"net/http/cookiejar"
	"strings"
)

const DefaultDomain = "Agility"

type Client struct {
	User           string
	Pass           string
	Domain         string
	BaseUrl        string
	TokenUrl       string
	ClientId       string
	RedirectListId string
	VerifySsl      bool

	hc          *http.Client
	authHeaders map[string]string
}

func NewClient(url string, opts ...func(*Client)) *Client {
	jar, _ := cookiejar.New(nil)

	client := &Client{
		BaseUrl: url,
		Domain:  DefaultDomain,
		hc: &http.Client{
			Jar: jar,
		},
	}

	for _, opt := range opts {
		opt(client)
	}

	return client
}

// Base returns the base URL
func (client *Client) Base() string {
	return strings.TrimRight(client.BaseUrl, "/")
}

// temporarily disable following a redirect for a request
// returns a function which toggles this back to default
func (client *Client) disableCheckRedirect() func() {
	client.hc.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}

	return func() {
		client.hc.CheckRedirect = nil
	}
}
