package dbsync_test

import (
	"database/sql"
	"log"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

type Operation = int

const(
	Add Operation = iota + 1
	Remove
)

type patch struct{
	// timestamp int64
	Operation Operation
	Value int
}

var testcases = []struct{
	Name string
	Server_patches []patch
	Client_patches []patch
	ExpectedRes []int
}{
	{
		"first test",
		[]patch{{Add, 1}},
		[]patch{{Remove, 1}},
		[]int{},
	},
	{
		"double remove",
		[]patch{{Add, 1}, {Remove, 1}},
		[]patch{{Remove, 1}},
		[]int{},
	},
	{
		"double create",
		[]patch{{Add, 1}},
		[]patch{{Add, 1}},
		[]int{1},
	},
	{
		"reversed double remove",
		[]patch{{Remove, 1}},
		[]patch{{Add, 1}, {Remove, 1}},
		[]int{},
	},
}

func runPatches(mypatches *[]patch, db_ptr *sql.DB) {
	for _, mypatch := range *mypatches {
		switch mypatch.Operation {
		case Add:
			if _, err := db_ptr.Exec(
				`
				INSERT INTO main (hello)
				SELECT (?)
				WHERE NOT EXISTS(
					SELECT 1 FROM main WHERE
					hello=?
				)
				`,
				mypatch.Value,
				mypatch.Value,
			); err != nil {
				log.Fatal(err)
			}
		case Remove:
			if _, err := db_ptr.Exec(
				`
				DELETE FROM main 
				WHERE hello=?;
				`,
				mypatch.Value,
			); err != nil {
				log.Fatal(err)
			}
		}
	}
}

// i wanna make servers patches
// and our patches
// and run them and see if result is correct
func TestEverthingDuh(t *testing.T) {
	for _, tc := range testcases {
		t.Run(tc.Name, func(t *testing.T) {
			db_ptr, err := sql.Open("sqlite3", "db")
			if err != nil {
				log.Fatal(err)
			}
			if _, err := db_ptr.Exec(
				`
				CREATE TABLE main (
					hello INTEGER
				);
				`,
			); err != nil {
				log.Fatal(err)
			}

			runPatches(&tc.Server_patches, db_ptr)
			runPatches(&tc.Client_patches, db_ptr)

			rows, err := db_ptr.Query(`SELECT * FROM main`)
			if err != nil {
				log.Fatal(err)
			}

			var dummyint sql.NullInt32
			ints := []int{}
			for rows.Next() {
				if err := rows.Scan(&dummyint); err != nil {
					log.Fatal(err)
				}
				if dummyint.Valid {
					ints = append(ints, int(dummyint.Int32))
				}
			}

			assert := assert.New(t)
			assert.Equal(tc.ExpectedRes, ints)

			db_ptr.Close()
			if err := os.Remove("db"); err != nil {
				log.Fatal(err)
			}
		})
	}
}
