// Package sql 数据库/数据处理相关工具
package sql

import (
	"database/sql"
	"errors"
	"reflect"
	"strings"

	_ "github.com/logoove/sqlite" // 引入sqlite
)

// Sqlite 数据库对象
type Sqlite struct {
	DB     *sql.DB
	DBPath string
}

// Open 打开数据库
func (db *Sqlite) Open() (err error) {
	if db.DB == nil {
		database, err := sql.Open("sqlite3", db.DBPath)
		if err != nil {
			return err
		}
		db.DB = database
	}
	return
}

// Close 关闭数据库
func (db *Sqlite) Close() (err error) {
	if db.DB != nil {
		err = db.DB.Close()
		db.DB = nil
	}
	return
}

// Create 生成数据库
// 默认结构体的第一个元素为主键
// 返回错误
func (db *Sqlite) Create(table string, objptr interface{}) (err error) {
	if db.DB == nil {
		database, err := sql.Open("sqlite3", db.DBPath)
		if err != nil {
			return err
		}
		db.DB = database
	}
	var (
		tags  = tags(objptr)
		kinds = kinds(objptr)
		top   = len(tags) - 1
		cmd   = []string{}
	)
	cmd = append(cmd, "CREATE TABLE IF NOT EXISTS")
	cmd = append(cmd, table)
	cmd = append(cmd, "(")
	for i := range tags {
		cmd = append(cmd, tags[i])
		cmd = append(cmd, kinds[i])
		switch i {
		default:
			cmd = append(cmd, "NULL,")
		case 0:
			cmd = append(cmd, "PRIMARY KEY")
			cmd = append(cmd, "NOT NULL,")
		case top:
			cmd = append(cmd, "NULL);")
		}
	}
	_, err = db.DB.Exec(strings.Join(cmd, " ") + ";")
	return
}

// Insert 插入数据集
// 如果 PK 存在会覆盖
// 默认结构体的第一个元素为主键
// 返回错误
func (db *Sqlite) Insert(table string, objptr interface{}) error {
	rows, err := db.DB.Query("SELECT * FROM " + table + " limit 1;")
	if err != nil {
		return err
	}
	if rows.Err() != nil {
		return rows.Err()
	}
	tags, _ := rows.Columns()
	rows.Close()
	var (
		vals = values(objptr)
		top  = len(tags) - 1
		cmd  = []string{}
	)
	cmd = append(cmd, "REPLACE INTO")
	cmd = append(cmd, table)
	for i := range tags {
		switch i {
		default:
			cmd = append(cmd, tags[i])
			cmd = append(cmd, ",")
		case 0:
			cmd = append(cmd, "(")
			cmd = append(cmd, tags[i])
			cmd = append(cmd, ",")
		case top:
			cmd = append(cmd, tags[i])
			cmd = append(cmd, ")")
		}
	}
	for i := range tags {
		switch i {
		default:
			cmd = append(cmd, "?")
			cmd = append(cmd, ",")
		case 0:
			cmd = append(cmd, "VALUES (")
			cmd = append(cmd, "?")
			cmd = append(cmd, ",")
		case top:
			cmd = append(cmd, "?")
			cmd = append(cmd, ")")
		}
	}
	stmt, err := db.DB.Prepare(strings.Join(cmd, " ") + ";")
	if err != nil {
		return err
	}
	_, err = stmt.Exec(vals...)
	if err != nil {
		return err
	}
	return stmt.Close()
}

