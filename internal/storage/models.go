package storage

type ContinueEntry struct {
	HistoryEntry
	Percent float64 `json:"percent"`
}

func (e *ContinueEntry) UpdatePercent() {
	if e.Duration > 0 {
		e.Percent = float64(e.Position) / float64(e.Duration) * 100
	}
}
