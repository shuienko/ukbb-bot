package main

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/prologic/bitcask"
	tb "gopkg.in/tucnak/telebot.v2"
)

const (
	startReplyMessage = `Привіт! Я буду надсилати повідомлення про наближення несприятливих погодніх умов.

Дощ, град, зливи тощо. Моніторинг відбувається з інтервалом в 10 хвилин в радіусі 30 км від с. Петропавлівська Борщагівка Київської області.

Нічого робити не потрібно. Просто чекай на звістки про погану погоду 😉`
	alertMessage     = `❗️ В радіусі 30 км погода погіршується. Пильнуй.`
	getRaratImageMsg = `Зображення з Радару 📡`
)

// handleStart save m.Sender record to DB. key: ID, value []byte from json.Marshal(m.Sender)
func handleStart(m *tb.Message) string {
	db, _ := bitcask.Open(DBPath)
	defer db.Close()

	userID := fmt.Sprintf("%d", m.Sender.ID)
	userObj, err := json.Marshal(m.Sender)
	if err != nil {
		log.Fatal("Cant't convert m.Sender to JSON")
	}

	if !db.Has([]byte(userID)) {
		err = db.Put([]byte(userID), userObj)
		if err == nil {
			log.Printf("User [%s] saved to database\n", userID)
		} else {
			log.Printf("Can't save user [%s] to database\n", userID)
		}
	} else {
		log.Printf("User [%s] is already in database\n", userID)
	}

	return startReplyMessage
}

// handleGetRadarImage upload now.png from disk to chat
func handleGetRadarImage(m *tb.Message) *tb.Photo {
	return &tb.Photo{File: tb.FromDisk(NowImageName)}
}
