package agfa

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestExtractForm(t *testing.T) {
	html := `
	<html>
	    <body>
	        <form action="/auth/realms/Agility/login" method="post">
	            <input type="hidden" name="session_code" value="abc123">
	            <input type="hidden" name="execution" value="xyz789">
	            <input type="hidden" name="client_id" value="myclient">
	            <input type="text" name="username">
	            <input type="password" name="password">
	        </form>
	    </body>
	</html>
`

	resp := makeResp(t, html)

	action, inputs, err := extractForm(resp)
	require.NoError(t, err)
	require.Equal(t, "/auth/realms/Agility/login", action)
	require.Equal(t, map[string]string{
		"session_code": "abc123",
		"execution":    "xyz789",
		"client_id":    "myclient",
		"username":     "",
		"password":     "",
	}, inputs)

	html = `<html><body>No forms found.</body></html>`

	resp = makeResp(t, html)

	action, inputs, err = extractForm(resp)
	require.Error(t, err)
	require.Contains(t, err.Error(), "form not found")
	require.Empty(t, action)
	require.Nil(t, inputs)

	html = `
	<html>
	    <body>
	        <form>
		    <input type="hidden" name="x" value="1">
		</form>
	    </body>
	</html>
`

	resp = makeResp(t, html)

	action, inputs, err = extractForm(resp)
	require.Error(t, err)
	require.Contains(t, err.Error(), "form action not found")
	require.Empty(t, action)
	require.Nil(t, inputs)

	html = `
	<html>
	    <body>
	        <form action="/first">
		    <input type="hidden" name="a" value="1">
		</form>
		<form action="/second">
		    <input type="hidden" name="b" value"2">
		</form>
	    </body>
	</html>
`

	resp = makeResp(t, html)

	action, inputs, err = extractForm(resp)
	require.NoError(t, err)
	require.Equal(t, "/first", action)
	require.Equal(t, map[string]string{"a": "1"}, inputs)

	html = `
	<html>
	    <body>
	        <form action="/login"></form>
	    </body>
	</html>
`

	resp = makeResp(t, html)

	action, inputs, err = extractForm(resp)
	require.NoError(t, err)
	require.Equal(t, "/login", action)
	require.Empty(t, inputs)

	html = `
	<html>
	    <body>
	        <form action="/ok">
		    <input type="hidden" name="x" value="y">
		<!-- Broken attributes; should recover -->
	    </body>
	</html>
`

	resp = makeResp(t, html)

	action, inputs, err = extractForm(resp)
	require.NoError(t, err)
	require.Equal(t, "/ok", action)
	require.Equal(t, map[string]string{"x": "y"}, inputs)

	html = `
	<html>
	    <body>
	        <div <span <p> missing closing tags
	    </body>
	</html>
`

	resp = makeResp(t, html)

	action, inputs, err = extractForm(resp)
	require.Error(t, err)
	require.Contains(t, err.Error(), "form not found")
	require.Empty(t, action)
	require.Nil(t, inputs)
}

func makeResp(t *testing.T, html string) *http.Response {
	t.Helper()

	rec := httptest.NewRecorder()
	rec.WriteString(html)
	return rec.Result()
}
