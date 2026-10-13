# sqlite
A pure-go wrapper for database/sql using logoove/sqlite

## Usage

### 1. Replace modified sqlite lib in your go.mod
```bash
replace modernc.org/sqlite => github.com/fumiama/sqlite3 v1.29.10-simp

replace modernc.org/libc => github.com/fumiama/libc v0.0.0-20240530081950-6f6d8586b5c5
```

### 2. Use it
```go
type row struct {
    ID   int // pk
    Name string
}

db := &sql.Sqlite{DBPath: "demo.db"}
err := db.Open(time.Hour)
if err != nil {
    panic(err)
}

err = db.Create("demotable", &row{})
if err != nil {
    panic(err)
}

err = db.Insert("demotable", &row{ID: 123, Name: "Anna"})
if err != nil {
    panic(err)
}

var r row
err = db.Find("demotable", &r, "WHERE ID=123")
if err != nil {
    panic(err)
}
fmt.Println(r)
```
