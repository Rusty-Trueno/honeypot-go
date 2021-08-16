/*

this is a golang timeseries driver restful

package main

import (
	"database/sql"
	"log"

	taossqlrestful "taossqlrestful/taosSqlRestful"
)

func main() {
	db, err := sql.Open("taossqlrestful", "mydb://dalong@127.0.0.1/demoapp")
	if err != nil {
		log.Fatalf("some error %s", err.Error())
	}
	rows, err := db.Query("select * from demoapp")
	if err != nil {
		log.Println("some wrong for query", err.Error())
	}
	for rows.Next() {
		rows.Scan()
	}
}
*/

package taossqlrestful
