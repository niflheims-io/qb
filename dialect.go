package qb

import (
	"strings"
)

func buildInsertSql(tableName string, fields []field) string {
	fieldQl := ``
	valueQl := ``
	fieldsLen := len(fields)
	for i := 0 ; i < fieldsLen ; i ++  {
		f := fields[i]
		fieldQl = fieldQl + `` + f.column + `,`
		valueQl = valueQl + `?,`
	}
	fieldQl = fieldQl[0:len(fieldQl) - 1]
	valueQl = valueQl[0:len(valueQl) - 1]
	ql := `INSERT INTO ` + tableName + ` (` + fieldQl + `) VALUES (` + valueQl + `) `
	ql = strings.ToUpper(ql)
	return ql
}

func buildUpdateSql(tableName string, pkField field, versionField field, fields []field) string {
	setQl := ``
	fieldsLen := len(fields)
	for i := 0; i < fieldsLen; i ++ {
		f := fields[i]
		setQl = setQl + f.column + ` = `
		setQl = setQl + `?,`
	}
	setQl = setQl[0:len(setQl) - 1]
	conditionQl := pkField.column + ` = ?`
	if versionField.column != "" {
		setQl = setQl + `, ` + versionField.column + ` = ` + versionField.column + ` + 1`
		conditionQl = conditionQl + ` AND ` + versionField.column + ` = ?`
	}
	ql := `UPDATE ` + tableName + ` SET ` + setQl + ` WHERE ` + conditionQl
	ql = strings.ToUpper(ql)
	return ql
}

func buildDeleteSql(tableName string, pkField field) string {
	ql := `DELETE FROM ` + tableName + ` WHERE ` + pkField.column + ` = ?`
	ql = strings.ToUpper(ql)
	return ql
}