// InsertUnique 插入数据集
// 如果 PK 存在会报错
// 默认结构体的第一个元素为主键
// 返回错误
func (db *Sqlite) InsertUnique(table string, objptr interface{}) error {
	rows, err := db.DB.Query("SELECT * FROM " + table + " limit 1;")
	if err != nil {
		return err
	}
	if rows.Err() != nil {
		return rows.Err()
	}
	tags, _ := rows.Columns()
	rows.Close()
	var (
		vals = values(objptr)
		top  = len(tags) - 1
		cmd  = []string{}
	)
	cmd = append(cmd, "INSERT INTO")
	cmd = append(cmd, table)
	for i := range tags {
		switch i {
		default:
			cmd = append(cmd, tags[i])
			cmd = append(cmd, ",")
		case 0:
			cmd = append(cmd, "(")
			cmd = append(cmd, tags[i])
			cmd = append(cmd, ",")
		case top:
			cmd = append(cmd, tags[i])
			cmd = append(cmd, ")")
		}
	}
	for i := range tags {
		switch i {
		default:
			cmd = append(cmd, "?")
			cmd = append(cmd, ",")
		case 0:
			cmd = append(cmd, "VALUES (")
			cmd = append(cmd, "?")
			cmd = append(cmd, ",")
		case top:
			cmd = append(cmd, "?")
			cmd = append(cmd, ")")
		}
	}
	stmt, err := db.DB.Prepare(strings.Join(cmd, " ") + ";")
	if err != nil {
		return err
	}
	_, err = stmt.Exec(vals...)
	if err != nil {
		return err
	}
	return stmt.Close()
}

// Find 查询数据库，写入最后一条结果到 objptr
// condition 可为"WHERE id = 0"
// 默认字段与结构体元素顺序一致
// 返回错误
func (db *Sqlite) Find(table string, objptr interface{}, condition string) error {
	var cmd = []string{}
	cmd = append(cmd, "SELECT * FROM ")
	cmd = append(cmd, table)
	cmd = append(cmd, condition)
	rows, err := db.DB.Query(strings.Join(cmd, " ") + ";")
	if err != nil {
		return err
	}
	if rows.Err() != nil {
		return rows.Err()
	}
	defer rows.Close()

	if !rows.Next() {
		return errors.New("sql.Find: null result")
	}
	err = rows.Scan(addrs(objptr)...)
	for rows.Next() {
		if err != nil {
			return err
		}
		err = rows.Scan(addrs(objptr)...)
	}
	return err
}

// CanFind 查询数据库是否有 condition
// condition 可为"WHERE id = 0"
// 默认字段与结构体元素顺序一致
// 返回错误
func (db *Sqlite) CanFind(table string, condition string) bool {
	var cmd = []string{}
	cmd = append(cmd, "SELECT * FROM")
	cmd = append(cmd, table)
	cmd = append(cmd, condition)
	rows, err := db.DB.Query(strings.Join(cmd, " ") + ";")
	if err != nil {
		return false
	}
	if rows.Err() != nil {
		return false
	}
	defer rows.Close()

	if !rows.Next() {
		return false
	}
	_ = rows.Close()
	return true
}

// FindFor 查询数据库，用函数 f 遍历结果
// condition 可为"WHERE id = 0"
// 默认字段与结构体元素顺序一致
// 返回错误
func (db *Sqlite) FindFor(table string, objptr interface{}, condition string, f func() error) error {
	var cmd = []string{}
	cmd = append(cmd, "SELECT * FROM")
	cmd = append(cmd, table)
	cmd = append(cmd, condition)
	rows, err := db.DB.Query(strings.Join(cmd, " ") + ";")
	if err != nil {
		return err
	}
	if rows.Err() != nil {
		return rows.Err()
	}
	defer rows.Close()

	if !rows.Next() {
		return errors.New("sql.FindFor: null result")
	}
	err = rows.Scan(addrs(objptr)...)
	if err == nil {
		err = f()
	}
	for rows.Next() {
		if err != nil {
			return err
		}
		err = rows.Scan(addrs(objptr)...)
		if err == nil {
			err = f()
		}
	}
	return err
}

// Pick 从 table 随机一行
func (db *Sqlite) Pick(table string, objptr interface{}) error {
	return db.Find(table, objptr, "ORDER BY RANDOM() limit 1")
}

// ListTables 列出所有表名
// 返回所有表名+错误
func (db *Sqlite) ListTables() (s []string, err error) {
	rows, err := db.DB.Query("SELECT name FROM sqlite_master where type='table' order by name;")
	if err != nil {
		return
	}
	if rows.Err() != nil {
		return nil, rows.Err()
	}
	defer rows.Close()

	for rows.Next() {
		if err != nil {
			return
		}
		objptr := new(string)
		err = rows.Scan(objptr)
		if err == nil {
			s = append(s, *objptr)
		}
	}
	return
}

