package localadmin

// StartupDifference exposes only fixed categories and aggregate counts. The
// comparison includes secrets, but neither their values nor hashes leave it.
type StartupDifference struct {
	Label        string `json:"label"`
	Changed      bool   `json:"changed"`
	RunningCount *int   `json:"runningCount,omitempty"`
	SavedCount   *int   `json:"savedCount,omitempty"`
	Unit         string `json:"unit,omitempty"`
}

func startupDifferences(running, saved []byte) []StartupDifference {
	// Both inputs have already passed decodeConfig. Use the same comparison as
	// restore previews to retain number precision and representation differences.
	impacts, err := fileRestoreImpacts("startup-config", running, saved)
	if err != nil {
		return nil
	}
	items := make([]StartupDifference, 0, len(impacts))
	for _, v := range impacts {
		items = append(items, StartupDifference{Label: v.Label, Changed: v.Changed, RunningCount: v.CurrentCount, SavedCount: v.BackupCount, Unit: v.Unit})
	}
	return items
}
