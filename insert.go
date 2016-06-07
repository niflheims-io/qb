package qb

import (
	"reflect"
	"errors"
	"encoding/json"
	"strings"
)

func (self *Tx) Insert(o interface{}) (int64, error) {
	oValue := reflect.Indirect(reflect.ValueOf(o))
	if oValue.Kind() == reflect.Slice {
		return self.insertSlice(o)
	} else if oValue.Kind() == reflect.Struct {
		return self.insertOne(o)
	}
	return 0, errors.New("target is not slice or struct.")
}

func (self *Tx) insertOne(o interface{}) (int64, error) {
	oValue := reflect.Indirect(reflect.ValueOf(o))
	tableNameFunc := oValue.MethodByName("TableName")
	if tableNameFunc.IsNil() {
		return 0, errors.New("func [ (recv Recv) TableName() string ] not found.")
	}
	tableName := tableNameFunc.Call([]reflect.Value{})[0].String()
	oType := oValue.Type()
	fields := make([]field,0,8)
	for i := 0; i < oValue.NumField(); i ++ {
		oValueField := oValue.Field(i)
		v := oValueField.Interface()
		oTypeField := oType.Field(i)
		col := strings.TrimSpace(oTypeField.Tag.Get(col))
		if col == "" {
			col = strings.TrimSpace(oTypeField.Tag.Get(pk))
		}
		if col == "" {
			col = strings.TrimSpace(oTypeField.Tag.Get(version))
		}
		if col == "" {
			continue
		}
		warp := oTypeField.Tag.Get(warp)
		if warp == "jsonb" {
			jsonB, marshalErr := json.Marshal(&v)
			if marshalErr != nil {
				return 0, marshalErr
			}
			v = jsonB
		}
		f := field{}
		f.column = col
		f.value = v
		f.kind = oValueField.Kind()
		f.elemType = oTypeField.Type
		fields = append(fields, f)
	}
	fieldsLen := len(fields)
	if fieldsLen == 0 {
		return 0, errors.New("no field to insert.")
	}
	ql := ""
	if self.driver == POSTGRES {
		ql = buildPostgresInsertSql(tableName, fields)
	} else {
		ql = buildInsertSql(tableName, fields)
	}
	lines := [][]field{fields}
	return self.exec(ql, lines)
}


func (self *Tx) insertSlice(o interface{}) (int64, error) {
	sliceValue := reflect.Indirect(reflect.ValueOf(o))
	elementType := sliceValue.Type().Elem()
	sliceLen := sliceValue.Len()
	lines := make([][]field,0,sliceLen)
	ql := ``
	for i := 0 ; i < sliceLen ; i ++ {
		element := sliceValue.Index(i)
		fields := make([]field,0,8)
		for i := 0; i < element.NumField(); i ++ {
			oValueField := element.Field(i)
			v := oValueField.Interface()
			oTypeField := elementType.Field(i)
			col := strings.TrimSpace(oTypeField.Tag.Get(col))
			if col == "" {
				col = strings.TrimSpace(oTypeField.Tag.Get(pk))
			}
			if col == "" {
				col = strings.TrimSpace(oTypeField.Tag.Get(version))
			}
			if col == "" {
				continue
			}
			warp := oTypeField.Tag.Get(warp)
			if warp == "jsonb" {
				jsonB, marshalErr := json.Marshal(&v)
				if marshalErr != nil {
					return 0, marshalErr
				}
				v = jsonB
			}
			f := field{}
			f.column = col
			f.value = v
			f.kind = oValueField.Kind()
			f.elemType = oTypeField.Type
			fields = append(fields, f)
		}
		fieldsLen := len(fields)
		if fieldsLen == 0 {
			return 0, errors.New("no field to insert.")
		}
		lines = append(lines, fields)
		if ql == "" {
			tableNameFunc := element.MethodByName("TableName")
			if tableNameFunc.IsNil() {
				return 0, errors.New("func [ (recv Recv) TableName() string ] not found.")
			}
			tableName := tableNameFunc.Call([]reflect.Value{})[0].String()
			if self.driver == POSTGRES {
				ql = buildPostgresInsertSql(tableName, fields)
			} else {
				ql = buildInsertSql(tableName, fields)
			}
		}
	}
	return self.exec(ql, lines)
}
