package qb

import (
	"errors"
	"reflect"
	"database/sql"
	"time"
	"encoding/json"
	"strings"
	"fmt"
)

type qor interface  {
	Prepare(string) (*sql.Stmt, error)
}

type query struct  {
	q qor
	ql string
	args []interface{}
}


func (self *QB) Query(ql string, args ...interface{}) *query {
	return &query{ql:ql, args:args, q:self.db}
}

func (self *query) List(targetSlice interface{}) error {
	if self.ql == "" {
		return errors.New("ql is empty")
	}
	sliceValue := reflect.Indirect(reflect.ValueOf(targetSlice))
	if sliceValue.Kind() != reflect.Slice && sliceValue.Kind() != reflect.Map {
		return errors.New("needs a pointer to a slice or a map")
	}
	sliceElementType := sliceValue.Type().Elem()
	fieldMap := make(map[string]field)
	for i := 0; i < sliceElementType.NumField(); i++ {
		colTag := strings.TrimSpace(sliceElementType.Field(i).Tag.Get(col))
		if colTag == "" {
			colTag = strings.TrimSpace(sliceElementType.Field(i).Tag.Get(pk))
		}
		if colTag == "" {
			colTag = strings.TrimSpace(sliceElementType.Field(i).Tag.Get(version))
		}
		if colTag == "" {
			continue
		}
		warpTag := strings.TrimSpace(sliceElementType.Field(i).Tag.Get(warp))
		fieldKind := sliceElementType.Field(i).Type.Kind()
		f := field{
			column:colTag,
			kind:fieldKind,
			warp:warpTag,
			elemType:sliceElementType.Field(i).Type,
		}
		fieldMap[colTag] = f
	}
	stmt, stmtErr := self.q.Prepare(self.ql)
	if stmtErr != nil {
		return stmtErr
	}
	defer stmt.Close()
	rows, rowErr := stmt.Query(self.args...)
	if rowErr != nil {
		return rowErr
	}
	defer rows.Close()
	columns, columnsErr := rows.Columns()
	if columnsErr != nil {
		return columnsErr
	}
	columnLen := len(columns)
	scans := make([]interface{},0, columnLen)
	for i := 0 ; i < columnLen ; i ++  {
		column := columns[i]
		if t, ok := fieldMap[column] ; ok {
			if strings.TrimSpace(t.warp) == "jsonb" && t.kind == reflect.Struct{
				var cell []byte
				scans = append(scans, &cell)
			} else {
				if t.kind == reflect.String {
					var cell sql.NullString
					scans = append(scans, &cell)
				} else if t.kind == reflect.Int64 {
					var cell sql.NullInt64
					scans = append(scans, &cell)
				} else if t.kind == reflect.Float64 {
					var cell sql.NullFloat64
					scans = append(scans, &cell)
				} else if t.kind == reflect.Bool {
					var cell sql.NullBool
					scans = append(scans, &cell)
				} else if t.kind == reflect.TypeOf(time.Time{}).Kind() {
					var cell time.Time
					scans = append(scans, &cell)
				} else if t.kind == reflect.TypeOf([]byte{}).Kind() {
					var cell []byte
					scans = append(scans, &cell)
				} else {
					return errors.New(fmt.Sprint("unknow type", column, t))
				}

			}
		} else {
			var cell interface{}
			scans = append(scans, &cell)
		}
	}
	for rows.Next() {
		scanErr := rows.Scan(scans...)
		if scanErr != nil {
			return scanErr
		}
		sliceElement := reflect.New(sliceElementType).Interface()
		fieldValueMap := make(map[string]field)
		for i := 0; i < len(scans); i++ {
			column := columns[i]
			val := scans[i]
			if t, ok := fieldMap[column] ; ok {
				if strings.TrimSpace(t.warp) == "jsonb" && t.kind == reflect.Struct{
					v := reflect.ValueOf(val).Elem()
					if v.Kind() == reflect.String {
						str := (val).(string)
						t.value = []byte(str)
					}
					bytes := reflect.Indirect(reflect.ValueOf(val)).Bytes()
					jsonVal := reflect.New(t.elemType).Interface()
					unMarshalErr := json.Unmarshal(bytes, &jsonVal)
					if unMarshalErr != nil {
						return unMarshalErr
					}
					t.value = jsonVal
				} else {
					if t.kind == reflect.String {
						t.value = (val).(*sql.NullString).String
					} else if t.kind == reflect.Int64 {
						t.value = (val).(*sql.NullInt64).Int64
					} else if t.kind == reflect.Float64 {
						t.value = (val).(*sql.NullFloat64).Float64
					} else if t.kind == reflect.Bool {
						t.value = (val).(*sql.NullBool).Bool
					} else if t.kind == reflect.TypeOf(time.Time{}).Kind() {
						t.value = (val).(*time.Time)
					} else if t.kind == reflect.TypeOf([]byte{}).Kind() {
						t.value = (val).(*[]byte)
					} else {
						return errors.New(fmt.Sprint("unknow type ", column, t))
					}
				}
				fieldValueMap[t.column] = t
			}
			rowElementMap(sliceElement, fieldValueMap)
		}
		sliceValue.Set(reflect.Append(sliceValue, reflect.Indirect(reflect.ValueOf(sliceElement))))
	}
	return nil
}

