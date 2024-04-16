package storage

import (
	"database/sql"
	"errors"

	_ "github.com/mattn/go-sqlite3"
)

const (
	name   = "twc_gen3"
	driver = "sqlite3"
)

type DB struct {
	db *sql.DB
}

func New() (*DB, error) {
	db, err := sql.Open(driver, "."+name+".db")
	if err != nil {
		return nil, err
	}

	return &DB{
		db: db,
	}, nil
}

func (d *DB) CreateTables() error {
	if _, err := d.db.Exec("create table if not exists twc (ip text primary key)"); err != nil {
		return err
	}
	return nil
}

func (d *DB) GetTWCIP() (string, error) {
	row := d.db.QueryRow("SELECT ip from twc")
	var ip string
	if err := row.Scan(&ip); err != nil {
		return "", err
	}
	if ip == "" {
		return "", errors.New("ip not found in db")
	}
	return ip, nil
}

func (d *DB) SaveTWCIP(ip string) error {
	tx, err := d.db.Prepare("INSERT into twc (ip) values(?)")
	if err != nil {
		return err
	}
	if _, err := tx.Exec(ip); err != nil {
		return err
	}
	return nil
}
