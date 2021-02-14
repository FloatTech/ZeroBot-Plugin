package utils

import (
	"database/sql"
	"errors"
	"fmt"
	"reflect"

	_ "github.com/mattn/go-sqlite3"
)

type Sqlite struct {
	DB     *sql.DB
	DBPath string
}

// DBCreate 根据结构体生成数据库table，tag为"id"为主键，自增
func (db *Sqlite) DBCreate(objptr interface{}) (err error) {
	if db.DB == nil {
		database, err := sql.Open("sqlite3", db.DBPath)
		if err != nil {
			return err
		}
		db.DB = database
	}

	table := fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (", Struct2name(objptr))
	for i, column := range strcut2columns(objptr) {
		table += fmt.Sprintf(" %s %s NULL", column, column2type(objptr, column))
		if i+1 != len(strcut2columns(objptr)) {
			table += ","
		} else {
			table += " );"
		}
	}
	if _, err := db.DB.Exec(table); err != nil {
		return err
	}
	return nil
}

// DBInsert 根据结构体插入一条数据
func (db *Sqlite) DBInsert(objptr interface{}) (err error) {
	defer func() {
		if err := recover(); err != nil {
			panic(err)
		}
	}()
	rows, err := db.DB.Query("SELECT * FROM " + Struct2name(objptr))
	if err != nil {
		return err
	}
	defer rows.Close()

	columns, _ := rows.Columns()
	index := -1
	names := "("
	insert := "("
	for i, column := range columns {
		if column == "id" {
			index = i
			continue
		}
		if i != len(columns)-1 {
			names += column + ","
			insert += "?,"
		} else {
			names += column + ")"
			insert += "?)"
		}
	}
	stmt, err := db.DB.Prepare("INSERT INTO " + Struct2name(objptr) + names + " values " + insert)
	if err != nil {
		return err
	}

	value := []interface{}{}
	if index == -1 {
		value = append(value, struct2values(objptr, columns)...)
	} else {
		value = append(value, append(struct2values(objptr, columns)[:index], struct2values(objptr, columns)[index+1:]...)...)
	}
	_, err = stmt.Exec(value...)
	if err != nil {
		return err
	}
	return nil
}

// DBSelect 根据结构体查询对应的表，cmd可为"WHERE id = 0 "
func (db *Sqlite) DBSelect(objptr interface{}, cmd string) (err error) {
	rows, err := db.DB.Query(fmt.Sprintf("SELECT * FROM %s %s", Struct2name(objptr), cmd))
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		columns, err := rows.Columns()
		if err != nil {
			return err
		}
		err = rows.Scan(struct2addrs(objptr, columns)...)
		if err != nil {
			return err
		}
		return nil
	}
	return errors.New("Database no such elem")
}

// DBDelete 删除struct对应表的一行，返回错误
func (db *Sqlite) DBDelete(objptr interface{}, cmd string) (err error) {
	stmt, err := db.DB.Prepare(fmt.Sprintf("DELETE FROM %s %s", Struct2name(objptr), cmd))
	if err != nil {
		return err
	}
	_, err = stmt.Exec()
	if err != nil {
		return err
	}
	return nil
}

// DBNum 查询struct对应表的行数,返回行数以及错误
func (db *Sqlite) DBNum(objptr interface{}) (num int, err error) {
	rows, err := db.DB.Query(fmt.Sprintf("SELECT * FROM %s", Struct2name(objptr)))
	if err != nil {
		return num, err
	}
	defer rows.Close()
	for rows.Next() {
		num++
	}
	return num, nil
}

// strcut2columns 反射得到结构体的 tag 数组
func strcut2columns(objptr interface{}) []string {
	var columns []string
	elem := reflect.ValueOf(objptr).Elem()
	// TODO 判断第一个元素是否为匿名字段
	if elem.Type().Field(0).Anonymous {
		elem = elem.Field(0)
	}
	for i, flen := 0, elem.Type().NumField(); i < flen; i++ {
		columns = append(columns, elem.Type().Field(i).Tag.Get("db"))
	}
	return columns
}

// Struct2name 反射得到结构体的名字
func Struct2name(objptr interface{}) string {
	return reflect.ValueOf(objptr).Elem().Type().Name()
}

// column2type 反射得到结构体对应 tag 的 数据库数据类型
func column2type(objptr interface{}, column string) string {
	type_ := ""
	elem := reflect.ValueOf(objptr).Elem()
	// TODO 判断第一个元素是否为匿名字段
	if elem.Type().Field(0).Anonymous {
		elem = elem.Field(0)
	}
	for i, flen := 0, elem.Type().NumField(); i < flen; i++ {
		if column == elem.Type().Field(i).Tag.Get("db") {
			type_ = elem.Field(i).Type().String()
		}
	}
	if column == "id" {
		return "INTEGER PRIMARY KEY"
	}
	switch type_ {
	case "int64":
		return "INT"
	case "string":
		return "TEXT"
	default:
		return "TEXT"
	}
}

// struct2addrs 反射得到结构体对应数据库字段的属性地址
func struct2addrs(objptr interface{}, columns []string) []interface{} {
	var addrs []interface{}
	elem := reflect.ValueOf(objptr).Elem()
	// TODO 判断第一个元素是否为匿名字段
	if elem.Type().Field(0).Anonymous {
		elem = elem.Field(0)
	}
	for _, column := range columns {
		for i, flen := 0, elem.Type().NumField(); i < flen; i++ {
			if column == elem.Type().Field(i).Tag.Get("db") {
				addrs = append(addrs, elem.Field(i).Addr().Interface())
			}
		}
	}
	return addrs
}

// struct2values 反射得到结构体对应数据库字段的属性值
func struct2values(objptr interface{}, columns []string) []interface{} {
	var values []interface{}
	elem := reflect.ValueOf(objptr).Elem()
	// TODO 判断第一个元素是否为匿名字段
	if elem.Type().Field(0).Anonymous {
		elem = elem.Field(0)
	}
	for _, column := range columns {
		for i, flen := 0, elem.Type().NumField(); i < flen; i++ {
			if column == elem.Type().Field(i).Tag.Get("db") {
				switch elem.Field(i).Type().String() {
				case "int64":
					values = append(values, elem.Field(i).Int())
				case "string":
					values = append(values, elem.Field(i).String())
				default:
					values = append(values, elem.Field(i).String())
				}
			}
		}
	}
	return values
}