func (self *query) One(target interface{}) error {
	elementValue := reflect.Indirect(reflect.ValueOf(target))
	if elementValue.Kind() == reflect.Struct {
		return self.loadInterface(elementValue)
	} else {
		return self.loadBasicVariable(elementValue)
	}
}

func (self *query) Int(target *int64) error {
	elementValue := reflect.Indirect(reflect.ValueOf(target))
	return self.loadBasicVariable(elementValue)
}

func (self *query) Float(target *float64) error {
	elementValue := reflect.Indirect(reflect.ValueOf(target))
	return self.loadBasicVariable(elementValue)
}

func (self *query) Bool(target *bool) error {
	elementValue := reflect.Indirect(reflect.ValueOf(target))
	return self.loadBasicVariable(elementValue)
}

func (self *query) Time(target *time.Time) error {
	elementValue := reflect.Indirect(reflect.ValueOf(target))
	return self.loadBasicVariable(elementValue)
}

func (self *query) Bytes(target *[]byte) error {
	elementValue := reflect.Indirect(reflect.ValueOf(target))
	return self.loadBasicVariable(elementValue)
}

func (self *query) String(target *string) error {
	elementValue := reflect.Indirect(reflect.ValueOf(target))
	return self.loadBasicVariable(elementValue)
}

func (self *query) loadInterface(elementValue reflect.Value) error {
	elementType := elementValue.Type()
	fieldMap := make(map[string]field)
	for i := 0; i < elementType.NumField(); i++ {
		colTag := strings.TrimSpace(elementType.Field(i).Tag.Get(col))
		if colTag == "" {
			colTag = strings.TrimSpace(elementType.Field(i).Tag.Get(pk))
		}
		if colTag == "" {
			colTag = strings.TrimSpace(elementType.Field(i).Tag.Get(version))
		}
		if colTag == "" {
			continue
		}
		warpTag := strings.TrimSpace(elementType.Field(i).Tag.Get(warp))
		fieldKind := elementType.Field(i).Type.Kind()
		f := field{
			column:colTag,
			kind:fieldKind,
			warp:warpTag,
			elemType:elementType.Field(i).Type,
		}
		fieldMap[colTag] = f
	}
	db := self.q
	stmt, stmtErr := db.Prepare(self.ql)
	if stmtErr != nil {
		return stmtErr
	}
	defer stmt.Close()
	rows, rowErr := stmt.Query(self.args...)
	if rowErr != nil {
		return rowErr
	}
	defer rows.Close()
	columns, columnsErr := rows.Columns()
	if columnsErr != nil {
		return columnsErr
	}
	columnLen := len(columns)
	scans := make([]interface{},0, columnLen)
	for i := 0 ; i < columnLen ; i ++  {
		column := columns[i]
		if t, ok := fieldMap[column] ; ok {
			if strings.TrimSpace(t.warp) == "jsonb" && t.kind == reflect.Struct{
				var cell []byte
				scans = append(scans, &cell)
			} else {
				if t.kind == reflect.String {
					var cell sql.NullString
					scans = append(scans, &cell)
				} else if t.kind == reflect.Int64 {
					var cell sql.NullInt64
					scans = append(scans, &cell)
				} else if t.kind == reflect.Float64 {
					var cell sql.NullFloat64
					scans = append(scans, &cell)
				} else if t.kind == reflect.Bool {
					var cell sql.NullBool
					scans = append(scans, &cell)
				} else if t.kind == reflect.TypeOf(time.Time{}).Kind() {
					var cell time.Time
					scans = append(scans, &cell)
				} else if t.kind == reflect.TypeOf([]byte{}).Kind() {
					var cell []byte
					scans = append(scans, &cell)
				} else {
					return errors.New(fmt.Sprint("unknow type", column, t))
				}

			}
		} else {
			var cell interface{}
			scans = append(scans, &cell)
		}
	}
	if rows.Next() {
		scanErr := rows.Scan(scans...)
		if scanErr != nil {
			return scanErr
		}
		fieldValueMap := make(map[string]field)
		for i := 0; i < len(scans); i++ {
			column := columns[i]
			val := scans[i]
			if t, ok := fieldMap[column] ; ok {
				if strings.TrimSpace(t.warp) == "jsonb" && t.kind == reflect.Struct{
					v := reflect.ValueOf(val).Elem()
					if v.Kind() == reflect.String {
						str := (val).(string)
						t.value = []byte(str)
					}
					bytes := reflect.Indirect(reflect.ValueOf(val)).Bytes()
					jsonVal := reflect.New(t.elemType).Interface()
					unMarshalErr := json.Unmarshal(bytes, &jsonVal)
					if unMarshalErr != nil {
						return unMarshalErr
					}
					t.value = jsonVal
				} else {
					if t.kind == reflect.String {
						t.value = (val).(*sql.NullString).String
					} else if t.kind == reflect.Int64 {
						t.value = (val).(*sql.NullInt64).Int64
					} else if t.kind == reflect.Float64 {
						t.value = (val).(*sql.NullFloat64).Float64
					} else if t.kind == reflect.Bool {
						t.value = (val).(*sql.NullBool).Bool
					} else if t.kind == reflect.TypeOf(time.Time{}).Kind() {
						t.value = (val).(*time.Time)
					} else if t.kind == reflect.TypeOf([]byte{}).Kind() {
						t.value = (val).(*[]byte)
					} else {
						return errors.New(fmt.Sprint("unknow type ", column, t))
					}
				}
				fieldValueMap[t.column] = t
			}
			for i := 0 ; i < elementType.NumField() ; i ++  {
				f := elementValue.Field(i)
				colTag := strings.TrimSpace(elementType.Field(i).Tag.Get(col))
				if colTag == "" {
					colTag = strings.TrimSpace(elementType.Field(i).Tag.Get(pk))
				}
				if colTag == "" {
					colTag = strings.TrimSpace(elementType.Field(i).Tag.Get(version))
				}
				if field, ok := fieldValueMap[colTag]; ok {
					f.Set(reflect.Indirect(reflect.ValueOf(field.value)))
				}
			}
		}
	}
	return nil
}


