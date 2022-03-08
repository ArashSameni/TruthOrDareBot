package handlers

import (
	"fmt"
	"strconv"

	"github.com/ArashSameni/TruthOrDareBot/dbhandler"
	_ "github.com/mattn/go-sqlite3"
	tele "gopkg.in/telebot.v3"
)

// Inline reply markups
var (
	GameTypeSelector    = &tele.ReplyMarkup{}
	PendingMenuSelector = &tele.ReplyMarkup{}
	ManualGameSelector  = &tele.ReplyMarkup{}
)

// Inline buttons
var (
	BtnAutoGameType   = GameTypeSelector.Data("Auto", "autoGameMode")
	BtnManualGameType = GameTypeSelector.Data("Manual", "manualGameMode")
	BtnStartGame      = GameTypeSelector.Data("StartGame", "startGame")
	BtnJoinGame       = GameTypeSelector.Data("Join", "joinGame")
	BtnLeaveGame      = GameTypeSelector.Data("Leave", "leaveGame")
	BtnEndGame        = GameTypeSelector.Data("End game", "endGame")
	BtnRoll           = GameTypeSelector.Data("Roll", "roll")
)

// Bot response messages
var (
	MsgSuccessGpAdd   = `گروه با موفقیت اضافه شد`
	MsgChatIsNotGp    = `این دستور فقط قابل استفاده در گروه است`
	MsgGpAlreadyExist = `گروه از قبل اضافه شده است`
	MsgGpIsNotAdded   = `گروه اضافه نشده است`
	MsgSelectGameType = `لطفا نوع بازی را مشخص کنید`
	MsgPlayersList    = `بازی با موفقیت ساخته شد.
بازیکنان:
%s`
	MsgAlreadyPlaying   = `یک بازی از قبل در حال اجرا است`
	MsgNoPlayingGame    = `بازی ای در جریان نیست`
	MsgAlreadyInGame    = `شما از قبل در بازی هستید`
	MsgSuccessfulJoin   = `با موفقیت به بازی اضافه شدید`
	MsgSuccessfulLeave  = `با موفقیت از بازی خارج شدید`
	MsgNotInGame        = `شما در بازی نمیباشید`
	MsgStarterCantLeave = `شما سازنده بازی هستید`
	MsgOnlyStarterCan   = `فقط سازنده بازی به این گزینه دسترسی دارد`
	MsgNotEnoughPlayers = `تعداد بازیکنان کافی نمیباشد`
	MsgManualGame       = `خب %s باید از %s بپرسه`
	MsgGameEnded        = `بازی به پایان رسید`
)

const MinimumPlayersForStart = 2

func init() {
	GameTypeSelector.Inline(
		GameTypeSelector.Row(BtnAutoGameType, BtnManualGameType),
	)

	PendingMenuSelector.Inline(
		GameTypeSelector.Row(BtnStartGame),
		GameTypeSelector.Row(BtnJoinGame, BtnLeaveGame),
	)

	ManualGameSelector.Inline(
		GameTypeSelector.Row(BtnRoll),
		GameTypeSelector.Row(BtnEndGame),
	)
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

	player := getPlayer(c)
	game, _ := dbhandler.GetGame(group.GameId)
	game.Reset()
	game.AddUserToGame(player.Id)
	game.WhoStarted = player.Id
	dbhandler.UpdateGame(game)

	c.Reply(MsgSelectGameType, GameTypeSelector)

	return nil
}

func OnAutoGameSelect(c tele.Context) error {
	return handleGameSelect(c, dbhandler.GameTypeAuto)
}

func OnManualGameSelect(c tele.Context) error {
	return handleGameSelect(c, dbhandler.GameTypeManual)
}

