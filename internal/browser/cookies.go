package browser

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/proto"
)

type cookieJSON struct {
	Name     string  `json:"name"`
	Value    string  `json:"value"`
	Domain   string  `json:"domain"`
	Path     string  `json:"path"`
	Expires  float64 `json:"expires"`
	HTTPOnly bool    `json:"http_only"`
	Secure   bool    `json:"secure"`
	SameSite string  `json:"same_site"`
}

func CookiesToJSON(cookies []*proto.NetworkCookie) (string, error) {
	out := make([]cookieJSON, 0, len(cookies))
	for _, c := range cookies {
		if c == nil {
			continue
		}
		out = append(out, cookieJSON{
			Name:     c.Name,
			Value:    c.Value,
			Domain:   c.Domain,
			Path:     c.Path,
			Expires:  float64(c.Expires),
			HTTPOnly: c.HTTPOnly,
			Secure:   c.Secure,
			SameSite: string(c.SameSite),
		})
	}
	b, err := json.Marshal(out)
	if err != nil {
		return "", fmt.Errorf("browser: marshal cookies: %w", err)
	}
	return string(b), nil
}

func CookiesFromJSON(s string) ([]*proto.NetworkCookieParam, error) {
	var in []cookieJSON
	if err := json.Unmarshal([]byte(s), &in); err != nil {
		return nil, fmt.Errorf("browser: unmarshal cookies: %w", err)
	}
	out := make([]*proto.NetworkCookieParam, 0, len(in))
	for _, c := range in {
		ss := proto.NetworkCookieSameSite(c.SameSite)
		out = append(out, &proto.NetworkCookieParam{
			Name:     c.Name,
			Value:    c.Value,
			Domain:   c.Domain,
			Path:     c.Path,
			Expires:  proto.TimeSinceEpoch(c.Expires),
			HTTPOnly: c.HTTPOnly,
			Secure:   c.Secure,
			SameSite: ss,
		})
	}
	return out, nil
}

func GetAllCookies(ctx context.Context, b *rod.Browser) ([]*proto.NetworkCookie, error) {
	// Use Storage domain which is always available
	res, err := proto.StorageGetCookies{}.Call(b)
	if err != nil {
		return nil, fmt.Errorf("browser: get cookies: %w", err)
	}
	return res.Cookies, nil
}

func SetCookies(ctx context.Context, b *rod.Browser, cookies []*proto.NetworkCookieParam) error {
	if len(cookies) == 0 {
		return nil
	}
	// Use Storage domain to set cookies
	for _, c := range cookies {
		err := proto.StorageSetCookies{Cookies: []*proto.NetworkCookieParam{c}}.Call(b)
		if err != nil {
			return fmt.Errorf("browser: set cookie %s: %w", c.Name, err)
		}
	}
	return nil
}
