package scalingo

import (
	"testing"
)

func assertEqual(t *testing.T, name string, a interface{}, b interface{}) {
	if a != b {
		t.Fatalf("%s: %s != %s", name, a, b)
	}
}

func TestPrepareRequestWithoutPrefix(t *testing.T) {
	s := NewScalingo("test")
	req := s.PrepareRequest("POST", "", "/v1/apps/foo/domains", "", nil)
	assertEqual(t, "URL", req.URL.String(), "https://api.scalingo.com/v1/apps/foo/domains")
}

func TestPrepareRequestWithPrefix(t *testing.T) {
	s := NewScalingo("test")
	req := s.PrepareRequest("POST", "https://foo.bar", "/v1/apps/foo/domains", "", nil)
	assertEqual(t, "URL without Prefix", req.URL.String(), "https://foo.bar/v1/apps/foo/domains")
}

func TestPrepareRequestForURLWithoutToken(t *testing.T) {
	s := NewScalingo("test")
	req := s.PrepareRequestForURL("POST", "http://foo.bar/v1/apps/foo/domains", "", nil)
	assertEqual(t, "Authorization", len(req.Header["Authorization"]), 0)
}

func TestPrepareRequestForURLWithToken(t *testing.T) {
	s := NewScalingo("test")
	req := s.PrepareRequestForURL("POST", "http://foo.bar/v1/apps/foo/domains", "test", nil)
	assertEqual(t, "Authorization", req.Header["Authorization"][0], "Basic OnRlc3Q=")
}