func (self *query) loadBasicVariable(elementValue reflect.Value) error {
	scans := make([]interface{},0,1)
	kind := elementValue.Kind()
	if kind == reflect.String {
		var cell sql.NullString
		scans = append(scans, &cell)
	} else if kind == reflect.Int64 {
		var cell sql.NullInt64
		scans = append(scans, &cell)
	} else if kind == reflect.Float64 {
		var cell sql.NullFloat64
		scans = append(scans, &cell)
	} else if kind == reflect.Bool {
		var cell sql.NullBool
		scans = append(scans, &cell)
	} else if kind == reflect.TypeOf(time.Time{}).Kind() {
		var cell time.Time
		scans = append(scans, &cell)
	} else if kind == reflect.TypeOf([]byte{}).Kind() {
		var cell []byte
		scans = append(scans, &cell)
	} else {
		return errors.New(fmt.Sprint("unknow type", kind))
	}
	db := self.q
	stmt, stmtErr := db.Prepare(self.ql)
	if stmtErr != nil {
		return stmtErr
	}
	defer stmt.Close()
	rows, rowErr := stmt.Query(self.args...)
	if rowErr != nil {
		return rowErr
	}
	defer rows.Close()
	if rows.Next() {
		scanErr := rows.Scan(scans...)
		if scanErr != nil {
			return scanErr
		}
	}
	if len(scans) == 0 {
		return errors.New("scan nothing from rows.")
	}
	cellValue := scans[0]
	if kind == reflect.String {
		cellValue = (cellValue).(*sql.NullString).String
	} else if kind == reflect.Int64 {
		cellValue = (cellValue).(*sql.NullInt64).Int64
	} else if kind == reflect.Float64 {
		cellValue = (cellValue).(*sql.NullFloat64).Float64
	} else if kind == reflect.Bool {
		cellValue = (cellValue).(*sql.NullBool).Bool
	} else if kind == reflect.TypeOf(time.Time{}).Kind() {
		cellValue = (cellValue).(*time.Time)
	} else if kind == reflect.TypeOf([]byte{}).Kind() {
		cellValue = (cellValue).(*[]byte)
	} else {
		return errors.New(fmt.Sprint("unknow type", kind, cellValue))
	}
	elementValue.Set(reflect.Indirect(reflect.ValueOf(cellValue)))
	return nil
}

func rowElementMap(elem interface{}, fieldMap map[string]field)  {
	v := reflect.ValueOf(elem).Elem()
	t := reflect.TypeOf(elem).Elem()
	for i := 0 ; i < t.NumField() ; i ++  {
		f := v.Field(i)
		colTag := strings.TrimSpace(t.Field(i).Tag.Get(col))
		if colTag == "" {
			colTag = strings.TrimSpace(t.Field(i).Tag.Get(pk))
		}
		if colTag == "" {
			colTag = strings.TrimSpace(t.Field(i).Tag.Get(version))
		}
		if field, ok := fieldMap[colTag]; ok {
			f.Set(reflect.Indirect(reflect.ValueOf(field.value)))
		}
	}
}