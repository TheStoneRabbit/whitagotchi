package render

import (
	"fmt"
	"strings"

	"github.com/whitagotchi/whitagotchi/shared"
)

// Creature renders a creature plus its stats as multi-line ASCII (one frame).
func Creature(c *shared.Creature) string {
	return CreatureFrame(c, 0)
}

// CreatureFrame renders a specific animation frame. Sprite blinks AND sways
// side-to-side. The bottom (stats) is unaffected so layout stays stable.
func CreatureFrame(c *shared.Creature, frame int) string {
	var b strings.Builder
	b.WriteString(AnimatedFrame(c.Species, c.Stage, frame))
	b.WriteString("\n")
	fmt.Fprintf(&b, "%s the %s (%s, %s)\n", c.OwnerName, c.Species, c.Rarity, c.Stage)
	if c.Stage == shared.StageAdult && c.AdultForm != "" {
		fmt.Fprintf(&b, "adult form: %s\n", c.AdultForm)
	}
	fmt.Fprintf(&b, "quirk: %s\n", c.Quirk)
	b.WriteString("\n")
	fmt.Fprintf(&b, "  hunger      %s\n", Bar(c.Stats.Hunger))
	fmt.Fprintf(&b, "  happiness   %s\n", Bar(c.Stats.Happiness))
	fmt.Fprintf(&b, "  cleanliness %s\n", Bar(c.Stats.Cleanliness))
	return b.String()
}

// Bar renders a 0-100 stat as a 20-wide bar.
func Bar(v int) string {
	filled := v / 5
	if filled < 0 {
		filled = 0
	}
	if filled > 20 {
		filled = 20
	}
	return fmt.Sprintf("[%s%s] %d", strings.Repeat("#", filled), strings.Repeat(".", 20-filled), v)
}

// Frame returns the ASCII art for a species+stage at the given animation frame.
// Eyes blink occasionally. Use this for stable layouts (e.g. cards in a row).
func Frame(s shared.Species, st shared.Stage, frame int) string {
	base := art(s, st)
	if shouldBlink(frame) {
		base = blink(base)
	}
	return base
}

// AnimatedFrame is like Frame but also sways side-to-side over a cycle.
// Use this in spots where a single sprite has the screen to itself (the home
// view) — sway changes the rendered width frame-to-frame, which would shove
// adjacent panels around in multi-column layouts.
func AnimatedFrame(s shared.Species, st shared.Stage, frame int) string {
	return padLeft(Frame(s, st, frame), swayCycle[frame%len(swayCycle)])
}

// shouldBlink fires once every blinkPeriod frames at a steady offset, so the
// blink is uncorrelated with the sway phase and feels natural together.
const blinkPeriod = 10

func shouldBlink(frame int) bool { return frame%blinkPeriod == 4 }

// swayCycle is the extra leading spaces beyond baseline at each tick.
// Stays non-negative so we never have to trim into the actual sprite.
var swayCycle = []int{0, 1, 2, 3, 3, 2, 1, 0, 0, 1, 2, 3, 3, 2, 1, 0}

func padLeft(s string, n int) string {
	if n <= 0 {
		return s
	}
	pad := strings.Repeat(" ", n)
	var b strings.Builder
	for i, line := range strings.Split(s, "\n") {
		if i > 0 {
			b.WriteByte('\n')
		}
		if line == "" {
			continue
		}
		b.WriteString(pad)
		b.WriteString(line)
	}
	return b.String()
}

// blink swaps common eye glyphs for a closed/half-closed variant.
// It's a cheap way to get animation without authoring two full sprites per stage.
var blinkPairs = []struct{ open, closed string }{
	{"o.o", "-.-"},
	{"o o", "- -"},
	{"o_o", "-_-"},
	{"o  o", "-  -"},
	{"o   o", "-   -"},
	{"^.^", "-.-"},
	{". .", "- -"},
	{"o ^ o", "- ^ -"},
	{"@_@", "x_x"},
	{"@   @", "x   x"},
	{"@     @", "x     x"},
	{"(o.o)", "(-.-)"},
}

func blink(s string) string {
	out := s
	for _, p := range blinkPairs {
		out = strings.ReplaceAll(out, p.open, p.closed)
	}
	return out
}

func art(s shared.Species, st shared.Stage) string {
	if st == shared.StageEgg {
		return eggArt
	}
	if frames, ok := speciesArt[s]; ok {
		if a, ok := frames[st]; ok {
			return a
		}
	}
	return fallback
}
