package candlespan

const (
	Minute = "1m"
	Day    = "1d"

	dayMs = 24 * 60 * 60 * 1000

	minSpanForDaily = 2 * dayMs
)

type Segment struct {
	Interval string
	StartMs  int64
	EndMs    int64
}

func Split(startMs, endMs int64) []Segment {
	minutes := []Segment{{Interval: Minute, StartMs: startMs, EndMs: endMs}}

	if endMs <= startMs || endMs-startMs < minSpanForDaily {
		return minutes
	}

	firstMidnight := startMs - startMs%dayMs
	if startMs%dayMs != 0 {
		firstMidnight += dayMs
	}
	lastMidnight := endMs - endMs%dayMs
	if lastMidnight <= firstMidnight {
		return minutes
	}

	segments := make([]Segment, 0, 3)
	if startMs < firstMidnight {
		segments = append(segments, Segment{Interval: Minute, StartMs: startMs, EndMs: firstMidnight - 1})
	}
	segments = append(segments, Segment{Interval: Day, StartMs: firstMidnight, EndMs: lastMidnight - 1})
	if lastMidnight <= endMs {
		segments = append(segments, Segment{Interval: Minute, StartMs: lastMidnight, EndMs: endMs})
	}
	return segments
}
