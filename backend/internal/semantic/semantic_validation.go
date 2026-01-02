package semantic

type SemanticValidationResult struct {
	Domain          string   `json:"domain"`
	Valid           bool     `json:"valid"`
	MissingSemantic []string `json:"missing_semantic"`
	PresentSemantic []string `json:"present_semantic"`
	Notes           string   `json:"notes"`
}
