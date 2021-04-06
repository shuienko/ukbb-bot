package main

import (
	"io/ioutil"
	"log"
	"os"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/guregu/dynamo"
	"github.com/robfig/cron/v3"
	tb "gopkg.in/tucnak/telebot.v2"
)

// TODO: database MUST be external. Options: Redis on DO/AWS Dynamodb/AWS Elastichache/Synology NFS/etc.

const (
	InfoDist     = 34 // pix
	WarningDist  = 23 // pix
	CriticalDist = 11 // pix

	HomeX        = 206
	HomeY        = 231
	RGBDeviation = 5 // for each R, G, B level

	NowImageName  = "now.png"
	PrevImageName = "prev.png"
	DBPath        = "ukbb-bot.db"

	CronSchedule = "@every 10m"

	BaseURL = "https://meteoinfo.by/radar"

	DynamoTable = "ukbb-bot"
	AWSRegion   = "us-east-1"
)

var (
	BotToken string

	PrecipLow  = Pixel{155, 234, 143}
	PrecipMed  = Pixel{88, 255, 67}
	PrecipHigh = Pixel{70, 194, 120}

	ConvPrecipLow  = Pixel{70, 147, 248}
	ConvPrecipMed  = Pixel{12, 89, 255}
	ConvPrecipHigh = Pixel{97, 83, 192}

	Storm70  = Pixel{255, 146, 163}
	Storm90  = Pixel{255, 63, 53}
	Storm100 = Pixel{194, 6, 17}

	HailLow  = Pixel{255, 234, 12}
	HailMed  = Pixel{255, 152, 17}
	HailHigh = Pixel{168, 76, 6}

	SquallLow  = Pixel{221, 168, 254}
	SquallMed  = Pixel{232, 90, 255}
	SquallHigh = Pixel{190, 28, 255}
)

// ##### INIT #####
func init() {
	// Get environment variables and check errors
	BotToken = os.Getenv("UKBB_BOT_TOKEN")
	if len(BotToken) == 0 {
		log.Fatal("UKBB_BOT_TOKEN environment variable is not set. Exit.")
	}
}

// ##### MAIN #####
func main() {
	log.Println("Starting Bot...")

	// Create new bot entity
	b, err := tb.NewBot(tb.Settings{
		Token:  BotToken,
		Poller: &tb.LongPoller{Timeout: 10 * time.Second},
	})
	if err != nil {
		log.Fatal("Can't create new bot object:", err)
		return
	}

	// Set Reply keyboard
	menu := &tb.ReplyMarkup{ResizeReplyKeyboard: true}
	btnGetRadarImage := menu.Text(getRaratImageMsg)
	menu.Reply(
		menu.Row(btnGetRadarImage),
	)

	// Add send options
	options := &tb.SendOptions{
		ParseMode:   "Markdown",
		ReplyMarkup: menu,
	}

	// Handle /start command
	b.Handle("/start", func(m *tb.Message) {
		_, err = b.Send(m.Sender, handleStart(m), options)
		if err != nil {
			log.Println("Failed to respond on /start command:", err)
		}
	})

	// Handle button
	b.Handle(&btnGetRadarImage, func(m *tb.Message) {
		_, err = b.Send(m.Sender, handleGetRadarImage(), options)
		if err != nil {
			log.Println("Failed to respond on btnGetRadarImage:", err)
		}
	})

	c := cron.New()

	// Add periodic job for downloading new images
	c.AddFunc(CronSchedule, func() {
		// Copy now.png to prev.png
		input, err := ioutil.ReadFile(NowImageName)
		if err != nil {
			log.Println("Can't read from file", NowImageName)
		}
		err = ioutil.WriteFile(PrevImageName, input, 0644)
		if err != nil {
			log.Println("Can't write to file", PrevImageName)
		}

		// Download now.png
		newImageURL := ImageURL()
		DownloadImage(newImageURL)

		log.Println("Starting weather check...")

		// Check weather
		gettingWorse := isItGettingWorse()

		// If weather conditions are bad
		if gettingWorse {
			log.Println("Weather is getting worse. Sending alerts.")

			// Open databse
			db := dynamo.New(session.New(), &aws.Config{Region: aws.String(AWSRegion)})
			table := db.Table(DynamoTable)

			// Get all keys
			var users []tb.User
			err = table.Scan().All(&users)
			if err != nil {
				log.Println("Can't get user list from databse", err)
			}

			// Send message to all users
			for _, user := range users {
				_, err = b.Send(&user, alertMessage, options)
				if err != nil {
					log.Printf("Can't send message to %d\n", user.ID)
				}
			}
		}
	})

	// Download first image bewfore bot and cron start
	DownloadImage(ImageURL())

	c.Start()
	b.Start()

}
