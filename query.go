package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

func (m *ManualAuth) reqUrl(endpoint ...string) *url.URL {
	path, err := url.JoinPath(m.Base(), endpoint...)
	if err != nil {
		panic(err)
	}

	u, err := url.Parse(path)
	if err != nil {
		panic(err)
	}

	return u
}

func (m *ManualAuth) get(req *http.Request) (*http.Response, error) {
	for k, v := range m.authHeaders {
		req.Header.Set(k, v)
	}
	req.Header.Set("Accept", "application/fhir+json")

	resp, err := m.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http GET: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("FHIR GET failed: %d %s", resp.StatusCode, body)
	}

	return resp, nil
}

func (m *ManualAuth) FetchListById(ctx context.Context, listId string) (io.Reader, error) {
	u := m.reqUrl("List")

	q := u.Query()
	q.Set("_id", listId)
	q.Set("_format", "json")
	// q.Set("_include:iterate", "List:entry.item")
	u.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("http.NewRequestWithContext: %v", err)
	}

	for k, v := range m.authHeaders {
		req.Header.Set(k, v)
	}
	req.Header.Set("Accept", "application/fhir+json")

	resp, err := m.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http GET: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("FHIR GET failed: %d %s", resp.StatusCode, body)
	}

	return resp.Body, nil
}

func (m *ManualAuth) FetchTaskById(ctx context.Context, taskId string) (Task, error) {
	u := m.reqUrl("Task", taskId)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return Task{}, fmt.Errorf("http.NewRequestWithContext: %v", err)
	}

	resp, err := m.get(req)
	if err != nil {
		return Task{}, err
	}
	defer resp.Body.Close()

	var task Task
	if err := json.NewDecoder(resp.Body).Decode(&task); err != nil {
		return Task{}, fmt.Errorf("json.Decode(resp.Body): %v", err)
	}

	return task, nil
}

func (m *ManualAuth) FetchServiceRequestById(ctx context.Context, reqId string) (ServiceRequest, error) {
	u := m.reqUrl("ServiceRequest", reqId)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return ServiceRequest{}, fmt.Errorf("http.NewRequestWithContext: %v", err)
	}

	resp, err := m.get(req)
	if err != nil {
		return ServiceRequest{}, err
	}
	defer resp.Body.Close()

	return decode(resp.Body, ServiceRequest{})
}

func decode[T any](r io.Reader, obj T) (T, error) {
	err := json.NewDecoder(r).Decode(&obj)
	if err != nil {
		return obj, fmt.Errorf("json.Decode(%T): %v", obj, err)
	}

	return obj, nil
}
