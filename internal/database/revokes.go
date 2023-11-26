package database

import (
	"time"
	"log"
)

func (db *DB) RevokeToken(token string) error {
	dbStruct, err := db.loadDB()
	if err != nil {
		return err
	}
	dbStruct.Revokes[token] = time.Now()
	err = db.writeDb(dbStruct)
	if err != nil {
		return err
	}
	return nil
}

func (db *DB) IsTokenRevoked(token string) bool {
	dbStruct, err := db.loadDB()
	if err != nil {
		log.Println("Error loading db")
		return true
	}
	_, ok := dbStruct.Revokes[token]
	return ok
}

