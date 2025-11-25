package agfa

import (
	"crypto/tls"
	"errors"
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"net/url"
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
			Transport: &http.Transport{
				// TODO: get certs so that we're not making insecure connections
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			},
			Jar: jar,
		},
	}

	for _, opt := range opts {
		opt(client)
	}

	return client
}

type SessionParams struct {
	ClientId string
	Username string
	Password string
}

// This is a hacky way to create a session with your client by impersonating
// someone logging into the remote system thru their browser. :^)
// Will fail if any part of the auth pipeline returns something "unexpected".
func (client *Client) Session(params SessionParams) (*Client, error) {
	client.ClientId = params.ClientId
	client.User = params.Username
	client.Pass = params.Password

	base := strings.TrimRight(client.BaseUrl, "/")

	// get redirect url from initial request attempt
	url, err := client.getRedirectUrl(base)
	if err != nil {
		return nil, fmt.Errorf("auth.getRedirectUrl: %v", err)
	}

	// follow login redirect
	resp, err := client.hc.Get(url)
	if err != nil {
		return nil, fmt.Errorf("login request failed: %v", err)
	}

	// generate form submission request
	req, err := newFormRequest(client.User, client.Pass, resp)
	if err != nil {
		return nil, err
	}

	// submit form & obtain redirect
	url, err = client.getAuthRedirect(req)
	if err != nil {
		return nil, err
	}

	resp, err = client.hc.Get(url)
	if err != nil {
		return nil, err
	}
	resp.Body.Close()

	if !strings.HasPrefix(resp.Request.URL.String(), base) {
		return nil, errors.New("could not resolve session-based authorization")
	}

	return client, nil
}

func (client *Client) getRedirectUrl(base string) (string, error) {
	path := "/List"
	if client.RedirectListId != "" {
		path = "/List?_id=" + client.RedirectListId
	}

	initUrl := base + path

	client.hc.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}
	defer func() {
		client.hc.CheckRedirect = nil
	}()

	resp, err := client.hc.Get(initUrl)
	if err != nil {
		return "", fmt.Errorf("init.Get: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusFound {
		return "", fmt.Errorf("expected 302 redirect, got %d", resp.StatusCode)
	}

	redirectUrl := resp.Header.Get("Location")
	if redirectUrl == "" {
		return "", errors.New("no redirect location")
	}

	return redirectUrl, nil
}

func newForm(user, pass string) string {
	form := url.Values{}
	form.Set("username", user)
	form.Set("password", pass)
	form.Set("login", "Sign In")

	return form.Encode()
}

func getActionUrl(resp *http.Response) (string, error) {
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("login page returned %d", resp.StatusCode)
	}

	actionUrl, _, err := extractForm(resp)
	if err != nil {
		return "", fmt.Errorf("extractForm: %v", err)
	}

	return actionUrl, nil
}

func (client *Client) getAuthRedirect(req *http.Request) (string, error) {
	reset := client.disableCheckRedirect()
	defer reset()

	postResp, err := client.hc.Do(req)
	if err != nil {
		return "", fmt.Errorf("postResp.Do: %v", err)
	}
	defer postResp.Body.Close()

	if postResp.StatusCode != http.StatusFound && postResp.StatusCode != http.StatusSeeOther {
		return "", fmt.Errorf("login submission failed: expected redirect, got %d", postResp.StatusCode)
	}

	authRedirect := postResp.Header.Get("Location")
	if authRedirect == "" {
		return "", errors.New("no redirect after login...")
	}

	return authRedirect, nil
}

func newFormRequest(user, pass string, resp *http.Response) (*http.Request, error) {
	defer resp.Body.Close()

	url, err := getActionUrl(resp)
	if err != nil {
		return nil, fmt.Errorf("newFormRequest: %v", err)
	}

	payload := newForm(user, pass)

	req, err := http.NewRequest(http.MethodPost, url, strings.NewReader(payload))
	if err != nil {
		return nil, fmt.Errorf("newFormRequest: %v", err)
	}

	origin := resp.Request.URL.Scheme + "://" + resp.Request.URL.Host
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Referer", resp.Request.URL.String())
	req.Header.Set("Origin", origin)
	req.Header.Set("User-Agent", "Mozilla/5.0")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")

	return req, nil
}

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
