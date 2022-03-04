package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/ArashSameni/TruthOrDareBot/dbhandler"
	"github.com/ArashSameni/TruthOrDareBot/handlers"
	"github.com/joho/godotenv"
	tele "gopkg.in/telebot.v3"
)

func initLog() error {
	logFile := os.Getenv("LOGFILE")
	if logFile != "" {
		f, err := os.OpenFile(logFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			return err
		}

		log.SetOutput(f)
	}
	return nil
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("error loading .env file")
	}

	err = initLog()
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}

	err = dbhandler.InitDB()
	if err != nil {
		log.Fatal(err)
	}

	pref := tele.Settings{
		Token:  os.Getenv("TOKEN"),
		Poller: &tele.LongPoller{Timeout: 10 * time.Second},
	}

	b, err := tele.NewBot(pref)
	if err != nil {
		log.Fatal(err)
	}

	b.Handle("/start", handlers.OnStart)
	b.Handle("/AddGp", handlers.OnAddGp)
	b.Handle("/NewGame", handlers.OnNewGame)
	b.Handle(&handlers.BtnAutoGameType, handlers.OnAutoGameSelect)
	b.Handle(&handlers.BtnManualGameType, handlers.OnManualGameSelect)
	b.Handle(&handlers.BtnStartGame, handlers.OnStartGame)
	b.Handle(&handlers.BtnJoinGame, handlers.OnJoinGame)
	b.Handle(&handlers.BtnLeaveGame, handlers.OnLeaveGame)

	fmt.Println("Bot started")
	log.Print("Bot started")
	b.Start()
}
