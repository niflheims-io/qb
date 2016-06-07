package qb

import "database/sql"

const (
	POSTGRES = "postgres"
	MYSQL = "mysql"
	ORACLE = "oracle"
)

type Tx struct  {
	tx *sql.Tx
	driver string
}

func (self *Tx) Commit() error {
	return self.tx.Commit()
}

func (self *Tx) Rollback() error {
	return self.tx.Rollback()
}
