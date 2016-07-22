qb
==========

QB is a light ORM for golang.


# Features

* Struct <-> Table Mapping Support.

* Struct <-> Jsonb type Mapping Support.

* Struct <-> Columns Mapping Support, Support for custom query results.

* Low dependency.

* Transaction Support.

* Both ORM and raw SQL operation Support


Installation
------------

```
go get github.com/niflheims-io/qb
```

Examples
--------

### Tables ###

```sql
CREATE TABLE "DOC"
(
  "UUID" character varying(64) NOT NULL,
  "DATA" jsonb,
  "CREATE_USER" character varying(64),
  "CREATE_TIME" timestamp without time zone,
  "UPDATE_USER" character varying(64),
  "UPDATE_TIME" timestamp without time zone,
  "DELETE_USER" character varying(64) DEFAULT ''::character varying,
  "DELETE_TIME" timestamp without time zone,
  "VERSION" numeric(8,0) NOT NULL DEFAULT 1,
  CONSTRAINT "DOC_PK" PRIMARY KEY ("UUID")
)
```

```sql
CREATE TABLE "DOC_COMMENT"
(
  "UUID" character varying(64) NOT NULL,
  "DOC_ID" character varying(64) NOT NULL,
  "TO" character varying(64) NOT NULL DEFAULT ''::character varying,
  "CONTENT" character varying(10240) NOT NULL DEFAULT ''::character varying,
  "CREATE_USER" character varying(64),
  "CREATE_TIME" timestamp without time zone,
  "UPDATE_USER" character varying(64),
  "UPDATE_TIME" timestamp without time zone,
  "DELETE_USER" character varying(64) DEFAULT ''::character varying,
  "DELETE_TIME" timestamp without time zone,
  "VERSION" numeric(8,0) NOT NULL DEFAULT 1,
  CONSTRAINT "DOC_COMMENT_PK" PRIMARY KEY ("UUID")
)
```

### structs ###


```go
type DocTable struct  {
	Id string		`pk:"UUID"`
	Data DocDataColumn		`col:"DATA" warp:"jsonb"`
	CreateUser string	`col:"CREATE_USER"`
	CreateTime time.Time	`col:"CREATE_TIME"`
	UpdateUser string	`col:"UPDATE_USER"`
	UpdateTime time.Time	`col:"UPDATE_TIME"`
	DeleteUser string	`col:"DELETE_USER"`
	DeleteTime time.Time	`col:"DELETE_TIME"`
	Version int64		`version:"VERSION"`
}

func (self DocTable) TableName() string {
	return "DOC"
}

type DocDataColumn struct {
    Title string   `json:"title"`
    Tags []string `json:"tags"`
    Content string `json:"content"`
    Likes int64 `json:"likes"`
}

type DocCommentTable struct  {
	Id string		`pk:"UUID"`
	CreateUser string	`col:"CREATE_USER"`
	CreateTime time.Time	`col:"CREATE_TIME"`
	UpdateUser string	`col:"UPDATE_USER"`
	UpdateTime time.Time	`col:"UPDATE_TIME"`
	DeleteUser string	`col:"DELETE_USER"`
	DeleteTime time.Time	`col:"DELETE_TIME"`
	Version int64		`version:"VERSION"`
	DocId string		`col:"DOC_ID"`
	Content string		`col:"CONTENT"`
	UserTo string		`col:"TO"`
}

func (self DocCommentTable) TableName() string {
	return "DOC_COMMENT"
}

```

Custom query results

```go

type DocOwnerCommentView struct  {
	DocId string		`col:"DOC_ID"`
	Content string		`col:"CONTENT"`
}

```

### Usage ###

Create qb

```go
    dbInstance, dbOpenErr := sql.Open("driver_name", "connect_url")
	defer dbInstance.Close()
	if dbOpenErr != nil {
		sys_log.Println(dbOpenErr)
		os.Exit(2)
	}
	dbInstance.SetMaxOpenConns(8)
	dbInstance.SetMaxIdleConns(2)
	db := qb.New(dbInstance, qb.POSTGRES) // postgres
```

Do some queries

```go
    // rows
    docRows := make([]DocTable, 0, 8)
    limit := int64(8)
    docRowsErr := db.Query(`SELECT * FROM "DOC" WHERE "DELETE_USER" = '' ORDER BY "CREATE_TIME" DESC LIMIT $1 OFFSET 0 `, &limit).List(&docRows)
    if docRowsErr != nil {
        // todo something
    }
    // detail with docRows

    // row
    docRow := DocTable{}
    docId := "uuid"
    docRowErr := db.Query(`SELECT * FROM "DOC" WHERE "UUID" = $1 `, &docId).One(&docRow)
    if docRowErr != nil {
        // todo something
    }
    // detail with docRow

    // custom query results
    customSQL := `SELECT "C"."DOC_ID", "C"."CONTENT"
    FROM "DOC_COMMENT" AS "C"
    LEFT JOIN "DOC" AS "D"
    WHERE "D"."CREATE_USER" = $1 `
    userId := "someId"
    customRows := make([]DocOwnerCommentView, 0, 8)
    customRowsErr := db.Query(customSQL, &userId).List(&customRows)
    if customRowsErr != nil {
        // todo something
    }
    // detail with customRows

    // one column
    countSQL := `SELECT COUNT(*) FROM "DOC" `
    docCount := int64(0)
    docCountErr := db.Query(countSQL).Int(&docCount) // String(), Float(), Bool() ...

```

Transaction

```go
    tx, txBegErr := db.Begin()
    tx.Query(...).Func(&{})
    tx.Exec(...)
    tx.Insert(&{})
    tx.Insert(&[]{})
    tx.Update(&{})
    tx.Update(&[]{})
    tx.Delete(&{})
    tx.Delete(&[]{})
    tx.Rollback()
    tx.Commit()
```


Status
------

* Golang >= 1.6.2



License
-------

GNU GENERAL PUBLIC LICENSE

Copyright (C) 2015-2016 niflheims-io 
