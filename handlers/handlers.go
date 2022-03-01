package handlers

import (
	"fmt"
	"strconv"

	tele "gopkg.in/telebot.v3"
)

var MustJoin []int64

func OnStart(c tele.Context) error {
	c.Send(strconv.Itoa(int(c.Chat().ID)))
	return nil
}

func OnChannelPost(c tele.Context) error {
	fmt.Println(c.ChatMember().Sender.FirstName)
	return nil
}
