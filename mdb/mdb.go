package mdb

import (
	"database/sql"
	"errors"
	"github.com/mattn/go-sqlite3"
	"log"
	"time"
)

type EmailEntry struct {
	Id           int64
	EmailAddress string
	ConfirmedAt  *time.Time
	OptOut       bool
}

func TryCreate(db *sql.DB) {
	_, err := db.Exec(`
	CREATE TABLE IF NOT EXISTS email_entries (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    email TEXT UNIQUE,
    confirmed_at INTEGER,
    opt_out INTEGER
	);`)

	if err != nil {
		var sqlError sqlite3.Error
		if errors.As(err, &sqlError) {
			//code 1 means table already exists
			if sqlError.Code != 1 {
				log.Fatal(sqlError)
			} else {
				log.Fatal(err)
			}
		}
	}
}

func emailEntryFromRow(row *sql.Rows) (*EmailEntry, error) {
	var id int64
	var emailAddress string
	var confirmedAt int64
	var optOut bool

	err := row.Scan(&id, &emailAddress, &confirmedAt, &optOut)

	if err != nil {
		log.Println(err)
		return nil, err
	}

	t := time.Unix(confirmedAt, 0)
	return &EmailEntry{Id: id, EmailAddress: emailAddress, ConfirmedAt: &t, OptOut: optOut}, nil
}

func CreateEmail(db *sql.DB, email string) error {
	_, err := db.Exec(`INSERT INTO email_entries (email, confirmed_at, opt_out) VALUES (?, 0, false)`, email)

	if err != nil {
		log.Println(err)
		return err
	}
	return nil
}

func GetEmail(db *sql.DB, email string) (*EmailEntry, error) {
	rows, err := db.Query(`SELECT * FROM email_entries WHERE email = ?`, email)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	defer func(rows *sql.Rows) {
		err := rows.Close()
		if err != nil {
			log.Println(err)
		}
	}(rows)

	for rows.Next() {
		return emailEntryFromRow(rows)
	}
	return nil, nil
}

func UpdateEmail(db *sql.DB, entry EmailEntry) error {
	t := entry.ConfirmedAt.Unix()

	_, err := db.Exec(`
		INSERT INTO email_entries(email, confirmed_at, opt_out) VALUES (?, ?, ?) 
		ON CONFLICT(email) DO UPDATE SET confirmed_at = ?, opt_out = ?`,
		entry.EmailAddress, t, entry.OptOut, t, entry.ConfirmedAt)
	if err != nil {
		log.Println(err)
		return err
	}
	return nil
}

func DeleteEmail(db *sql.DB, email string) error {
	_, err := db.Exec(`UPDATE email_entries SET opt_out = true WHERE email = ?`, email)
	if err != nil {
		log.Println(err)
		return err
	}
	return nil
}

type GetEmailPaginatedQueryParams struct {
	Page  int
	Count int
}

func GetEmailPaginated(db *sql.DB, params GetEmailPaginatedQueryParams) ([]EmailEntry, error) {
	var empty []EmailEntry

	rows, err := db.Query(`
			SELECT id, email, confirmed_at, opt_out FROM email_entries 
        	WHERE opt_out = FALSE ORDER BY id DESC LIMIT ? OFFSET ?`, params.Count, (params.Page-1)*params.Count,
	)

	if err != nil {
		log.Println(err)
		return empty, err
	}

	defer func(rows *sql.Rows) {
		err := rows.Close()
		if err != nil {
			log.Println(err)
		}
	}(rows)

	emails := make([]EmailEntry, 0, params.Count)

	for rows.Next() {
		email, err := emailEntryFromRow(rows)
		if err != nil {
			log.Println(err)
			return nil, err
		}
		emails = append(emails, *email)
	}
	return emails, nil
}
