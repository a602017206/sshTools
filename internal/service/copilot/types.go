package copilot

// Result is the local danger classification for SQL or shell content.
type Result struct {
	Destructive bool
	Reason      string
}

// Artifact is a structured SQL or shell payload extracted from a model reply.
type Artifact struct {
	Type        string `json:"type"`
	Content     string `json:"content"`
	Summary     string `json:"summary"`
	Destructive bool   `json:"destructive"`
}
