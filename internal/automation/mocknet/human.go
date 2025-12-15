package mocknet

import (
	"context"
	"math/rand"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/proto"
)

// HumanBehavior configures human-like interaction patterns
type HumanBehavior struct {
	// RandomDelay adds random pauses between actions
	RandomDelayMin time.Duration
	RandomDelayMax time.Duration

	// TypingSpeed controls character input speed
	TypingSpeedMin time.Duration
	TypingSpeedMax time.Duration

	// MouseMovement controls mouse behavior
	EnableMouseMovement bool
	MouseSteps          int
}

// DefaultHumanBehavior returns realistic default settings
func DefaultHumanBehavior() HumanBehavior {
	return HumanBehavior{
		RandomDelayMin:      500 * time.Millisecond,
		RandomDelayMax:      2000 * time.Millisecond,
		TypingSpeedMin:      50 * time.Millisecond,
		TypingSpeedMax:      150 * time.Millisecond,
		EnableMouseMovement: true,
		MouseSteps:          20,
	}
}

// FastHumanBehavior returns faster but still realistic settings
func FastHumanBehavior() HumanBehavior {
	return HumanBehavior{
		RandomDelayMin:      200 * time.Millisecond,
		RandomDelayMax:      800 * time.Millisecond,
		TypingSpeedMin:      20 * time.Millisecond,
		TypingSpeedMax:      80 * time.Millisecond,
		EnableMouseMovement: true,
		MouseSteps:          10,
	}
}

// RandomDelay waits for a random duration within configured bounds
func (h HumanBehavior) RandomDelay(ctx context.Context) {
	if h.RandomDelayMax == 0 {
		return
	}
	delay := h.RandomDelayMin + time.Duration(rand.Int63n(int64(h.RandomDelayMax-h.RandomDelayMin)))
	select {
	case <-time.After(delay):
	case <-ctx.Done():
	}
}

// HumanType simulates human-like typing with character-by-character input
func (h HumanBehavior) HumanType(ctx context.Context, el *rod.Element, text string) error {
	for _, char := range text {
		if err := el.Input(string(char)); err != nil {
			return err
		}
		if h.TypingSpeedMax > 0 {
			delay := h.TypingSpeedMin + time.Duration(rand.Int63n(int64(h.TypingSpeedMax-h.TypingSpeedMin)))
			select {
			case <-time.After(delay):
			case <-ctx.Done():
				return ctx.Err()
			}
		}
	}
	return nil
}

// HumanClick performs a click with optional delay simulation
func (h HumanBehavior) HumanClick(ctx context.Context, page *rod.Page, el *rod.Element) error {
	// Small random delay before clicking (simulates human thinking/targeting)
	if h.RandomDelayMax > 0 {
		delay := h.RandomDelayMin + time.Duration(rand.Int63n(int64(h.RandomDelayMax-h.RandomDelayMin)))
		select {
		case <-time.After(delay / 3): // Shorter delay for clicks
		case <-ctx.Done():
			return ctx.Err()
		}
	}

	// Perform the click
	return el.Click(proto.InputMouseButtonLeft, 1)
}

// ScrollRandom performs a random scroll to simulate reading
func (h HumanBehavior) ScrollRandom(ctx context.Context, page *rod.Page) error {
	// Random scroll amount (between 100 and 500 pixels)
	scrollAmount := 100 + rand.Intn(400)

	// Scroll down
	if err := page.Mouse.Scroll(0, float64(scrollAmount), h.MouseSteps); err != nil {
		return err
	}

	// Random delay while "reading"
	h.RandomDelay(ctx)

	return nil
}
