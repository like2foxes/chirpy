package database

import (
	"encoding/json"
	"log"
	"os"
	"sync"
	"time"
)

type DB struct {
	path string
	mux  *sync.RWMutex
}

type DBStructure struct {
	Chirps  []Chirp              `json:"chirps"`
	Users   []User               `json:"users"`
	Revokes map[string]time.Time `json:"revokes"`
}

func NewDB(path string) (*DB, error) {
	db := &DB{
		path: path,
		mux:  &sync.RWMutex{},
	}
	err := db.ensureDb()
	if err != nil {
		return nil, err
	}

	return db, nil
}

func (db *DB) ensureDb() error {
	var file *os.File
	if _, err := os.Stat(db.path); os.IsNotExist(err) {
		file, err = os.Create(db.path)
		if err != nil {
			return err
		}
		dbStruct := DBStructure{
			Chirps: []Chirp{},
			Users:  []User{},
			Revokes: map[string]time.Time{},
		}
		content, err := json.Marshal(dbStruct)
		if err != nil {
			return err
		}
		_, err = file.Write(content)
		if err != nil {
			return err
		}
	} else {
		file, err = os.Open(db.path)
		if err != nil {
			return err
		}
	}
	defer file.Close()
	return nil
}

func (db *DB) loadDB() (DBStructure, error) {
	db.mux.RLock()
	defer db.mux.RUnlock()
	f, err := os.Open(db.path)
	if err != nil {
		log.Println("Error opening file")
		return DBStructure{}, err
	}
	defer f.Close()

	dbStruct := DBStructure{}
	decoder := json.NewDecoder(f)
	err = decoder.Decode(&dbStruct)
	if err != nil {
		log.Println("Error decoding file")
		return DBStructure{}, err
	}

	return dbStruct, nil
}

func (db *DB) writeDb(dbStruct DBStructure) error {
	db.mux.Lock()
	defer db.mux.Unlock()
	updatedDB, err := json.Marshal(dbStruct)
	if err != nil {
		log.Println("Error encoding file")
		return err
	}
	os.WriteFile(db.path, updatedDB, 0666)
	return nil
}

func calculateId[T HasId](data []T) int {
	if len(data) == 0 {
		return 1
	}
	maxId := 1
	for _, item := range data {
		if item.GetId() > maxId {
			maxId = item.GetId()
		}
	}
	return maxId + 1
}

type HasId interface {
	GetId() int
}
