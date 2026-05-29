package shared

// REST payloads.

type RegisterRequest struct {
	Username string `json:"username"`
}

type RegisterResponse struct {
	Username string `json:"username"`
	Token    string `json:"token"`
	Creature Creature `json:"creature"`
}

type StatusResponse struct {
	Creature Creature `json:"creature"`
}

// PeerInfoResponse describes another user's creature (read-only public view).
type PeerInfoResponse struct {
	Username  string    `json:"username"`
	Species   Species   `json:"species"`
	Rarity    Rarity    `json:"rarity"`
	Stage     Stage     `json:"stage"`
	AdultForm AdultForm `json:"adult_form,omitempty"`
	Quirk     Quirk     `json:"quirk"`
}

type ActionRequest struct {
	Action string `json:"action"` // feed, play, bathe
}

// WebSocket chat envelopes.

type ChatInbound struct {
	To      string `json:"to"`
	Message string `json:"message"`
}

type ChatOutbound struct {
	From            string `json:"from"`
	FromSpecies     Species `json:"from_species"`
	FromQuirk       Quirk   `json:"from_quirk"`
	OriginalMessage string  `json:"original_message,omitempty"` // only echoed back to sender
	Paraphrased     string  `json:"paraphrased"`
	At              string  `json:"at"` // RFC3339
}
