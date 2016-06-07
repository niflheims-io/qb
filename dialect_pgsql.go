package qb

import (
	"fmt"
	"strings"
)

func buildPostgresInsertSql(tableName string, fields []field) string {
	fieldQl := ``
	valueQl := ``
	fieldsLen := len(fields)
	for i := 0 ; i < fieldsLen ; i ++  {
		f := fields[i]
		fieldQl = fieldQl + `"` + f.column + `",`
		vIndex := i + 1
		valueQl = valueQl + `$` + fmt.Sprint(vIndex) + `,`
	}
	fieldQl = fieldQl[0:len(fieldQl) - 1]
	valueQl = valueQl[0:len(valueQl) - 1]
	ql := `INSERT INTO "` + tableName + `" (` + fieldQl + `) VALUES (` + valueQl + `) `
	ql = strings.ToUpper(ql)
	return ql
}

func buildPostgresUpdateSql(tableName string, pkField field, versionField field, fields []field) string {
	setQl := ``
	fieldsLen := len(fields)
	for i := 0 ; i < fieldsLen ; i ++  {
		f := fields[i]
		setQl = setQl + `"` + f.column + `" = `
		vIndex := i + 1
		setQl = setQl + `$` + fmt.Sprint(vIndex) + `,`
	}
	setQl = setQl[0:len(setQl) - 1]
	conditionQl := `"` + pkField.column + `" = $` + fmt.Sprint(fieldsLen + 1)
	if versionField.column != "" {
		setQl = setQl + `, "` + versionField.column + `" = "` + versionField.column + `" + 1`
		conditionQl = conditionQl + ` AND "` + versionField.column + `" = $` + fmt.Sprint(fieldsLen + 2)
	}
	ql := `UPDATE "` + tableName + `" SET ` + setQl + ` WHERE ` + conditionQl
	ql = strings.ToUpper(ql)
	return ql
}

func buildPostgresDeleteSql(tableName string, pkField field) string {
	ql := `DELETE FROM "` + tableName + `" WHERE "` + pkField.column + `" = $1`
	ql = strings.ToUpper(ql)
	return ql
}

