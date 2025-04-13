package main

import (
	"log"

	"github.com/joho/godotenv"
	_ "github.com/mattn/go-sqlite3"
	"github.com/moo-d/AwaraBot/internal/bot"
	"go.mau.fi/whatsmeow/store/sqlstore"
	waLog "go.mau.fi/whatsmeow/util/log"
)

func main() {
	waLog.Stdout("NETWORK", "DEBUG", true)
	waLog.Stdout("DATABASE", "INFO", true)

	err := godotenv.Load("../.env")
	if err != nil {
		log.Fatalln("cannot load .env files")
	}

	container, err := sqlstore.New("sqlite3", "file:bot.db?_foreign_keys=on&_journal_mode=WAL&_timeout=5000", nil)
	if err != nil {
		log.Fatalf("DB error: %v", err)
	}

	device, err := container.GetFirstDevice()
	if err != nil {
		log.Fatalf("Device error: %v", err)
	}

	botInstance := bot.NewBot(device, waLog.Stdout("BOT", "INFO", true))
	log.Println("Starting bot...")
	botInstance.Run()
}
