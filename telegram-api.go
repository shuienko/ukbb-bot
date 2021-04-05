package main

import tb "gopkg.in/tucnak/telebot.v2"

const (
	startReplyMessage = `Привіт! Я буду надсилати повідомлення про наближення несприятливих погодніх умов.

Дощ, град, зливи тощо. Моніторинг відбувається з інтервалом в 10 хвилин в радіусі 30 км від с. Петропавлівська Борщагівка Київської області.

Нічого робити не потрібно. Просто чекай на звістки про погану погоду 😉`
	getRaratImageMsg = `Зображення з Радару 📡`
)

func handleStart(m *tb.Message) string {
	// save m.Sender record to DB. key: ID, value []byte from json.Marshal(m.Sender)
	return startReplyMessage
}

func handleGetRadarImage(m *tb.Message) *tb.Photo {
	return &tb.Photo{File: tb.FromDisk(NowImageName)}
}
