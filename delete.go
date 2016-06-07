package qb

import (
	"reflect"
	"errors"
	"strings"
)

func (self *Tx) Delete(o interface{}) (int64, error) {
	oValue := reflect.Indirect(reflect.ValueOf(o))
	if oValue.Kind() == reflect.Slice {
		return self.deleteSlice(o)
	} else if oValue.Kind() == reflect.Struct {
		return self.deleteOne(o)
	}
	return 0, errors.New("target is not slice or struct.")
}

func (self *Tx) deleteOne(o interface{}) (int64, error) {
	oValue := reflect.Indirect(reflect.ValueOf(o))
	tableNameFunc := oValue.MethodByName("TableName")
	if tableNameFunc.IsNil() {
		return 0, errors.New("func [ (recv Recv) TableName() string ] not found.")
	}
	tableName := tableNameFunc.Call([]reflect.Value{})[0].String()
	oType := oValue.Type()
	pkField := field{}
	for i := 0; i < oValue.NumField(); i ++ {
		oValueField := oValue.Field(i)
		v := oValueField.Interface()
		oTypeField := oType.Field(i)
		pk := strings.TrimSpace(oTypeField.Tag.Get(pk))
		if pk != "" {
			pkField.column = pk
			pkField.value = v
			pkField.kind = oValueField.Kind()
			pkField.elemType = oTypeField.Type
			break
		}
	}
	if pkField.column == "" {
		return 0, errors.New("pk not found to update.")
	}
	ql := ""
	if self.driver == POSTGRES {
		ql = buildPostgresDeleteSql(tableName, pkField)
	} else {
		ql = buildDeleteSql(tableName, pkField)
	}
	fields := []field{pkField}
	lines := [][]field{fields}
	return self.exec(ql, lines)
}


func (self *Tx) deleteSlice(o interface{}) (int64, error) {
	sliceValue := reflect.Indirect(reflect.ValueOf(o))
	elementType := sliceValue.Type().Elem()
	sliceLen := sliceValue.Len()
	lines := make([][]field,0,sliceLen)
	ql := ``
	for i := 0 ; i < sliceLen ; i ++ {
		element := sliceValue.Index(i)
		pkField := field{}
		for i := 0; i < element.NumField(); i ++ {
			oValueField := element.Field(i)
			v := oValueField.Interface()
			oTypeField := elementType.Field(i)
			pk := strings.TrimSpace(oTypeField.Tag.Get(pk))
			if pk != "" {
				pkField.column = pk
				pkField.value = v
				pkField.kind = oValueField.Kind()
				pkField.elemType = oTypeField.Type
				break
			}
		}
		if pkField.column == "" {
			return 0, errors.New("pk not found to update.")
		}
		fields := []field{pkField}
		lines = append(lines, fields)
		if ql == "" {
			tableNameFunc := element.MethodByName("TableName")
			if tableNameFunc.IsNil() {
				return 0, errors.New("func [ (recv Recv) TableName() string ] not found.")
			}
			tableName := tableNameFunc.Call([]reflect.Value{})[0].String()
			if self.driver == POSTGRES {
				ql = buildPostgresDeleteSql(tableName, pkField)
			} else {
				ql = buildDeleteSql(tableName, pkField)
			}
		}
	}
	return self.exec(ql, lines)
}