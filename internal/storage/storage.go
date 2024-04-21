package storage

import (
	"database/sql"
	"encoding/json"
	"errors"
	"time"

	"github.com/0verbyte/twc-gen3/pkg/twc"
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

func (d *DB) Init() error {
	if _, err := d.db.Exec("create table if not exists twc (ip text primary key)"); err != nil {
		return err
	}

	tableTwcVitals := `
create table if not exists twc_vitals (
	id integer primary key,
	ip text,
	vital blob,
	timestamp numeric
)`
	if _, err := d.db.Exec(tableTwcVitals); err != nil {
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

func (d *DB) RecordVital(ip string, vital *twc.Vital) error {
	b, err := json.Marshal(vital)
	if err != nil {
		return err
	}

	tx, err := d.db.Prepare("INSERT into twc_vitals(ip, vital, timestamp) values(?,?,?)")
	if err != nil {
		return err
	}
	if _, err := tx.Exec(ip, b, time.Now()); err != nil {
		return err
	}
	return nil
}

func (d *DB) QueryVitals(startTime time.Time) ([]*twc.VitalQueryResponse, error) {
	stmt, err := d.db.Prepare("SELECT timestamp, vital from twc_vitals where timestamp > ?")
	if err != nil {
		return nil, err
	}
	rows, err := stmt.Query(startTime)
	if err != nil {
		return nil, err
	}

	var vitals []*twc.VitalQueryResponse
	for rows.Next() {
		var (
			timestamp string
			vital     []byte
		)

		if err := rows.Scan(&timestamp, &vital); err != nil {
			return nil, err
		}

		v := &twc.Vital{}
		if err := json.Unmarshal(vital, v); err != nil {
			return nil, err
		}

		vitals = append(vitals, &twc.VitalQueryResponse{Timestamp: timestamp, Vital: v})
	}

	return vitals, nil
}
