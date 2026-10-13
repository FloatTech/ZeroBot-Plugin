module github.com/FloatTech/sqlite

go 1.20

require (
	github.com/FloatTech/ttl v0.0.0-20240716161252-965925764562
	modernc.org/sqlite v1.33.1
)

require (
	github.com/dustin/go-humanize v1.0.1 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/ncruces/go-strftime v0.1.9 // indirect
	github.com/remyoudompheng/bigfft v0.0.0-20230129092748-24d4a6f8daec // indirect
	golang.org/x/sys v0.26.0 // indirect
	modernc.org/libc v1.61.0 // indirect
	modernc.org/mathutil v1.6.0 // indirect
	modernc.org/memory v1.8.0 // indirect
)

replace modernc.org/sqlite => github.com/fumiama/sqlite3 v1.29.10-simp

replace modernc.org/libc => github.com/fumiama/libc v0.0.0-20240530081950-6f6d8586b5c5
