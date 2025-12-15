package browser

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/go-rod/rod/lib/proto"
)

type Config struct {
	Headless      bool
	SlowMo        time.Duration
	Leakless      bool
	BinPath       string
	AllowDownload bool
}

type Client struct {
	b *rod.Browser
}

func New(ctx context.Context, cfg Config) (*Client, func() error, error) {
	l := launcher.New().Headless(cfg.Headless).Leakless(cfg.Leakless)

	r, err := resolveBrowserBin(cfg)
	if err != nil {
		return nil, nil, err
	}
	if r.Path != "" {
		l = l.Bin(r.Path)
	} else if !cfg.AllowDownload {
		return nil, nil, errors.New(
			"browser: no system browser found. Set browser.bin_path (env SUBSPACE_BROWSER__BIN_PATH) to Chrome/Edge, or set browser.allow_download=true to let Rod download Chromium",
		)
	}

	controlURL, err := l.Launch()
	if err != nil {
		return nil, nil, fmt.Errorf("browser: launch: %w", err)
	}

	b := rod.New().ControlURL(controlURL)
	if cfg.SlowMo > 0 {
		b = b.SlowMotion(cfg.SlowMo)
	}

	b = b.Context(ctx)
	if err := b.Connect(); err != nil {
		l.Cleanup()
		return nil, nil, fmt.Errorf("browser: connect: %w", err)
	}

	cl := &Client{b: b}
	cleanup := func() error {
		_ = b.Close()
		l.Cleanup()
		return nil
	}
	return cl, cleanup, nil
}

func (c *Client) Browser() *rod.Browser {
	return c.b
}

func (c *Client) NewPage(url string) (*rod.Page, error) {
	p, err := c.b.Page(proto.TargetCreateTarget{URL: "about:blank"})
	if err != nil {
		return nil, fmt.Errorf("browser: new page: %w", err)
	}
	if err := (proto.NetworkEnable{}).Call(p); err != nil {
		_ = p.Close()
		return nil, fmt.Errorf("browser: enable network: %w", err)
	}
	if err := p.Navigate(url); err != nil {
		_ = p.Close()
		return nil, fmt.Errorf("browser: navigate: %w", err)
	}
	return p, nil
}
