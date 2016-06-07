package qb

import (
	"reflect"
	"errors"
	"strings"
	"encoding/json"
)

func (self *Tx) Update(o interface{}) (int64, error) {
	oValue := reflect.Indirect(reflect.ValueOf(o))
	if oValue.Kind() == reflect.Slice {
		return self.updateSlice(o)
	} else if oValue.Kind() == reflect.Struct {
		return self.updateOne(o)
	}
	return 0, errors.New("target is not slice or struct.")
}

func (self *Tx) updateOne(o interface{}) (int64, error) {
	oValue := reflect.Indirect(reflect.ValueOf(o))
	tableNameFunc := oValue.MethodByName("TableName")
	if tableNameFunc.IsNil() {
		return 0, errors.New("func [ (recv Recv) TableName() string ] not found.")
	}
	tableName := tableNameFunc.Call([]reflect.Value{})[0].String()
	oType := oValue.Type()
	fields := make([]field,0,8)
	pkField := field{}
	versionField := field{}
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
			continue
		}
		version := strings.TrimSpace(oTypeField.Tag.Get(version))
		if version != "" {
			versionField.column = version
			versionField.value = v
			versionField.kind = oValueField.Kind()
			versionField.elemType = oTypeField.Type
			continue
		}
		col := strings.TrimSpace(oTypeField.Tag.Get(col))
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
	if pkField.column == "" {
		return 0, errors.New("pk not found to update.")
	}
	fieldsLen := len(fields)
	if fieldsLen == 0 {
		return 0, errors.New("no field to update.")
	}
	ql := ""
	if self.driver == POSTGRES {
		ql = buildPostgresUpdateSql(tableName, pkField, versionField, fields)
	} else {
		ql = buildUpdateSql(tableName, pkField, versionField, fields)
	}
	fields = append(fields, pkField)
	if versionField.column != "" {
		fields = append(fields, versionField)
	}
	lines := [][]field{fields}
	return self.exec(ql, lines)
}


func (self *Tx) updateSlice(o interface{}) (int64, error) {
	sliceValue := reflect.Indirect(reflect.ValueOf(o))
	elementType := sliceValue.Type().Elem()
	sliceLen := sliceValue.Len()
	lines := make([][]field,0,sliceLen)
	ql := ``
	for i := 0 ; i < sliceLen ; i ++ {
		element := sliceValue.Index(i)
		fields := make([]field,0,8)
		pkField := field{}
		versionField := field{}
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
				continue
			}
			version := strings.TrimSpace(oTypeField.Tag.Get(version))
			if version != "" {
				versionField.column = version
				versionField.value = v
				versionField.kind = oValueField.Kind()
				versionField.elemType = oTypeField.Type
				continue
			}
			col := strings.TrimSpace(oTypeField.Tag.Get(col))
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
		if pkField.column == "" {
			return 0, errors.New("pk not found to update.")
		}
		fieldsLen := len(fields)
		if fieldsLen == 0 {
			return 0, errors.New("no field to update.")
		}
		if ql == "" {
			tableNameFunc := element.MethodByName("TableName")
			if tableNameFunc.IsNil() {
				return 0, errors.New("func [ (recv Recv) TableName() string ] not found.")
			}
			tableName := tableNameFunc.Call([]reflect.Value{})[0].String()
			if self.driver == POSTGRES {
				ql = buildPostgresUpdateSql(tableName, pkField, versionField, fields)
			} else {
				ql = buildUpdateSql(tableName, pkField, versionField, fields)
			}
		}
		fields = append(fields, pkField)
		if versionField.column != "" {
			fields = append(fields, versionField)
		}
		lines = append(lines, fields)
	}
	return self.exec(ql, lines)
}