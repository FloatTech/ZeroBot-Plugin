package localadmin

import "net/http"

type OverviewAlerts struct {
	Entries      []LogEntry `json:"entries"`
	WarningCount int        `json:"warningCount"`
	ErrorCount   int        `json:"errorCount"`
	Retained     int        `json:"retained"`
	Evicted      bool       `json:"evicted"`
}

// recentAlerts reads only the existing bounded buffer, newest first. Counts
// describe retained records, not an inferred number of unresolved incidents.
func (l *LogStore) recentAlerts() *OverviewAlerts {
	l.mu.Lock()
	defer l.mu.Unlock()
	value := &OverviewAlerts{Entries: []LogEntry{}, Retained: l.count, Evicted: l.next > uint64(l.count)}
	for i := l.count - 1; i >= 0; i-- {
		row := l.rows[(l.start+i)%logCapacity]
		switch row.Level {
		case "warning":
			value.WarningCount++
		case "error", "fatal", "panic":
			value.ErrorCount++
		default:
			continue
		}
		if len(value.Entries) < 5 {
			row.Message = l.maskLocked(row.Message)
			text, clipped := clipLogMessage(row.Message)
			row.Message, row.Truncated = text, row.Truncated || clipped
			value.Entries = append(value.Entries, row)
		}
	}
	return value
}

func (s *Server) getOverview(w http.ResponseWriter, r *http.Request) {
	if r.Context().Err() != nil {
		return
	}
	value := s.cfg.Overview()
	value.ModelRequests = nil
	if s.cfg.ModelRequests != nil {
		stats := s.cfg.ModelRequests()
		// Only use the identity captured by the host. The statistics callback cannot
		// inject another executable, path or identity into authenticated reports.
		stats.Process = nil
		if s.cfg.ProcessIdentity != nil && s.cfg.ProcessIdentity.Valid() {
			identity := *s.cfg.ProcessIdentity
			stats.Process = &identity
		}
		value.ModelRequests = &stats
	}
	// Log failures must not turn a healthy transport into a disconnected one.
	value.ReplyTasks = nil
	if s.cfg.ReplyTasks != nil {
		stats := s.cfg.ReplyTasks()
		stats.Process = nil
		if s.cfg.ProcessIdentity != nil && s.cfg.ProcessIdentity.Valid() {
			identity := *s.cfg.ProcessIdentity
			stats.Process = &identity
		}
		value.ReplyTasks = &stats
	}
	// Do not return any log text until current configuration secrets are loaded.
	value.Alerts = nil
	if s.cfg.Logs != nil && s.refreshLogSecrets() == nil {
		value.Alerts = s.cfg.Logs.recentAlerts()
	}
	jsonReply(w, http.StatusOK, value)
}
