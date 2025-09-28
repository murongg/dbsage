package ai

// ToolConfirmationConfig defines which tools need confirmation (copied from cli package to avoid circular import)
type ToolConfirmationConfig struct {
	RequiresConfirmation map[string]bool   `json:"requires_confirmation"`
	RiskLevels           map[string]string `json:"risk_levels"`
	Descriptions         map[string]string `json:"descriptions"`
}

// NewToolConfirmationConfig creates a new tool confirmation config
func NewToolConfirmationConfig(requiresConfirmation map[string]bool, riskLevels map[string]string, descriptions map[string]string) *ToolConfirmationConfig {
	return &ToolConfirmationConfig{
		RequiresConfirmation: requiresConfirmation,
		RiskLevels:           riskLevels,
		Descriptions:         descriptions,
	}
}
