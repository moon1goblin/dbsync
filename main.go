package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"time"
	_ "github.com/mattn/go-sqlite3"
)

type DBLogger struct {
	encoder json.Encoder
	db_ptr *sql.DB
}

type dbl_query struct {
	Timestamp_unix int64
	Query string
	Args []string
}

func (dbl DBLogger) Exec(query string, args ...any) (sql.Result, error) {
	query_args := make([]string, len(args))
	for index, value := range args {
		// lets fucking pray this works
		query_args[index] = fmt.Sprintf("%v", value)
	}
	if err := dbl.encoder.Encode(dbl_query{
		Timestamp_unix: time.Now().Unix(),
		Query: query,
		Args: query_args,
	}); err != nil {
		return nil, err
	}
	return dbl.db_ptr.Exec(query, args...)
}

func PlayQueries(db_ptr *sql.DB, logfile *os.File) error {
	decoder := json.NewDecoder(logfile)
	var (
		dblq dbl_query 
		my_errs error
	)
	for {
		if err := decoder.Decode(&dblq); err == io.EOF {
			break
		} else if err != nil {
			return err
		}
		if _, err := db_ptr.Exec(dblq.Query, dblq.Args); err != nil {
			my_errs = errors.Join(my_errs, err)
		}
	}
	if err := os.Truncate(logfile.Name(), 0); err != nil {
		my_errs = errors.Join(my_errs, err)
	}
	return my_errs
}

type event struct {
	Id int
	Name string
	Timestamp sql.NullInt64
}

func main() {
	dsn := "db"
	db_ptr, err := sql.Open("sqlite3", dsn)
	if err != nil {
		log.Fatal(err)
	}
	defer db_ptr.Close()

	logfile, err := os.OpenFile(
		"exec.log",
		os.O_APPEND|os.O_CREATE|os.O_RDWR,
		0666,
	)
	if err != nil {
		log.Fatal(err)
	}

	dbl := DBLogger{*json.NewEncoder(logfile), db_ptr}

	// dont log the db creation

	if _, err := db_ptr.Exec(
		`
		CREATE TABLE IF NOT EXISTS main (
			id INTEGER PRIMARY KEY
			, name TEXT NOT NULL
			, time INTEGER
		);
		`,
	); err != nil {
		log.Fatal(err)
	}

	_, err = dbl.Exec(`INSERT INTO main (name, time) VALUES (?, ?);`, "moon1goblin", nil)

	//--------------------

	// play the queries from log

	if err := PlayQueries(db_ptr, logfile); err != nil {
		log.Fatal(err)
	}

	rows, err := db_ptr.Query(`SELECT * FROM main`)
	if err != nil {
		log.Fatal(err)
	}

	my_event := event{}
	for rows.Next() {
		if err := rows.Scan(&my_event.Id, &my_event.Name, &my_event.Timestamp); err != nil {
			log.Fatal(err)
		}
		fmt.Println(my_event)
	}
}