func OnJoinGame(c tele.Context) error {
	group, _ := dbhandler.GetGp(int(c.Chat().ID))
	game, _ := dbhandler.GetGame(group.GameId)
	if game.Status == dbhandler.GameStatusFinished {
		notifyPlayer(c, MsgNoPlayingGame)
		return nil
	}

	player := getPlayer(c)
	if isPlaying, _ := game.IsUserPlaying(player.Id); isPlaying {
		notifyPlayer(c, MsgAlreadyInGame)
		return nil
	}

	game.AddUserToGame(player.Id)
	notifyPlayer(c, MsgSuccessfulJoin)

	editPlayersList(c, game)

	return nil
}

func OnLeaveGame(c tele.Context) error {
	group, _ := dbhandler.GetGp(int(c.Chat().ID))
	game, _ := dbhandler.GetGame(group.GameId)
	if game.Status == dbhandler.GameStatusFinished {
		notifyPlayer(c, MsgNoPlayingGame)
		return nil
	}
	player := getPlayer(c)

	if game.WhoStarted == player.Id {
		notifyPlayer(c, MsgStarterCantLeave)
		return nil
	}

	if isPlaying, _ := game.IsUserPlaying(player.Id); !isPlaying {
		notifyPlayer(c, MsgNotInGame)
		return nil
	}

	game.RemoveUserFromGame(player.Id)
	notifyPlayer(c, MsgSuccessfulLeave)

	editPlayersList(c, game)

	return nil
}

func OnStartGame(c tele.Context) error {
	group, _ := dbhandler.GetGp(int(c.Chat().ID))
	game, _ := dbhandler.GetGame(group.GameId)
	if game.Status == dbhandler.GameStatusPlaying {
		notifyPlayer(c, MsgAlreadyPlaying)
		return nil
	}
	if game.Status == dbhandler.GameStatusFinished {
		notifyPlayer(c, MsgNoPlayingGame)
		return nil
	}

	player := getPlayer(c)
	if game.WhoStarted != player.Id {
		notifyPlayer(c, MsgOnlyStarterCan)
		return nil
	}

	if game.PlayersCount() < MinimumPlayersForStart {
		notifyPlayer(c, MsgNotEnoughPlayers)
		return nil
	}

	game.Status = dbhandler.GameStatusPlaying
	dbhandler.UpdateGame(game)

	p1, p2, _ := game.TwoRandomPlayers()

	msg := tele.StoredMessage{ChatID: int64(game.GroupId), MessageID: strconv.Itoa(game.MessageId)}
	c.Bot().Edit(&msg, fmt.Sprintf(MsgManualGame, getNickName(p1), getNickName(p2)), ManualGameSelector)
	if c.Callback() == nil {
		c.Send(fmt.Sprintf(MsgManualGame, getNickName(p1), getNickName(p2)))
	}

	return nil
}

func OnRoll(c tele.Context) error {
	group, _ := dbhandler.GetGp(int(c.Chat().ID))
	game, _ := dbhandler.GetGame(group.GameId)
	if game.Status != dbhandler.GameStatusPlaying {
		notifyPlayer(c, MsgNoPlayingGame)
		return nil
	}

	p1, p2, _ := game.TwoRandomPlayers()

	if c.Callback() != nil {
		c.Edit(fmt.Sprintf(MsgManualGame, getNickName(p1), getNickName(p2)), ManualGameSelector)
	} else {
		c.Send(fmt.Sprintf(MsgManualGame, getNickName(p1), getNickName(p2)))
	}

	return nil
}

func OnEndGame(c tele.Context) error {
	group, _ := dbhandler.GetGp(int(c.Chat().ID))
	game, _ := dbhandler.GetGame(group.GameId)
	if game.Status != dbhandler.GameStatusPlaying {
		notifyPlayer(c, MsgNoPlayingGame)
		return nil
	}

	player := getPlayer(c)
	if game.WhoStarted != player.Id {
		notifyPlayer(c, MsgOnlyStarterCan)
		return nil
	}

	game.Status = dbhandler.GameStatusFinished
	dbhandler.UpdateGame(game)

	msg := tele.StoredMessage{ChatID: int64(game.GroupId), MessageID: strconv.Itoa(game.MessageId)}
	c.Bot().Edit(&msg, MsgGameEnded)
	c.Send(MsgGameEnded)

	return nil
}
