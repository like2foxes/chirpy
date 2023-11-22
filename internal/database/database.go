package database

import (
	"encoding/json"
	"errors"
	"log"
	"os"
	"sync"

	"golang.org/x/crypto/bcrypt"
)

type DB struct {
	path string
	mux  *sync.RWMutex
}

type DBStructure struct {
	Chirps []Chirp `json:"chirps"`
	Users  []User  `json:"users"`
}

type Chirp struct {
	Id   int    `json:"id"`
	Body string `json:"body"`
}

type User struct {
	Id       int    `json:"id"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (c Chirp) GetId() int {
	return c.Id
}

func (u User) GetId() int {
	return u.Id
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

func (db *DB) CreateUser(email string, password string) (User, error) {
	dbStruct, err := db.loadDB()
	if err != nil {
		log.Println("Error loading db")
		return User{}, err
	}
	_, err = db.GetUserByEmail(email)
	if err == nil {
		log.Println("a user with that email already exists")
		return User{}, errors.New("a user with that email already exists")
	}
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Println("Error hashing password")
		return User{}, err
	}
	id := calculateId(dbStruct.Users)
	user := User{
		Id:       id,
		Email:    email,
		Password: string(hashed),
	}
	dbStruct.Users = append(dbStruct.Users, user)
	err = db.writeDb(dbStruct)
	if err != nil {
		log.Println("Error writing db")
		return User{}, err
	}
	return user, nil
}

func (db *DB) UpdateUser(user User) (User, error) {
	dbStruct, err := db.loadDB()
	if err != nil {
		return User{}, err
	}

	for i, dbUser := range dbStruct.Users {
		if dbUser.Id == user.Id {
			hashed, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
			if err != nil {
				return User{}, err
			}
			user.Password = string(hashed)
			dbStruct.Users[i] = user
			err = db.writeDb(dbStruct)
			if err != nil {
				return User{}, err
			}
			return dbStruct.Users[i], nil
		}
	}
	if err != nil {
		return User{}, err
	}
	return User{}, errors.New("user does not exist")
}

func (db *DB) GetUserByEmail(email string) (User, error) {
	dbStruct, err := db.loadDB()
	if err != nil {
		return User{}, err
	}
	for _, user := range dbStruct.Users {
		if user.Email == email {
			return user, nil
		}
	}
	return User{}, errors.New("not found")
}

func (db *DB) PrintUsers() {
	dbStruct, err := db.loadDB()
	if err != nil {
		log.Println("Error loading db")
		return
	}
	for _, user := range dbStruct.Users {
		log.Println(user.Id)
		log.Println(user.Email)
		log.Println(user.Password)
		log.Println("----")
	}
}

func (db *DB) GetUser(id int) (User, error) {
	dbStruct, err := db.loadDB()
	if err != nil {
		return User{}, err
	}
	for _, user := range dbStruct.Users {
		if user.Id == id {
			return user, nil
		}
	}
	return User{}, errors.New("not found")
}

func (db *DB) GetUsers() ([]User, error) {
	dbStruct, err := db.loadDB()
	if err != nil {
		return nil, err
	}
	return dbStruct.Users, nil
}

func (db *DB) CreateChirp(body string) (Chirp, error) {
	dbStruct, err := db.loadDB()
	if err != nil {
		return Chirp{}, err
	}

	id := calculateId(dbStruct.Chirps)
	chirp := Chirp{
		Id:   id,
		Body: body,
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

func (db *DB) ensureDb() error {
	var file *os.File
	if _, err := os.Stat(db.path); os.IsNotExist(err) {
		file, err = os.Create(db.path)
		if err != nil {
			return err
		}
		dbStruct := DBStructure{
			Chirps: []Chirp{},
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
