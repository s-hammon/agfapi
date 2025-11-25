package agfa

import (
	"errors"
	"net/http"

	"golang.org/x/net/html"
)

func extractForm(resp *http.Response) (string, map[string]string, error) {
	z := html.NewTokenizer(resp.Body)

	inForm := false
	action := ""
	inputs := make(map[string]string)

	for {
		switch tt := z.Next(); tt {
		case html.ErrorToken:
			if inForm {
				return action, inputs, nil
			}

			return "", nil, errors.New("form not found")
		case html.StartTagToken:
			t := z.Token()
			if t.Data == "form" {
				inForm = true
				for _, attr := range t.Attr {
					if attr.Key == "action" {
						action = attr.Val
					}
				}
			}

			if inForm && t.Data == "input" {
				var name, value string
				for _, attr := range t.Attr {
					switch attr.Key {
					case "name":
						name = attr.Val
					case "value":
						value = attr.Val
					}
				}

				if name != "" {
					inputs[name] = value
				}
			}
		case html.EndTagToken:
			t := z.Token()
			if inForm && t.Data == "form" {
				if action == "" {
					return "", nil, errors.New("form action not found")
				}

				return action, inputs, nil
			}
		}
	}
}
