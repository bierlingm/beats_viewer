package alert

import (
	"fmt"
	"time"

	"github.com/bierlingm/beats_viewer/pkg/attention"
	"github.com/bierlingm/beats_viewer/pkg/divergence"
	"github.com/bierlingm/beats_viewer/pkg/inference"
)

type AlertType int

const (
	AlertActivation AlertType = iota
	AlertEmergence
	AlertDormancy
	AlertDivergence
	AlertCrystallization
)

func (t AlertType) String() string {
	switch t {
	case AlertActivation:
		return "activation"
	case AlertEmergence:
		return "emergence"
	case AlertDormancy:
		return "dormancy"
	case AlertDivergence:
		return "divergence"
	case AlertCrystallization:
		return "crystallization"
	default:
		return "unknown"
	}
}

type AlertSeverity int

const (
	SeverityInfo AlertSeverity = iota
	SeverityNotable
	SeverityUrgent
)

func (s AlertSeverity) String() string {
	switch s {
	case SeverityInfo:
		return "info"
	case SeverityNotable:
		return "notable"
	case SeverityUrgent:
		return "urgent"
	default:
		return "unknown"
	}
}

type Alert struct {
	ID        string        `json:"id"`
	Type      AlertType     `json:"type"`
	Severity  AlertSeverity `json:"severity"`
	Title     string        `json:"title"`
	Detail    string        `json:"detail"`
	Actions   []AlertAction `json:"actions"`
	CreatedAt time.Time     `json:"created_at"`
	SeenAt    *time.Time    `json:"seen_at,omitempty"`
}

type AlertAction struct {
	Key   string `json:"key"`
	Label string `json:"label"`
	Cmd   string `json:"cmd"`
}

type AlertConfig struct {
	Enabled              bool          `json:"enabled"`
	ActivationThreshold  int           `json:"activation_threshold"`
	ActivationWindow     time.Duration `json:"activation_window"`
	DormancyThreshold    time.Duration `json:"dormancy_threshold"`
	DivergenceThreshold  int           `json:"divergence_threshold"`
	CrystallizationConf  float64       `json:"crystallization_conf"`
}

func DefaultConfig() AlertConfig {
	return AlertConfig{
		Enabled:              true,
		ActivationThreshold:  3,
		ActivationWindow:     72 * time.Hour,
		DormancyThreshold:    30 * 24 * time.Hour,
		DivergenceThreshold:  3,
		CrystallizationConf:  0.6,
	}
}

type Detector struct {
	config AlertConfig
}

func NewDetector(config AlertConfig) *Detector {
	return &Detector{config: config}
}

func (d *Detector) DetectAll(
	state *attention.AttentionState,
	crystallizations []inference.CrystallizationResult,
	divergenceReport *divergence.DivergenceReport,
) []Alert {
	if !d.config.Enabled {
		return nil
	}

	var alerts []Alert
	now := time.Now()

	alerts = append(alerts, d.detectActivationAlerts(state, now)...)
	alerts = append(alerts, d.detectEmergenceAlerts(state, now)...)
	alerts = append(alerts, d.detectDormancyAlerts(state, now)...)
	alerts = append(alerts, d.detectDivergenceAlerts(divergenceReport, now)...)
	alerts = append(alerts, d.detectCrystallizationAlerts(crystallizations, now)...)

	return alerts
}

func (d *Detector) detectActivationAlerts(state *attention.AttentionState, now time.Time) []Alert {
	if state == nil || len(state.Activations) == 0 {
		return nil
	}

	var alerts []Alert
	for _, act := range state.Activations {
		if act.BeatCount >= d.config.ActivationThreshold {
			severity := SeverityInfo
			if act.BeatCount >= d.config.ActivationThreshold*2 {
				severity = SeverityNotable
			}
			if act.BeatCount >= d.config.ActivationThreshold*3 {
				severity = SeverityUrgent
			}

			alerts = append(alerts, Alert{
				ID:        fmt.Sprintf("act-%s-%d", act.ClusterID, now.Unix()),
				Type:      AlertActivation,
				Severity:  severity,
				Title:     fmt.Sprintf("Activation: %s", act.ClusterName),
				Detail:    fmt.Sprintf("%d beats in %s", act.BeatCount, act.Window),
				CreatedAt: now,
				Actions: []AlertAction{
					{Key: "view", Label: "View Cluster", Cmd: fmt.Sprintf("cluster %s", act.ClusterID)},
				},
			})
		}
	}
	return alerts
}

