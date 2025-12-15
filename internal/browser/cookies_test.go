package browser_test

import (
	"encoding/json"
	"testing"

	"github.com/go-rod/rod/lib/proto"
)

func TestCookieJSONMarshaling(t *testing.T) {
	// Test that NetworkCookieParam can be marshaled to JSON
	cookie := &proto.NetworkCookieParam{
		Name:     "session",
		Value:    "abc123",
		Domain:   "example.com",
		Path:     "/",
		Secure:   true,
		HTTPOnly: true,
		SameSite: proto.NetworkCookieSameSiteStrict,
	}

	data, err := json.Marshal(cookie)
	if err != nil {
		t.Fatalf("failed to marshal cookie: %v", err)
	}

	var decoded proto.NetworkCookieParam
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal cookie: %v", err)
	}

	if decoded.Name != cookie.Name || decoded.Value != cookie.Value {
		t.Error("Cookie JSON roundtrip failed: name or value mismatch")
	}
}

func TestCookieFields(t *testing.T) {
	// Verify cookie structure has expected fields
	cookie := &proto.NetworkCookie{
		Name:     "test",
		Value:    "value123",
		Domain:   "test.com",
		Path:     "/path",
		Secure:   false,
		HTTPOnly: false,
		SameSite: proto.NetworkCookieSameSiteLax,
	}

	if cookie.Name != "test" {
		t.Errorf("expected name=test, got %s", cookie.Name)
	}
	if cookie.Value != "value123" {
		t.Errorf("expected value=value123, got %s", cookie.Value)
	}
	if cookie.Domain != "test.com" {
		t.Errorf("expected domain=test.com, got %s", cookie.Domain)
	}
}
