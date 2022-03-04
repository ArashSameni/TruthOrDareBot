package dbhandler

import (
	"strconv"
)

// Game Status
const (
	GameStatusPending = iota
	GameStatusPlaying
	GameStatusFinished
)

// Game Type
const (
	GameTypeAuto = iota
	GameTypeManual
)

type Game struct {
	Id            int
	GroupId       int
	Type          int
	Status        int
	WhoStarted    int
	CurrentPlayer int
	MessageId     int
}

func InsertGame(groupId, gameType, status, whoStarted, currentPlayer int) (*Game, error) {
	if db == nil {
		return nil, ErrDBNotInitialized
	}

	stmt, err := db.Prepare("INSERT INTO Games(GroupId, Type, Status, WhoStarted, CurrentPlayer) VALUES(?,?,?,?,?)")
	if err != nil {
		return nil, err
	}

	res, err := stmt.Exec(groupId, gameType, status, whoStarted, currentPlayer)
	if err != nil {
		return nil, err
	}

	id, err := res.LastInsertId()
	return &Game{int(id), groupId, gameType, status, whoStarted, currentPlayer, NIL}, err
}

func ExistsGame(gameId int) (bool, error) {
	if db == nil {
		return false, ErrDBNotInitialized
	}

	rows, err := db.Query("SELECT * FROM Games WHERE Id=" + strconv.Itoa(gameId))
	if err != nil {
		return false, err
	}
	defer rows.Close()
	return rows.Next(), nil
}

func GetGame(gameId int) (*Game, error) {
	if db == nil {
		return nil, ErrDBNotInitialized
	}

	rows, err := db.Query("SELECT * FROM Games WHERE Id=" + strconv.Itoa(gameId))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	game := &Game{}
	rows.Next()
	err = rows.Scan(&game.Id, &game.GroupId, &game.Type, &game.Status, &game.WhoStarted, &game.CurrentPlayer, &game.MessageId)
	if err != nil {
		return nil, err
	}

	return game, nil
}

func UpdateGame(game *Game) error {
	if db == nil {
		return ErrDBNotInitialized
	}

	stmt, err := db.Prepare("UPDATE Games SET GroupId=?, Type=?, Status=?, WhoStarted=?, CurrentPlayer=?, MessageId=? WHERE Id=?")
	if err != nil {
		return err
	}

	res, err := stmt.Exec(game.GroupId, game.Type, game.Status, game.WhoStarted, game.CurrentPlayer, game.MessageId, game.Id)
	if err != nil {
		return err
	}

	_, err = res.RowsAffected()
	return err
}

func (g *Game) Reset() error {
	g.Status = GameStatusPending
	g.WhoStarted = NIL
	g.CurrentPlayer = NIL
	g.MessageId = NIL
	err := UpdateGame(g)
	if err != nil {
		return err
	}

	stmt, err := db.Prepare("DELETE FROM GamesUsers WHERE GameId=?")
	if err != nil {
		return err
	}

	_, err = stmt.Exec(g.Id)
	if err != nil {
		return err
	}

	return nil
}

func (g *Game) IsUserPlaying(userId int) (bool, error) {
	if db == nil {
		return false, ErrDBNotInitialized
	}

	rows, err := db.Query("SELECT * FROM GamesUsers WHERE GameId=" + strconv.Itoa(g.Id) + " AND UserId=" + strconv.Itoa(userId))
	if err != nil {
		return false, err
	}
	defer rows.Close()
	return rows.Next(), nil
}

func (g *Game) AddUserToGame(userId int) error {
	if db == nil {
		return ErrDBNotInitialized
	}

	stmt, err := db.Prepare("INSERT INTO GamesUsers(GameId, UserId) VALUES(?,?)")
	if err != nil {
		return err
	}

	res, err := stmt.Exec(g.Id, userId)
	if err != nil {
		return err
	}

	_, err = res.LastInsertId()
	return err
}

func (g *Game) RemoveUserFromGame(userId int) error {
	stmt, err := db.Prepare("DELETE FROM GamesUsers WHERE GameId=? AND UserId=?")
	if err != nil {
		return err
	}

	_, err = stmt.Exec(g.Id, userId)
	if err != nil {
		return err
	}

	return nil
}

func (g *Game) Players() (players []*User, err error) {
	rows, err := db.Query("SELECT UserId FROM GamesUsers WHERE GameId=" + strconv.Itoa(g.Id))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var playerId int
		err = rows.Scan(&playerId)
		if err != nil {
			return nil, err
		}

		player, err := GetUser(playerId)
		if err != nil {
			return nil, err
		}

		players = append(players, player)
	}

	return players, nil
}

func (g *Game) PlayersCount() int {
	rows, err := db.Query("SELECT COUNT(UserId) FROM GamesUsers WHERE GameId=" + strconv.Itoa(g.Id))
	if err != nil {
		return NIL
	}
	defer rows.Close()

	var count int
	rows.Next()
	err = rows.Scan(&count)
	if err != nil {
		return NIL
	}

	return count
}

func (g *Game) TwoRandomPlayers() (a *User, b *User, err error) {
	rows, err := db.Query("SELECT UserId FROM GamesUsers WHERE GameId=" + strconv.Itoa(g.Id) + " ORDER BY RANDOM() LIMIT 2")
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var playerId int
		err = rows.Scan(&playerId)
		if err != nil {
			return nil, nil, err
		}

		player, err := GetUser(playerId)
		if err != nil {
			return nil, nil, err
		}

		if a == nil {
			a = player
		} else {
			b = player
		}
	}

	return a, b, nil
}
