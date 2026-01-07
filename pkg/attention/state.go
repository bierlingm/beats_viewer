package attention

import "time"

// AttentionState holds all computed attention analysis
type AttentionState struct {
	ComputedAt  time.Time          `json:"computed_at"`
	Activations []Activation       `json:"activations"`
	DriftReport *DriftReport       `json:"drift_report"`
	Orientation *OrientationSummary `json:"orientation"`
	Heartbeat   *Heartbeat         `json:"heartbeat"`
	Dormant     []DormantCluster   `json:"dormant"`
	Emergent    []EmergentPattern  `json:"emergent"`
}

// NewAttentionState creates an empty attention state
func NewAttentionState() *AttentionState {
	return &AttentionState{
		ComputedAt:  time.Now(),
		Activations: []Activation{},
		Dormant:     []DormantCluster{},
		Emergent:    []EmergentPattern{},
	}
}

// HasActivations returns true if there are any activations
func (s *AttentionState) HasActivations() bool {
	return s != nil && len(s.Activations) > 0
}

// HasDormant returns true if there are dormant clusters
func (s *AttentionState) HasDormant() bool {
	return s != nil && len(s.Dormant) > 0
}

// HasEmergent returns true if there are emergent patterns
func (s *AttentionState) HasEmergent() bool {
	return s != nil && len(s.Emergent) > 0
}

// IsStale returns true if the state is older than the given duration
func (s *AttentionState) IsStale(maxAge time.Duration) bool {
	if s == nil {
		return true
	}
	return time.Since(s.ComputedAt) > maxAge
}
