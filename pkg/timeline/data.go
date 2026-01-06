package timeline

import (
	"sort"
	"time"

	"github.com/bierlingm/beats_viewer/pkg/model"
)

type ZoomLevel int

const (
	ZoomDay ZoomLevel = iota
	ZoomWeek
	ZoomMonth
	ZoomQuarter
)

func (z ZoomLevel) String() string {
	switch z {
	case ZoomDay:
		return "Day"
	case ZoomWeek:
		return "Week"
	case ZoomMonth:
		return "Month"
	case ZoomQuarter:
		return "Quarter"
	default:
		return "Unknown"
	}
}

type TimelineBucket struct {
	Date      time.Time
	BeatCount int
	BeatIDs   []string
	ByChannel map[model.Channel]int
}

type TimelineData struct {
	Start     time.Time
	End       time.Time
	Buckets   []TimelineBucket
	ZoomLevel ZoomLevel
}

func BuildTimeline(beats []model.EnrichedBeat, zoom ZoomLevel) *TimelineData {
	if len(beats) == 0 {
		return &TimelineData{ZoomLevel: zoom}
	}

	sorted := make([]model.EnrichedBeat, len(beats))
	copy(sorted, beats)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].CreatedAt.Before(sorted[j].CreatedAt)
	})

	start := sorted[0].CreatedAt
	end := sorted[len(sorted)-1].CreatedAt

	bucketMap := make(map[string]*TimelineBucket)

	for _, beat := range sorted {
		key := bucketKey(beat.CreatedAt, zoom)
		bucket, ok := bucketMap[key]
		if !ok {
			bucket = &TimelineBucket{
				Date:      truncateToZoom(beat.CreatedAt, zoom),
				ByChannel: make(map[model.Channel]int),
			}
			bucketMap[key] = bucket
		}
		bucket.BeatCount++
		bucket.BeatIDs = append(bucket.BeatIDs, beat.ID)
		bucket.ByChannel[beat.Taxonomy.Channel]++
	}

	var buckets []TimelineBucket
	for _, b := range bucketMap {
		buckets = append(buckets, *b)
	}
	sort.Slice(buckets, func(i, j int) bool {
		return buckets[i].Date.Before(buckets[j].Date)
	})

	return &TimelineData{
		Start:     start,
		End:       end,
		Buckets:   buckets,
		ZoomLevel: zoom,
	}
}

func bucketKey(t time.Time, zoom ZoomLevel) string {
	switch zoom {
	case ZoomDay:
		return t.Format("2006-01-02")
	case ZoomWeek:
		year, week := t.ISOWeek()
		return t.Format("2006") + "-W" + string(rune('0'+week/10)) + string(rune('0'+week%10)) + "-" + string(rune('0'+year%10))
	case ZoomMonth:
		return t.Format("2006-01")
	case ZoomQuarter:
		q := (t.Month()-1)/3 + 1
		return t.Format("2006") + "-Q" + string(rune('0'+q))
	}
	return t.Format("2006-01-02")
}

func truncateToZoom(t time.Time, zoom ZoomLevel) time.Time {
	switch zoom {
	case ZoomDay:
		return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
	case ZoomWeek:
		year, week := t.ISOWeek()
		jan1 := time.Date(year, 1, 1, 0, 0, 0, 0, t.Location())
		weekday := int(jan1.Weekday())
		if weekday == 0 {
			weekday = 7
		}
		firstMonday := jan1.AddDate(0, 0, -weekday+1)
		return firstMonday.AddDate(0, 0, (week-1)*7)
	case ZoomMonth:
		return time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, t.Location())
	case ZoomQuarter:
		q := (t.Month()-1)/3 + 1
		return time.Date(t.Year(), time.Month((q-1)*3+1), 1, 0, 0, 0, 0, t.Location())
	}
	return t
}

func (td *TimelineData) MaxBeatCount() int {
	max := 0
	for _, b := range td.Buckets {
		if b.BeatCount > max {
			max = b.BeatCount
		}
	}
	return max
}

func (td *TimelineData) GetBucketAt(index int) *TimelineBucket {
	if index >= 0 && index < len(td.Buckets) {
		return &td.Buckets[index]
	}
	return nil
}

func (td *TimelineData) FindGaps(threshold time.Duration) []struct{ Start, End time.Time } {
	var gaps []struct{ Start, End time.Time }

	for i := 1; i < len(td.Buckets); i++ {
		prev := td.Buckets[i-1].Date
		curr := td.Buckets[i].Date
		diff := curr.Sub(prev)
		if diff > threshold {
			gaps = append(gaps, struct{ Start, End time.Time }{prev, curr})
		}
	}

	return gaps
}
