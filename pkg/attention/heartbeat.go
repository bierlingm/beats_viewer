package attention

import (
	"sort"
	"time"

	"github.com/bierlingm/beats_viewer/pkg/model"
)

// HeartbeatConfig holds analysis parameters
type HeartbeatConfig struct {
	Window      time.Duration // Default 90 days
	MinGapDays  int           // Minimum days to consider a gap (default 3)
	BurstThresh int           // Beats per day to consider a burst (default 3)
}

// DefaultHeartbeatConfig returns default heartbeat analysis settings
func DefaultHeartbeatConfig() HeartbeatConfig {
	return HeartbeatConfig{
		Window:      90 * 24 * time.Hour,
		MinGapDays:  3,
		BurstThresh: 3,
	}
}

// DayBucket represents captures on a single day
type DayBucket struct {
	Date  time.Time
	Count int
}

// Gap represents a period with no captures
type Gap struct {
	Start    time.Time
	End      time.Time
	Duration time.Duration
}

// Burst represents a period of high activity
type Burst struct {
	Start    time.Time
	End      time.Time
	Duration time.Duration
	Total    int // Total beats in the burst
}

// Heartbeat represents the temporal rhythm analysis
type Heartbeat struct {
	Window       time.Duration
	ComputedAt   time.Time
	TotalBeats   int
	DailyAverage float64
	CurrentRate  float64 // Rate in last 7 days
	RateChange   float64 // Percent change from average
	DailyBuckets []DayBucket
	Bursts       []Burst
	Gaps         []Gap
	LongestGap   Gap
}

// ComputeHeartbeat analyzes capture rhythm
func ComputeHeartbeat(beats []model.Beat, config HeartbeatConfig) *Heartbeat {
	now := time.Now()
	cutoff := now.Add(-config.Window)
	windowBeats := BeatsInWindow(beats, config.Window)

	h := &Heartbeat{
		Window:     config.Window,
		ComputedAt: now,
		TotalBeats: len(windowBeats),
	}

	// Build daily buckets
	h.DailyBuckets = buildDailyBuckets(windowBeats, cutoff, now)

	// Calculate daily average
	days := int(config.Window.Hours() / 24)
	if days > 0 {
		h.DailyAverage = float64(h.TotalBeats) / float64(days)
	}

	// Calculate current rate (last 7 days)
	last7Days := BeatsInWindow(beats, 7*24*time.Hour)
	h.CurrentRate = float64(len(last7Days)) / 7.0

	// Rate change percentage
	if h.DailyAverage > 0 {
		h.RateChange = ((h.CurrentRate - h.DailyAverage) / h.DailyAverage) * 100
	}

	// Find gaps and bursts
	h.Gaps = FindGaps(h.DailyBuckets, config.MinGapDays)
	h.Bursts = FindBursts(h.DailyBuckets, config.BurstThresh)

	// Find longest gap
	if len(h.Gaps) > 0 {
		h.LongestGap = h.Gaps[0]
		for _, g := range h.Gaps {
			if g.Duration > h.LongestGap.Duration {
				h.LongestGap = g
			}
		}
	}

	return h
}

// buildDailyBuckets creates a bucket for each day in the range
func buildDailyBuckets(beats []model.Beat, start, end time.Time) []DayBucket {
	// Normalize to day boundaries
	startDay := time.Date(start.Year(), start.Month(), start.Day(), 0, 0, 0, 0, start.Location())
	endDay := time.Date(end.Year(), end.Month(), end.Day(), 0, 0, 0, 0, end.Location())

	// Count beats per day
	dayCounts := make(map[string]int)
	for _, b := range beats {
		day := time.Date(b.CreatedAt.Year(), b.CreatedAt.Month(), b.CreatedAt.Day(), 0, 0, 0, 0, b.CreatedAt.Location())
		key := day.Format("2006-01-02")
		dayCounts[key]++
	}

	// Build buckets for every day
	var buckets []DayBucket
	for d := startDay; !d.After(endDay); d = d.AddDate(0, 0, 1) {
		key := d.Format("2006-01-02")
		buckets = append(buckets, DayBucket{
			Date:  d,
			Count: dayCounts[key],
		})
	}

	return buckets
}

// FindGaps identifies periods with no captures
func FindGaps(buckets []DayBucket, minGapDays int) []Gap {
	if len(buckets) == 0 {
		return nil
	}

	var gaps []Gap
	var gapStart *time.Time
	zeroDays := 0

	for _, bucket := range buckets {
		if bucket.Count == 0 {
			if gapStart == nil {
				start := bucket.Date
				gapStart = &start
			}
			zeroDays++
		} else {
			if gapStart != nil && zeroDays >= minGapDays {
				// End of gap (day before current bucket)
				end := bucket.Date.AddDate(0, 0, -1)
				gaps = append(gaps, Gap{
					Start:    *gapStart,
					End:      end,
					Duration: end.Sub(*gapStart) + 24*time.Hour, // Include both days
				})
			}
			gapStart = nil
			zeroDays = 0
		}
	}

	// Handle gap at end
	if gapStart != nil && zeroDays >= minGapDays {
		end := buckets[len(buckets)-1].Date
		gaps = append(gaps, Gap{
			Start:    *gapStart,
			End:      end,
			Duration: end.Sub(*gapStart) + 24*time.Hour,
		})
	}

	// Sort by duration descending
	sort.Slice(gaps, func(i, j int) bool {
		return gaps[i].Duration > gaps[j].Duration
	})

	return gaps
}

// FindBursts identifies periods of high activity
func FindBursts(buckets []DayBucket, minBeatsPerDay int) []Burst {
	if len(buckets) == 0 {
		return nil
	}

	var bursts []Burst
	var burstStart *time.Time
	burstTotal := 0
	burstDays := 0

	for _, bucket := range buckets {
		if bucket.Count >= minBeatsPerDay {
			if burstStart == nil {
				start := bucket.Date
				burstStart = &start
				burstTotal = 0
				burstDays = 0
			}
			burstTotal += bucket.Count
			burstDays++
		} else {
			if burstStart != nil && burstDays >= 2 {
				// End of burst (day before current bucket)
				end := bucket.Date.AddDate(0, 0, -1)
				bursts = append(bursts, Burst{
					Start:    *burstStart,
					End:      end,
					Duration: end.Sub(*burstStart) + 24*time.Hour,
					Total:    burstTotal,
				})
			}
			burstStart = nil
			burstTotal = 0
			burstDays = 0
		}
	}

	// Handle burst at end
	if burstStart != nil && burstDays >= 2 {
		end := buckets[len(buckets)-1].Date
		bursts = append(bursts, Burst{
			Start:    *burstStart,
			End:      end,
			Duration: end.Sub(*burstStart) + 24*time.Hour,
			Total:    burstTotal,
		})
	}

	// Sort by total descending
	sort.Slice(bursts, func(i, j int) bool {
		return bursts[i].Total > bursts[j].Total
	})

	return bursts
}
