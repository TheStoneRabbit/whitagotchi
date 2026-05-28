package game

import (
	"log"
	"time"

	"github.com/whitagotchi/whitagotchi/server/internal/store"
	"github.com/whitagotchi/whitagotchi/shared"
)

type Config struct {
	// PaceMultiplier scales how fast time passes. 1.0 = default fast pace
	// (egg -> adult in ~1-2 hours). 0.1 = ten times slower.
	PaceMultiplier float64
	TickInterval   time.Duration
}

type Engine struct {
	store *store.Store
	cfg   Config
}

func NewEngine(s *store.Store, cfg Config) *Engine {
	if cfg.TickInterval <= 0 {
		cfg.TickInterval = 30 * time.Second
	}
	if cfg.PaceMultiplier <= 0 {
		cfg.PaceMultiplier = 1.0
	}
	return &Engine{store: s, cfg: cfg}
}

// Run is the authoritative game loop. It decays stats and advances stages
// even when no clients are connected.
func (e *Engine) Run() {
	t := time.NewTicker(e.cfg.TickInterval)
	defer t.Stop()
	for range t.C {
		for _, c := range e.store.AllCreatures() {
			e.advance(c, time.Now().UTC())
			e.store.UpdateCreature(c)
		}
	}
}

// AdvanceNow brings a creature forward to the current wall clock.
// Used both by the ticker and on demand (e.g. when a user views status).
func (e *Engine) AdvanceNow(c *shared.Creature) {
	e.advance(c, time.Now().UTC())
}

func (e *Engine) advance(c *shared.Creature, now time.Time) {
	elapsed := now.Sub(c.LastTickAt).Seconds() * e.cfg.PaceMultiplier
	if elapsed <= 0 {
		return
	}

	// Stat decay: ~1 point per minute of scaled time.
	dec := int(elapsed / 60)
	if dec > 0 {
		c.Stats.Hunger = clamp(c.Stats.Hunger - dec)
		c.Stats.Happiness = clamp(c.Stats.Happiness - dec)
		c.Stats.Cleanliness = clamp(c.Stats.Cleanliness - dec)
	}

	// Stage progression based on scaled age.
	scaledAge := now.Sub(c.HatchedAt).Seconds() * e.cfg.PaceMultiplier
	newStage := stageForAge(scaledAge)
	if newStage != c.Stage {
		log.Printf("creature %s (%s) advanced %s -> %s", c.ID, c.OwnerName, c.Stage, newStage)
		c.Stage = newStage
		if newStage == shared.StageAdult {
			c.AdultForm = adultFormFor(c.Stats)
		}
	}

	c.LastTickAt = now
}

func stageForAge(seconds float64) shared.Stage {
	switch {
	case seconds < 10*60: // 0-10m: egg
		return shared.StageEgg
	case seconds < 30*60: // 10-30m: baby
		return shared.StageBaby
	case seconds < 75*60: // 30-75m: teen
		return shared.StageTeen
	default:
		return shared.StageAdult
	}
}

func adultFormFor(s shared.Stats) shared.AdultForm {
	avg := (s.Hunger + s.Happiness + s.Cleanliness) / 3
	switch {
	case avg >= 70:
		return shared.AdultThriving
	case avg >= 40:
		return shared.AdultNeutral
	default:
		return shared.AdultNeglect
	}
}

func clamp(v int) int {
	if v < 0 {
		return 0
	}
	if v > 100 {
		return 100
	}
	return v
}

// ApplyAction updates stats for a user-initiated action.
func ApplyAction(c *shared.Creature, action string) {
	switch action {
	case "feed":
		c.Stats.Hunger = clamp(c.Stats.Hunger + 30)
	case "play":
		c.Stats.Happiness = clamp(c.Stats.Happiness + 30)
	case "bathe":
		c.Stats.Cleanliness = clamp(c.Stats.Cleanliness + 30)
	}
}
