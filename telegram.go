package main

import (
	"log"

	tb "gopkg.in/tucnak/telebot.v2"
)

const (
	alertMessage     = `❗️ В радіусі 30 км зафіксовано опади. Пильнуй.`
	getRaratImageMsg = `Зображення з Радару 📡`

	startReplyMessage = `Привіт! Я буду надсилати повідомлення про наближення несприятливих погодніх умов.

Дощ, град, зливи тощо. Моніторинг відбувається з інтервалом в 10 хвилин в радіусі 30 км від с. Петропавлівська Борщагівка Київської області.

Кнопка внизу - щоб отримати актуальне зображення з радару в аеропорту Бориспіль.

Загалом, нічого робити не потрібно. Просто чекай на звістки про погану погоду 😉`
)

// handleStart saves m.Sender object to dynamoDB table
func handleStart(m *tb.Message) string {
	table := dynamodbTable()

	err := table.Put(m.Sender).Run()
	if err != nil {
		log.Println("Can't save user to database. ID:", m.Sender.ID, err)
	} else {
		log.Printf("User [%d] saved to database\n", m.Sender.ID)
	}

	return startReplyMessage
}

// handleGetRadarImage upload now.png from disk to chat
func handleGetRadarImage() *tb.Photo {
	return &tb.Photo{File: tb.FromDisk(nowImageName)}
}
