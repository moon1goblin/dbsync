package dbsync

import (
	"database/sql"
	"encoding/json"
	"errors"
	"io"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type DBLogger struct {
	Encoder json.Encoder
	Db_ptr *sql.DB
}

var Dbl_g DBLogger

type dbl_query struct {
	Timestamp_unix int64
	Query string
	Args []any
}

func (dbl DBLogger) Exec(onlylog bool, query string, args ...any) (sql.Result, error) {
	if err := dbl.Encoder.Encode(dbl_query{
		Timestamp_unix: time.Now().Unix(),
		Query: query,
		Args: args,
	}); err != nil {
		return nil, err
	}
	if onlylog {
		return nil, nil
	}
	return dbl.Db_ptr.Exec(query, args...)
}

func PlayQueries(db_ptr *sql.DB, reader io.Reader) error {
	decoder := json.NewDecoder(reader)
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
		if _, err := db_ptr.Exec(dblq.Query, dblq.Args...); err != nil {
			my_errs = errors.Join(my_errs, err)
		}
	}
	// if err := os.Truncate(logfile.Name(), 0); err != nil {
	// 	log.Fatal(err)
	// }
	return my_errs
}

func DBAdd(value int) error {
	_, err := Dbl_g.Exec(
		false,
		`
		INSERT INTO main (hello)
		SELECT (?)
		WHERE NOT EXISTS(
			SELECT 1 FROM main WHERE
			hello=?
		)
		`,
		value,
		value,
	)
	return err
}

func DBRemove(value int) error {
	_, err := Dbl_g.Exec(
		false,
		`
		DELETE FROM main 
		WHERE hello=?;
		`,
		value,
	)
	return err
}

// // TODO: store logs in ram until we exit the program
// logfile, err := os.OpenFile(
// 	"exec.log",
// 	os.O_APPEND|os.O_CREATE|os.O_RDWR,
// 	0666,
// )
// if err != nil {
// 	return err
// }
//

func InitDBAndLoggerWith(dsn string, writer io.Writer) error {
	db_ptr, err := sql.Open("sqlite3", dsn)
	if err != nil {
		return err
	}

	Dbl_g = DBLogger{*json.NewEncoder(writer), db_ptr}

	// dont log the db creation
	if _, err := db_ptr.Exec(
		`
		CREATE TABLE IF NOT EXISTS main (
			hello INTEGER NOT NULL
		);
		`,
	); err != nil {
		return err
	}
	return nil
}

func GetEverythingFromDB() ([]int, error) {
	rows, err := Dbl_g.Db_ptr.Query(`SELECT * FROM main`)
	if err != nil {
		return nil, err
	}

	var dummyint sql.NullInt32
	ints := make([]int, 1)
	for rows.Next() {
		if err := rows.Scan(&dummyint); err != nil {
			return ints, err
		}
		if dummyint.Valid {
			ints = append(ints, int(dummyint.Int32))
		}
	}
	return ints, nil
}
