package dbhandler

import (
	"strconv"
)

type Group struct {
	Id     int
	Title  string
	GameId int
}

func InsertGp(groupId int, title string) (*Group, error) {
	if db == nil {
		return nil, ErrDBNotInitialized
	}

	stmt, err := db.Prepare("INSERT INTO Groups(Id, Title) VALUES(?,?)")
	if err != nil {
		return nil, err
	}

	res, err := stmt.Exec(groupId, title)
	if err != nil {
		return nil, err
	}

	id, err := res.LastInsertId()
	return &Group{int(id), title, NIL}, err
}

func ExistsGp(groupId int) (bool, error) {
	if db == nil {
		return false, ErrDBNotInitialized
	}

	rows, err := db.Query("SELECT * FROM Groups WHERE Id=" + strconv.Itoa(groupId))
	if err != nil {
		return false, err
	}
	defer rows.Close()
	return rows.Next(), nil
}

func GetGp(groupId int) (*Group, error) {
	if db == nil {
		return nil, ErrDBNotInitialized
	}

	rows, err := db.Query("SELECT * FROM Groups WHERE Id=" + strconv.Itoa(groupId))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	group := &Group{}
	rows.Next()
	err = rows.Scan(&group.Id, &group.Title, &group.GameId)
	if err != nil {
		return nil, err
	}

	return group, nil
}

func UpdateGp(group *Group) error {
	if db == nil {
		return ErrDBNotInitialized
	}

	stmt, err := db.Prepare("UPDATE Groups SET Title=?, GameId=? WHERE Id=?")
	if err != nil {
		return err
	}

	res, err := stmt.Exec(group.Title, group.GameId, group.Id)
	if err != nil {
		return err
	}

	_, err = res.RowsAffected()
	return err
}
