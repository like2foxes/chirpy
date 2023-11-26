package database

import (
	"errors"
	"log"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	Id       int    `json:"id"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (u User) GetId() int {
	return u.Id
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
