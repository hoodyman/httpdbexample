package db

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

var dbPool *pgxpool.Pool

type DbConnection struct {
	conn *pgxpool.Conn
}

type DbTableData struct {
	Id    int
	Value string
}

const (
	dbTableName   = "datatable"
	dbIdColumn    = "id"
	dbValueColumn = "value"
)

func InitDb() error {
	const url = "postgres://testdb_adm:123@localhost:5432/testdb"
	var err error
	dbPool, err = pgxpool.Connect(context.Background(), url)
	if err != nil {
		return errors.New("init db pool error:" + err.Error())
	}
	return nil
}

func CloseDb() {
	dbPool.Close()
}

func AcquireConn() (DbConnection, error) {
	conn, err := dbPool.Acquire(context.Background())
	if err != nil {
		return DbConnection{}, fmt.Errorf("db acquire conn error: %v", err.Error())
	}
	return DbConnection{conn}, nil
}

func (dbconn *DbConnection) Release() {
	dbconn.conn.Release()
}

func (dbconn *DbConnection) Conn() *pgxpool.Conn {
	return dbconn.conn
}

func (dbconn *DbConnection) IsTableExist() (bool, error) {
	const query = "SELECT tablename FROM pg_tables WHERE tablename='" + dbTableName + "'"
	r := dbconn.conn.QueryRow(context.Background(),
		query)
	var x string
	err := r.Scan(&x)
	if err == nil {
		return true, nil
	} else if err == pgx.ErrNoRows {
		return false, nil
	} else {
		return false, err
	}
}

func (dbconn *DbConnection) CreateTable() error {
	const query = "CREATE TABLE " +
		dbTableName + " (" +
		dbIdColumn + " SERIAL PRIMARY KEY, " +
		dbValueColumn + " text)"
	_, err := dbconn.conn.Exec(context.Background(),
		query)
	return err
}

func (dbconn *DbConnection) DeleteTable() error {
	const query = "DROP TABLE " + dbTableName
	_, err := dbconn.conn.Exec(context.Background(),
		query)
	return err
}

func (dbconn *DbConnection) AppendData(data DbTableData) error {
	const query = "INSERT INTO " +
		dbTableName + " (" +
		dbValueColumn + ") VALUES ($1)"
	_, err := dbconn.conn.Exec(context.Background(),
		query,
		data.Value)
	return err
}

func (dbconn *DbConnection) DeleteData(data DbTableData) error {
	const query = "DELETE FROM " + dbTableName + " WHERE " + dbIdColumn + "=$1"
	_, err := dbconn.conn.Exec(context.Background(),
		query,
		strconv.Itoa(data.Id))
	return err
}

func (dbconn *DbConnection) GetDataScanner() (DbTableDataScanner, error) {
	ds := DbTableDataScanner{}
	const query = "SELECT * FROM " + dbTableName
	var err error
	ds.pgxRows, err = dbconn.conn.Query(context.Background(),
		query)
	return ds, err
}

type DbTableDataScanner struct {
	pgxRows pgx.Rows
}

func (ds *DbTableDataScanner) Scan() (DbTableData, error) {
	td := DbTableData{}
	if ds.pgxRows.Next() {
		err := ds.pgxRows.Scan(&td.Id, &td.Value)
		return td, err
	} else {
		return td, pgx.ErrNoRows
	}
}
