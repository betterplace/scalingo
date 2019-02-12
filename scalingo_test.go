package scalingo

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPrepareRequestWithoutPrefix(t *testing.T) {
	s := NewScalingo("test")
	req := s.PrepareRequest("POST", "", "/v1/apps/foo/domains", "", nil)
	assert.Equal(t, "https://api.scalingo.com/v1/apps/foo/domains", req.URL.String(), "correct URL")
}

func TestPrepareRequestAddsPathPrefix(t *testing.T) {
	s := NewScalingo("test")
	req := s.PrepareRequest("POST", "", "v1/apps/foo/domains", "", nil)
	assert.Equal(t, "https://api.scalingo.com/v1/apps/foo/domains", req.URL.String(), "correct URL")
}

func TestPrepareRequestWithoutPrefixAndRequestBody(t *testing.T) {
	s := NewScalingo("test")
	req := s.PrepareRequestForURL("POST", "https://foo.bar/v1/apps/foo/domains", "test", bytes.NewBufferString(`{"foo":"bar"}`))
	var body = make([]byte, 13)
	n, err := req.Body.Read(body)
	assert.Nil(t, err, "no errors")
	assert.Equal(t, n, 13, "correct body length")
	assert.Equal(t, `{"foo":"bar"}`, string(body), "correct body")
	assert.Equal(t, "application/json", req.Header.Get("Accept"), "accept header set")
	assert.Equal(t, "application/json", req.Header.Get("Content-Type"), "content-type header set")
}

func TestPrepareRequestWithPrefix(t *testing.T) {
	s := NewScalingo("test")
	req := s.PrepareRequest("POST", "https://foo.bar", "/v1/apps/foo/domains", "", nil)
	assert.Equal(t, "https://foo.bar/v1/apps/foo/domains", req.URL.String(), "URL without Prefix")
}

func TestPrepareRequestForURLWithoutToken(t *testing.T) {
	s := NewScalingo("test")
	req := s.PrepareRequestForURL("POST", "http://foo.bar/v1/apps/foo/domains", "", nil)
	assert.Equal(t, 0, len(req.Header["Authorization"]), "Authorization not set")
}

func TestPrepareRequestForURLWithToken(t *testing.T) {
	s := NewScalingo("test")
	req := s.PrepareRequestForURL("POST", "http://foo.bar/v1/apps/foo/domains", "test", nil)
	assert.Equal(t, "Basic OnRlc3Q=", req.Header["Authorization"][0], "Authorization set")
}

func TestFetchBearerToken(t *testing.T) {
	s := NewScalingo("")
	token := s.fetchBearerToken()
	assert.NotEmpty(t, token, "Token is present")
}

func TestFetchBearerTokenFailure(t *testing.T) {
	s := NewScalingo("test")
	token := s.fetchBearerToken()
	assert.Empty(t, token, "Token is not present")
}

func TestGetAppNames(t *testing.T) {
	s := NewScalingo("")
	appNames := s.GetAppNames()
	assert.NotEmpty(t, appNames, "array of app names was returned")
}

func TestGetAppVariables(t *testing.T) {
	s := NewScalingo("")
	appNames := s.GetAppNames()
	if len(appNames) > 0 {
		appVars := s.GetAppVariables(appNames[0])
		assert.NotEmpty(t, appVars, "map of app vars was returned")
	}
}
