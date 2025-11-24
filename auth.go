package main

import (
	"crypto/tls"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
)

const DefaultDomain = "Agility"

type ManualAuth struct {
	User           string
	Pass           string
	Domain         string
	BaseUrl        string
	TokenUrl       string
	ClientId       string
	RedirectListId string
	VerifySsl      bool

	client      *http.Client
	authHeaders map[string]string
}

func NewManualAuth(user, pass string, opts ...func(*ManualAuth)) (*ManualAuth, error) {
	jar, _ := cookiejar.New(nil)

	m := &ManualAuth{
		User: user,
		Pass: pass,
		client: &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			},
			Jar: jar,
		},
	}

	for _, opt := range opts {
		opt(m)
	}

	return m, nil
}

func (m *ManualAuth) getRedirectUrl(base string) (string, error) {
	log.Println("redirecting to login...")
	path := "/List"
	if m.RedirectListId != "" {
		path = "/List?_id=" + m.RedirectListId
	}

	initUrl := base + path
	log.Printf("initial request: %s\n", initUrl)

	m.client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}
	defer func() {
		m.client.CheckRedirect = nil
	}()

	resp, err := m.client.Get(initUrl)
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

func (m *ManualAuth) getAuthRedirect(req *http.Request) (string, error) {
	log.Println("logging in...")

	m.client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}
	defer func() {
		m.client.CheckRedirect = nil
	}()

	postResp, err := m.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("postResp.Do: %v", err)
	}
	defer postResp.Body.Close()

	if postResp.StatusCode == http.StatusOK {
		return "", fmt.Errorf("login failed, expected redirect")
	}

	if postResp.StatusCode != http.StatusFound && postResp.StatusCode != http.StatusSeeOther {
		return "", fmt.Errorf("login submission failed: %d", postResp.StatusCode)
	}

	authRedirect := postResp.Header.Get("Location")
	if authRedirect == "" {
		return "", errors.New("no redirect after login...")
	}

	return authRedirect, nil
}

func (m *ManualAuth) GetAuthHeader() (map[string]string, error) {
	if m.authHeaders != nil {
		return m.authHeaders, nil
	}
	if m.BaseUrl == "" {
		return nil, errors.New("auth.GetAuthHeader: must provide base url")
	}

	base := strings.TrimRight(m.BaseUrl, "/")

	redirectUrl, err := m.getRedirectUrl(base)
	if err != nil {
		return nil, fmt.Errorf("auth.getRedirectUrl: %v", err)
	}

	loginResp, err := m.client.Get(redirectUrl)
	if err != nil {
		return nil, fmt.Errorf("loginResp.Get: %v", err)
	}
	defer loginResp.Body.Close()

	actionUrl, err := getActionUrl(loginResp)
	if err != nil {
		return nil, err
	}

	payload := newForm(m.User, m.Pass)
	origin := loginResp.Request.URL.Scheme + "://" + loginResp.Request.URL.Host

	req, err := http.NewRequest(http.MethodPost, actionUrl, strings.NewReader(payload))
	if err != nil {
		return nil, fmt.Errorf("newFormRequest: %v", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Referer", loginResp.Request.URL.String())
	req.Header.Set("Origin", origin)
	req.Header.Set("User-Agent", "Mozilla/5.0")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")

	authRedirect, err := m.getAuthRedirect(req)

	authResp, err := m.client.Get(authRedirect)
	if err != nil {
		return nil, fmt.Errorf("authResp.Get: %v", err)
	}
	defer authResp.Body.Close()

	finalUrl := authResp.Request.URL.String()
	if strings.HasPrefix(finalUrl, base) {
		log.Println("session-based auth confirmed")

		m.authHeaders = map[string]string{}
		return m.authHeaders, nil
	}

	parsed, _ := url.Parse(finalUrl)
	code := parsed.Query().Get("code")
	if code == "" {
		return nil, errors.New("no authorization code found in redirect")
	}

	return nil, errors.New("could not resolve authentication")
}

func (m *ManualAuth) Base() string {
	return strings.TrimRight(m.BaseUrl, "/")
}
