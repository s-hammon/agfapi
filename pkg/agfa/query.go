package agfa

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/s-hammon/p"
)

func (client *Client) reqUrl(endpoint ...string) *url.URL {
	path, err := url.JoinPath(client.Base(), endpoint...)
	if err != nil {
		panic(err)
	}

	u, err := url.Parse(path)
	if err != nil {
		panic(err)
	}

	return u
}

func (client *Client) Get(endpoint string, params map[string]string, obj any) error {
	u := client.reqUrl(endpoint)
	if len(params) != 0 {
		q := u.Query()
		for k, v := range params {
			q.Set(k, v)
		}
		u.RawQuery = q.Encode()
	}

	resp, err := client.get(u)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return decode(resp.Body, obj)
}

func (client *Client) get(u *url.URL) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("http.NewRequest: %v", err)
	}

	for k, v := range client.authHeaders {
		req.Header.Set(k, v)
	}
	req.Header.Set("Accept", "application/fhir+json")

	resp, err := client.hc.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http GET: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("FHIR GET failed: %d %s", resp.StatusCode, body)
	}

	return resp, nil
}

func (client *Client) FetchListById(listId string) (List, error) {
	params := map[string]string{
		"_format": "json",
	}

	var list List
	err := client.Get(p.Format("List/%s", listId), params, &list)
	return list, err
}

func (client *Client) FetchBundleById(bundleId string) (Bundle, error) {
	params := map[string]string{
		"_id":     bundleId,
		"_format": "json",
		// NOTE: I don't think we have to set this.
		// "_include:iterate": "List:entry.item",
	}
	var bundle Bundle
	err := client.Get("List", params, &bundle)
	return bundle, err
}

func (client *Client) FetchTaskById(taskId string) (Task, error) {
	params := map[string]string{
		"_format": "json",
	}
	var task Task
	err := client.Get(p.Format("Task/%s", taskId), params, &task)
	return task, err
}

func (client *Client) FetchServiceRequestById(reqId string) (ServiceRequest, error) {
	params := map[string]string{
		"_id":     reqId,
		"_format": "json",
	}
	var serviceRequest ServiceRequest
	err := client.Get(p.Format("ServiceRequest/%s", reqId), params, &serviceRequest)
	return serviceRequest, err
}

func decode(r io.Reader, obj any) error {
	return json.NewDecoder(r).Decode(obj)
}
