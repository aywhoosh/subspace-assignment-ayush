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

// HumanClick performs a click with optional mouse movement simulation
func (h HumanBehavior) HumanClick(ctx context.Context, page *rod.Page, el *rod.Element) error {
	if h.EnableMouseMovement {
		// Get element position
		box, err := el.Shape()
		if err != nil {
			return err
		}

		// Calculate center of element using Quads
		quads := box.Quads
		if len(quads) > 0 {
			quad := quads[0]
			// Average of quad corners
			centerX := (quad[0] + quad[2] + quad[4] + quad[6]) / 4
			centerY := (quad[1] + quad[3] + quad[5] + quad[7]) / 4

			// Move mouse to element (Rod's MoveLinear)
			if err := page.Mouse.MoveLinear(proto.InputPoint{X: centerX, Y: centerY}, h.MouseSteps); err != nil {
				return err
			}
		}

		// Small random delay before clicking
		h.RandomDelay(ctx)
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
