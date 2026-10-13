# Local Agent history patch

Baseline: `github.com/fumiama/go-onebot-agent` at `9c92a03776e3` (2026-04-25).
The original LICENSE and embedded prompts/permissions are retained.

An optional `HistoryStore` receives events, action attempts, results and end
markers. When enabled, prompt reconstruction reads that store instead of the
upstream 16-batch/8-item in-memory ring. Unmodified users retain the original
ring. Record methods return errors so execution can fail before an unrecorded
action. The image preprocessing hook also applies to restored typed events.

The application supplies private SQLite storage, conversation isolation,
redaction, bounded model input and retention without a turn/count limit. The
upstream action format, JSON decoder and permission table remain unchanged.
