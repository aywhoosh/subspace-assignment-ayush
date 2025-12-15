package browser

import (
	"context"
	"fmt"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/go-rod/rod/lib/proto"
)

type Config struct {
	Headless bool
	SlowMo   time.Duration
}

type Client struct {
	b *rod.Browser
}

func New(ctx context.Context, cfg Config) (*Client, func() error, error) {
	l := launcher.New().Headless(cfg.Headless)
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

	// Ensure Network domain is enabled so cookie APIs work reliably.
	if err = (proto.NetworkEnable{}).Call(b); err != nil {
		_ = b.Close()
		l.Cleanup()
		return nil, nil, fmt.Errorf("browser: enable network: %w", err)
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
	if err := p.Navigate(url); err != nil {
		_ = p.Close()
		return nil, fmt.Errorf("browser: navigate: %w", err)
	}
	return p, nil
}
