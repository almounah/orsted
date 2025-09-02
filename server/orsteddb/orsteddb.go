package orsteddb

import (
	"log"

	"github.com/boltdb/bolt"

)

func Initialise() *bolt.DB {
	db, err := bolt.Open("my.db", 0600, nil)
	if err != nil {
		log.Fatal(err)
	}
	return db
}

