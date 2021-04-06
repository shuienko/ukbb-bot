package main

import (
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/guregu/dynamo"
	tb "gopkg.in/tucnak/telebot.v2"
)

const (
	startReplyMessage = `Привіт! Я буду надсилати повідомлення про наближення несприятливих погодніх умов.

Дощ, град, зливи тощо. Моніторинг відбувається з інтервалом в 15 хвилин в радіусі 30 км від с. Петропавлівська Борщагівка Київської області.

Нічого робити не потрібно. Просто чекай на звістки про погану погоду 😉`
	alertMessage     = `❗️ В радіусі 30 км погода погіршується. Пильнуй.`
	getRaratImageMsg = `Зображення з Радару 📡`
)

// handleStart saves m.Sender object to dynamoDB table
func handleStart(m *tb.Message) string {
	db := dynamo.New(session.New(), &aws.Config{Region: aws.String(AWSRegion)})
	table := db.Table(DynamoTable)

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
	return &tb.Photo{File: tb.FromDisk(NowImageName)}
}
