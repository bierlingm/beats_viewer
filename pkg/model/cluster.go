package model

import "time"

// Cluster represents a theme grouping of beats
type Cluster struct {
	ID            string    `json:"id"`
	Name          string    `json:"name"`
	BeatIDs       []string  `json:"beat_ids"`
	Centroid      []float64 `json:"centroid,omitempty"`
	Keywords      []string  `json:"keywords"`
	CreatedAt     time.Time `json:"created_at"`
	RipenessScore float64   `json:"ripeness_score,omitempty"`
}

// Chain represents an ordered sequence of related beats
type Chain struct {
	ID            string    `json:"id"`
	Name          string    `json:"name"`
	BeatIDs       []string  `json:"beat_ids"`
	CreatedAt     time.Time `json:"created_at"`
	RipenessScore float64   `json:"ripeness"`
}
