package render

import (
	"fmt"
	"strings"

	"github.com/whitagotchi/whitagotchi/shared"
)

// Creature renders a creature plus its stats as multi-line ASCII.
func Creature(c *shared.Creature) string {
	var b strings.Builder
	b.WriteString(art(c.Species, c.Stage))
	b.WriteString("\n")
	fmt.Fprintf(&b, "%s the %s (%s, %s)\n", c.OwnerName, c.Species, c.Rarity, c.Stage)
	if c.Stage == shared.StageAdult && c.AdultForm != "" {
		fmt.Fprintf(&b, "adult form: %s\n", c.AdultForm)
	}
	fmt.Fprintf(&b, "quirk: %s\n", c.Quirk)
	b.WriteString("\n")
	fmt.Fprintf(&b, "  hunger      %s\n", bar(c.Stats.Hunger))
	fmt.Fprintf(&b, "  happiness   %s\n", bar(c.Stats.Happiness))
	fmt.Fprintf(&b, "  cleanliness %s\n", bar(c.Stats.Cleanliness))
	return b.String()
}

func bar(v int) string {
	filled := v / 5 // 0-20
	return fmt.Sprintf("[%s%s] %d", strings.Repeat("#", filled), strings.Repeat(".", 20-filled), v)
}

// art is a placeholder lookup; real frames live in art.go.
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