func (d *Detector) detectEmergenceAlerts(state *attention.AttentionState, now time.Time) []Alert {
	if state == nil || len(state.Emergent) == 0 {
		return nil
	}

	var alerts []Alert
	for i, pattern := range state.Emergent {
		severity := SeverityInfo
		if pattern.BeatCount >= 5 {
			severity = SeverityNotable
		}

		alerts = append(alerts, Alert{
			ID:        fmt.Sprintf("emg-%d-%d", i, now.Unix()),
			Type:      AlertEmergence,
			Severity:  severity,
			Title:     "New pattern forming",
			Detail:    fmt.Sprintf("%d unclustered beats, signal: %s", pattern.BeatCount, pattern.Signal),
			CreatedAt: now,
			Actions: []AlertAction{
				{Key: "review", Label: "Review Beats", Cmd: "review unclustered"},
			},
		})
	}
	return alerts
}

func (d *Detector) detectDormancyAlerts(state *attention.AttentionState, now time.Time) []Alert {
	if state == nil || len(state.Dormant) == 0 {
		return nil
	}

	var alerts []Alert
	for _, dormant := range state.Dormant {
		inactiveDays := int(now.Sub(dormant.LastActivityAt).Hours() / 24)
		if time.Duration(inactiveDays)*24*time.Hour < d.config.DormancyThreshold {
			continue
		}

		severity := SeverityInfo
		if dormant.RipenessScore >= 0.8 {
			severity = SeverityNotable
		}
		if inactiveDays > 60 && dormant.RipenessScore >= 0.8 {
			severity = SeverityUrgent
		}

		alerts = append(alerts, Alert{
			ID:        fmt.Sprintf("dor-%s-%d", dormant.ClusterID, now.Unix()),
			Type:      AlertDormancy,
			Severity:  severity,
			Title:     fmt.Sprintf("Dormant: %s", dormant.ClusterName),
			Detail:    fmt.Sprintf("Ripe cluster inactive for %d days (%d ripe beats)", inactiveDays, dormant.RipeBeatCount),
			CreatedAt: now,
			Actions: []AlertAction{
				{Key: "view", Label: "View Cluster", Cmd: fmt.Sprintf("cluster %s", dormant.ClusterID)},
			},
		})
	}
	return alerts
}

func (d *Detector) detectDivergenceAlerts(report *divergence.DivergenceReport, now time.Time) []Alert {
	if report == nil || len(report.BlindSpots) == 0 {
		return nil
	}

	var alerts []Alert
	for _, blindSpot := range report.BlindSpots {
		for _, item := range report.AgentOnly {
			if item.Topic == blindSpot && item.AgentCount >= d.config.DivergenceThreshold {
				severity := SeverityNotable
				if item.AgentCount >= d.config.DivergenceThreshold*2 {
					severity = SeverityUrgent
				}

				alerts = append(alerts, Alert{
					ID:        fmt.Sprintf("div-%s-%d", blindSpot, now.Unix()),
					Type:      AlertDivergence,
					Severity:  severity,
					Title:     fmt.Sprintf("Blind spot: %s", blindSpot),
					Detail:    fmt.Sprintf("Agent captured %d beats, human 0", item.AgentCount),
					CreatedAt: now,
					Actions: []AlertAction{
						{Key: "review", Label: "Review Topic", Cmd: fmt.Sprintf("divergence %s", blindSpot)},
					},
				})
				break
			}
		}
	}
	return alerts
}

func (d *Detector) detectCrystallizationAlerts(crystallizations []inference.CrystallizationResult, now time.Time) []Alert {
	if len(crystallizations) == 0 {
		return nil
	}

	var alerts []Alert
	for _, cryst := range crystallizations {
		if cryst.Confidence >= d.config.CrystallizationConf {
			severity := SeverityInfo
			if cryst.Confidence >= 0.8 {
				severity = SeverityNotable
			}

			alerts = append(alerts, Alert{
				ID:        fmt.Sprintf("cry-%s-%d", cryst.BeadID, now.Unix()),
				Type:      AlertCrystallization,
				Severity:  severity,
				Title:     fmt.Sprintf("Crystallization: %s", cryst.BeadTitle),
				Detail:    fmt.Sprintf("%d beats linked (%.0f%% confidence)", len(cryst.BeatIDs), cryst.Confidence*100),
				CreatedAt: now,
				Actions: []AlertAction{
					{Key: "view", Label: "View Bead", Cmd: fmt.Sprintf("bead %s", cryst.BeadID)},
				},
			})
		}
	}
	return alerts
}
