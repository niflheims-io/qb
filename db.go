package qb

import "database/sql"

type QB struct  {
	db *sql.DB
	driver string
}

func New(db *sql.DB, driver string) *QB {
	return &QB{db:db, driver:driver}
}

func (self *QB) BeginTx() (*Tx, error) {
	sqlTx, txBeginErr := self.db.Begin()
	if txBeginErr != nil {
		return nil, txBeginErr
	}
	self.db.Driver()
	tx := Tx{}
	tx.tx = sqlTx
	tx.driver = self.driver
	return &tx, nil
}

