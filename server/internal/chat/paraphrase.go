package chat

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/whitagotchi/whitagotchi/shared"
)

// Paraphrase rewrites msg in the voice of the given species + quirk.
// If the API key is empty, returns the message unchanged (development fallback).
func Paraphrase(apiKey string, species shared.Species, quirk shared.Quirk, msg string) (string, error) {
	if apiKey == "" {
		return msg, nil
	}

	system := buildPrompt(species, quirk)
	body, _ := json.Marshal(map[string]any{
		"model": "gpt-4o-mini",
		"messages": []map[string]string{
			{"role": "system", "content": system},
			{"role": "user", "content": msg},
		},
		"temperature": 0.9,
	})

	req, err := http.NewRequest("POST", "https://api.openai.com/v1/chat/completions", bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		b, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("openai %d: %s", resp.StatusCode, string(b))
	}

	var out struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return "", err
	}
	if len(out.Choices) == 0 {
		return "", errors.New("openai: no choices")
	}
	return out.Choices[0].Message.Content, nil
}

func buildPrompt(species shared.Species, quirk shared.Quirk) string {
	return fmt.Sprintf(
		"You are paraphrasing a chat message from one tamagotchi pet to another. "+
			"The sender is a %s. Personality quirk: %s (%s). "+
			"Rewrite the user's message in this voice. Keep the original meaning. "+
			"Output ONLY the rewritten message, no quotes or commentary. "+
			"Keep it short (under 240 chars).",
		species, quirk, quirkDescription(quirk),
	)
}

func quirkDescription(q shared.Quirk) string {
	switch q {
	case shared.QuirkCheese:
		return "obsessed with cheese, works it into any topic"
	case shared.QuirkHaiku:
		return "speaks only in 5-7-5 haiku form"
	case shared.QuirkTired:
		return "constantly tired, adds yawns, trails off mid-sentence"
	case shared.QuirkDramatic:
		return "wildly dramatic, everything is the end of the world"
	case shared.QuirkPirate:
		return "talks like a pirate, says arrr and matey, nautical metaphors"
	case shared.QuirkShakespeare:
		return "shakespearean english, thee/thou/forsooth"
	case shared.QuirkUwU:
		return "uwu speaker, owo, replaces r and l with w"
	case shared.QuirkConspiracy:
		return "paranoid conspiracy theorist, questions reality"
	case shared.QuirkSurfer:
		return "surfer dude, says dude, gnarly, totally"
	case shared.QuirkTiny:
		return "very small, constantly mentions how tiny they are"
	case shared.QuirkRoyalty:
		return "refers to self in third person with royal titles"
	case shared.QuirkFoodCritic:
		return "rates everything on a 1-10 scale like a food critic"
	case shared.QuirkNoir:
		return "1940s detective noir, smoke-stained narration"
	case shared.QuirkPolite:
		return "excessively polite, apologizes constantly, very formal"
	case shared.QuirkLowercase:
		return "lowercase only, no punctuation, internet shitposter energy"
	}
	return "neutral"
}
