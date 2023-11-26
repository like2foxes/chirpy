package database

import (
	"errors"
)

type Chirp struct {
	Id       int    `json:"id"`
	Body     string `json:"body"`
	AuthorId int    `json:"author_id"`
}

func (c Chirp) GetId() int {
	return c.Id
}

func (db *DB) CreateChirp(body string, authorId int) (Chirp, error) {
	dbStruct, err := db.loadDB()
	if err != nil {
		return Chirp{}, err
	}

	id := calculateId(dbStruct.Chirps)
	chirp := Chirp{
		Id:   id,
		Body: body,
		AuthorId: authorId,
	}
	dbStruct.Chirps = append(dbStruct.Chirps, chirp)
	err = db.writeDb(dbStruct)
	if err != nil {
		return Chirp{}, err
	}

	return chirp, nil
}

func (db *DB) GetChirps() ([]Chirp, error) {
	data, err := db.loadDB()
	if err != nil {
		return nil, err
	}
	return data.Chirps, nil
}

func (db *DB) GetChirp(id int) (Chirp, error) {
	data, err := db.loadDB()
	if err != nil {
		return Chirp{}, err
	}
	for _, chirp := range data.Chirps {
		if chirp.Id == id {
			return chirp, nil
		}
	}
	return Chirp{}, errors.New("not found")
}
