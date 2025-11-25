package agfa

import (
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

	rec := httptest.NewRecorder()
	rec.WriteString(html)
	resp := rec.Result()

	action, hidden, err := extractForm(resp)
	require.NoError(t, err)
	require.Equal(t, "/auth/realms/Agility/login", action)
	require.Equal(t, map[string]string{
		"session_code": "abc123",
		"execution":    "xyz789",
		"client_id":    "myclient",
		"username":     "",
		"password":     "",
	}, hidden)
}
