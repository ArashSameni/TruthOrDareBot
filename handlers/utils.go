package handlers

import (
	"fmt"
	"log"
	"strconv"

	"github.com/ArashSameni/TruthOrDareBot/dbhandler"
	tele "gopkg.in/telebot.v3"
)

func checkIsGp(c tele.Context) bool {
	if c.Chat().Type != tele.ChatGroup && c.Chat().Type != tele.ChatSuperGroup {
		c.Reply(MsgChatIsNotGp)
		return false
	}
	return true
}

func checkIsGpAdded(c tele.Context) bool {
	chat := c.Chat()
	exist, err := dbhandler.ExistsGp(int(chat.ID))
	if err != nil {
		log.Println(err)
	}
	if !exist {
		c.Reply(MsgGpIsNotAdded)
		return false
	}
	return true
}

func checkAlreadyPlaying(c tele.Context, group *dbhandler.Group) bool {
	game, _ := dbhandler.GetGame(group.GameId)
	if game.Status == dbhandler.GameStatusPlaying {
		c.Reply(MsgAlreadyPlaying)
		return true
	}
	return false
}

func getFullName(u *tele.User) string {
	fullName := u.FirstName
	if u.LastName != "" {
		fullName += " " + u.LastName
	}
	return fullName
}

func getNickName(u *dbhandler.User) string {
	if u.Username != "" {
		return "@" + u.Username
	}
	return u.Fullname
}

func editPlayersList(c tele.Context, game *dbhandler.Game) error {
	msg := tele.StoredMessage{ChatID: int64(game.GroupId), MessageID: strconv.Itoa(game.MessageId)}
	if game.Status == dbhandler.GameStatusPending {
		players, err := game.Players()
		if err != nil {
			return err
		}

		var playersString string
		for _, p := range players {
			playersString += getNickName(p) + "\n"
		}
		c.Bot().Edit(&msg, fmt.Sprintf(MsgPlayersList, playersString), PendingMenuSelector)
	}

	return nil
}

func notifyPlayer(c tele.Context, msg string) error {
	if c.Callback() == nil {
		return c.Reply(msg)
	}

	return c.Respond(&tele.CallbackResponse{Text: msg})
}

func getPlayer(c tele.Context) *dbhandler.User {
	var player *dbhandler.User
	if c.Callback() == nil {
		player, _ = dbhandler.InsertUserIfNotExist(int(c.Sender().ID), getFullName(c.Sender()), c.Sender().Username)
	} else {
		player, _ = dbhandler.InsertUserIfNotExist(int(c.Callback().Sender.ID), getFullName(c.Callback().Sender), c.Callback().Sender.Username)
	}
	return player
}
