package dbhandler

import (
	"strconv"
)

type User struct {
	Id       int
	Fullname string
	Username string
}

func InsertUser(userId int, fullname, username string) (*User, error) {
	if db == nil {
		return nil, ErrDBNotInitialized
	}

	stmt, err := db.Prepare("INSERT INTO Users(Id, Fullname, Username) VALUES(?,?,?)")
	if err != nil {
		return nil, err
	}

	res, err := stmt.Exec(userId, fullname, username)
	if err != nil {
		return nil, err
	}

	id, err := res.LastInsertId()
	return &User{int(id), fullname, username}, err
}

func ExistsUser(userId int) (bool, error) {
	if db == nil {
		return false, ErrDBNotInitialized
	}

	rows, err := db.Query("SELECT * FROM Users WHERE Id=" + strconv.Itoa(userId))
	if err != nil {
		return false, err
	}
	defer rows.Close()
	return rows.Next(), nil
}

func GetUser(userId int) (*User, error) {
	if db == nil {
		return nil, ErrDBNotInitialized
	}

	rows, err := db.Query("SELECT * FROM Users WHERE Id=" + strconv.Itoa(userId))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	user := &User{}
	rows.Next()
	err = rows.Scan(&user.Id, &user.Fullname, &user.Username)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func UpdateUser(user *User) error {
	if db == nil {
		return ErrDBNotInitialized
	}

	stmt, err := db.Prepare("UPDATE Users SET Fullname=?, Username=? WHERE Id=?")
	if err != nil {
		return err
	}

	res, err := stmt.Exec(user.Fullname, user.Username, user.Id)
	if err != nil {
		return err
	}

	_, err = res.RowsAffected()
	return err
}

func InsertUserIfNotExist(userId int, fullname, username string) (*User, error) {
	exists, err := ExistsUser(userId)
	if err != nil {
		return nil, err
	}
	if exists {
		return GetUser(userId)
	}

	return InsertUser(userId, fullname, username)
}
