package qb

import (
	"reflect"
)

const (
	col = "col"
	pk = "pk"
	version = "version"
	warp = "warp"
)

type field struct  {
	column string
	kind reflect.Kind
	warp string
	value interface{}
	elemType reflect.Type
}

type DataBaseTable interface {
	TableName() string
	PkColumn() string
	DeleteColumn() string
}


