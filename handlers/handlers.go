package handlers

import (
	"fmt"
	"log"
	"strconv"

	"github.com/ArashSameni/TruthOrDareBot/dbhandler"
	_ "github.com/mattn/go-sqlite3"
	tele "gopkg.in/telebot.v3"
)

// Inline reply markups
var (
	GameTypeSelector   = &tele.ReplyMarkup{}
	ManualGameSelector = &tele.ReplyMarkup{}
)

// Inline buttons
var (
	BtnAutoGameType   = GameTypeSelector.Data("Auto", "autoGameMode")
	BtnManualGameType = GameTypeSelector.Data("Manual", "manualGameMode")
	BtnJoinGame       = GameTypeSelector.Data("Join", "joinGame")
	BtnLeaveGame      = GameTypeSelector.Data("Leave", "leaveGame")
	BtnEndGame        = GameTypeSelector.Data("End game", "endGame")
)

// Bot response messages
var (
	MsgSuccessGpAdd   = `گروه با موفقیت اضافه شد`
	MsgChatIsNotGp    = `این دستور فقط قابل استفاده در گروه است`
	MsgGpAlreadyExist = `گروه از قبل اضافه شده است`
	MsgGpIsNotAdded   = `گروه اضافه نشده است`
	MsgSelectGameType = `لطفا نوع بازی را مشخص کنید`
	MsgManualGame     = `بازی با موفقیت ساخته شد.
بازیکنان:
%s`
	MsgAlreadyPlaying       = `یک بازی از قبل در حال اجرا است`
	MsgNoPlayingGame        = `بازی ای در جریان نیست`
	MsgAlreadyInGame        = `شما از قبل در بازی هستید`
	MsgSuccessfulJoin       = `با موفقیت به بازی اضافه شدید`
	MsgSuccessfulLeave      = `با موفقیت از بازی خارج شدید`
	MsgNotInGame            = `شما در بازی نمیباشید`
	MsgStarterCantLeave     = `شما سازنده بازی هستید`
	MsgStarterCanChooseMode = `فقط سازنده بازی میتواند مود را انتخاب کند`
)

func init() {
	GameTypeSelector.Inline(
		GameTypeSelector.Row(BtnAutoGameType, BtnManualGameType),
	)

	ManualGameSelector.Inline(
		GameTypeSelector.Row(BtnJoinGame, BtnLeaveGame),
		GameTypeSelector.Row(BtnEndGame),
	)
}

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

func OnStart(c tele.Context) error {
	c.Send(strconv.Itoa(int(c.Chat().ID)))
	return nil
}

func OnAddGp(c tele.Context) error {
	if !checkIsGp(c) {
		return nil
	}

	chat := c.Chat()
	exist, err := dbhandler.ExistsGp(int(chat.ID))
	if err != nil {
		return err
	}
	if exist {
		c.Reply(MsgGpAlreadyExist)
		return nil
	}
	group, err := dbhandler.InsertGp(int(chat.ID), chat.Title)
	if err != nil {
		return err
	}
	game, err := dbhandler.InsertGame(group.Id, dbhandler.GameTypeManual, dbhandler.GameStatusPending, dbhandler.NIL, dbhandler.NIL)
	if err != nil {
		return err
	}
	group.GameId = game.Id
	dbhandler.UpdateGp(group)

	c.Reply(MsgSuccessGpAdd)
	return nil
}

func OnNewGame(c tele.Context) error {
	if !checkIsGp(c) {
		return nil
	}
	if !checkIsGpAdded(c) {
		return nil
	}
	group, _ := dbhandler.GetGp(int(c.Chat().ID))
	if checkAlreadyPlaying(c, group) {
		return nil
	}

	player, _ := dbhandler.InsertUserIfNotExist(int(c.Sender().ID), getFullName(c.Sender()), c.Sender().Username)

	game, _ := dbhandler.GetGame(group.GameId)
	game.Reset()
	game.AddUserToGame(player.Id)
	game.WhoStarted = player.Id
	dbhandler.UpdateGame(game)

	c.Reply(MsgSelectGameType, GameTypeSelector)

	return nil
}

//TODO: Complete this shit
func OnAutoGameSelect(c tele.Context) error {
	if !checkIsGp(c) {
		return nil
	}
	if !checkIsGpAdded(c) {
		return nil
	}
	group, _ := dbhandler.GetGp(int(c.Chat().ID))
	if checkAlreadyPlaying(c, group) {
		return nil
	}

	return nil
}

func OnManualGameSelect(c tele.Context) error {
	if !checkIsGp(c) {
		return nil
	}
	if !checkIsGpAdded(c) {
		return nil
	}
	group, _ := dbhandler.GetGp(int(c.Chat().ID))
	if checkAlreadyPlaying(c, group) {
		return nil
	}

	player, _ := dbhandler.InsertUserIfNotExist(int(c.Callback().Sender.ID), getFullName(c.Callback().Sender), c.Callback().Sender.Username)

	game, _ := dbhandler.GetGame(group.GameId)

	if game.WhoStarted != player.Id {
		c.Respond(&tele.CallbackResponse{Text: MsgStarterCanChooseMode})
		return nil
	}

	game.Type = dbhandler.GameTypeManual
	dbhandler.UpdateGame(game)

	c.Edit(fmt.Sprintf(MsgManualGame, getNickName(player)), ManualGameSelector)

	return nil
}

func OnJoinGame(c tele.Context) error {
	group, _ := dbhandler.GetGp(int(c.Chat().ID))
	game, _ := dbhandler.GetGame(group.GameId)
	if game.Status == dbhandler.GameStatusFinished {
		c.Respond(&tele.CallbackResponse{Text: MsgNoPlayingGame})
		return nil
	}
	player, _ := dbhandler.InsertUserIfNotExist(int(c.Callback().Sender.ID), getFullName(c.Callback().Sender), c.Callback().Sender.Username)

	if isPlaying, _ := game.IsUserPlaying(player.Id); isPlaying {
		c.Respond(&tele.CallbackResponse{Text: MsgAlreadyInGame})
		return nil
	}

	game.AddUserToGame(player.Id)
	c.Respond(&tele.CallbackResponse{Text: MsgSuccessfulJoin})

	editPlayersList(c, game)

	return nil
}

func OnLeaveGame(c tele.Context) error {
	group, _ := dbhandler.GetGp(int(c.Chat().ID))
	game, _ := dbhandler.GetGame(group.GameId)
	if game.Status == dbhandler.GameStatusFinished {
		c.Respond(&tele.CallbackResponse{Text: MsgNoPlayingGame})
		return nil
	}
	player, _ := dbhandler.InsertUserIfNotExist(int(c.Callback().Sender.ID), getFullName(c.Callback().Sender), c.Callback().Sender.Username)

	if game.WhoStarted == player.Id {
		c.Respond(&tele.CallbackResponse{Text: MsgStarterCantLeave})
		return nil
	}

	if isPlaying, _ := game.IsUserPlaying(player.Id); !isPlaying {
		c.Respond(&tele.CallbackResponse{Text: MsgNotInGame})
		return nil
	}

	game.RemoveUserFromGame(player.Id)
	c.Respond(&tele.CallbackResponse{Text: MsgSuccessfulLeave})

	editPlayersList(c, game)

	return nil
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
	if game.Status == dbhandler.GameStatusPending {
		players, err := game.Players()
		if err != nil {
			return err
		}

		var playersString string
		for _, p := range players {
			playersString += getNickName(p) + "\n"
		}
		c.Edit(fmt.Sprintf(MsgManualGame, playersString), ManualGameSelector)
	}
	return nil
}