// Del 删除数据库表项
// condition 可为"WHERE id = 0"
// 返回错误
func (db *Sqlite) Del(table string, condition string) error {
	var cmd = []string{}
	cmd = append(cmd, "DELETE FROM")
	cmd = append(cmd, table)
	cmd = append(cmd, condition)
	stmt, err := db.DB.Prepare(strings.Join(cmd, " ") + ";")
	if err != nil {
		return err
	}
	_, err = stmt.Exec()
	if err != nil {
		return err
	}
	return stmt.Close()
}

// Truncate 清空数据库表
func (db *Sqlite) Truncate(table string) error {
	var cmd = []string{}
	cmd = append(cmd, "TRUNCATE TABLE")
	cmd = append(cmd, table)
	stmt, err := db.DB.Prepare(strings.Join(cmd, " ") + ";")
	if err != nil {
		return err
	}
	_, err = stmt.Exec()
	if err != nil {
		return err
	}
	return stmt.Close()
}

// Count 查询数据库行数
// 返回行数以及错误
func (db *Sqlite) Count(table string) (num int, err error) {
	var cmd = []string{}
	cmd = append(cmd, "SELECT COUNT(1) FROM")
	cmd = append(cmd, table)
	rows, err := db.DB.Query(strings.Join(cmd, " ") + ";")
	if err != nil {
		return num, err
	}
	if rows.Err() != nil {
		return num, rows.Err()
	}
	if rows.Next() {
		err = rows.Scan(&num)
	}
	rows.Close()
	return num, err
}

// tags 反射 返回结构体对象的 tag 数组
func tags(objptr interface{}) []string {
	var tags []string
	elem := reflect.ValueOf(objptr).Elem()
	// 判断第一个元素是否为匿名字段
	if elem.Type().Field(0).Anonymous {
		elem = elem.Field(0)
	}
	for i, flen := 0, elem.Type().NumField(); i < flen; i++ {
		tags = append(tags, elem.Type().Field(i).Tag.Get("db"))
	}
	return tags
}

// kinds 反射 返回结构体对象的 kinds 数组
func kinds(objptr interface{}) []string {
	var kinds []string
	elem := reflect.ValueOf(objptr).Elem()
	// 判断第一个元素是否为匿名字段
	if elem.Type().Field(0).Anonymous {
		elem = elem.Field(0)
	}
	for i, flen := 0, elem.Type().NumField(); i < flen; i++ {
		switch elem.Field(i).Type().String() {
		case "int8":
			kinds = append(kinds, "TINYINT")
		case "uint8", "byte":
			kinds = append(kinds, "UNSIGNED TINYINT")
		case "int16":
			kinds = append(kinds, "SMALLINT")
		case "uint16":
			kinds = append(kinds, "UNSIGNED SMALLINT")
		case "int32":
			kinds = append(kinds, "INT")
		case "uint32":
			kinds = append(kinds, "UNSIGNED INT")
		case "int64":
			kinds = append(kinds, "BIGINT")
		case "uint64":
			kinds = append(kinds, "UNSIGNED BIGINT")
		case "string":
			kinds = append(kinds, "TEXT")
		default:
			kinds = append(kinds, "TEXT")
		}
	}
	return kinds
}

// values 反射 返回结构体对象的 values 数组
func values(objptr interface{}) []interface{} {
	var values []interface{}
	elem := reflect.ValueOf(objptr).Elem()
	// 判断第一个元素是否为匿名字段
	if elem.Type().Field(0).Anonymous {
		elem = elem.Field(0)
	}
	for i, flen := 0, elem.Type().NumField(); i < flen; i++ {
		values = append(values, elem.Field(i).Interface())
	}
	return values
}

// addrs 反射 返回结构体对象的 addrs 数组
func addrs(objptr interface{}) []interface{} {
	var addrs []interface{}
	elem := reflect.ValueOf(objptr).Elem()
	// 判断第一个元素是否为匿名字段
	if elem.Type().Field(0).Anonymous {
		elem = elem.Field(0)
	}
	for i, flen := 0, elem.Type().NumField(); i < flen; i++ {
		addrs = append(addrs, elem.Field(i).Addr().Interface())
	}
	return addrs
}
