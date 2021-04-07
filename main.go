package main

import (
	"log"
	"os"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/guregu/dynamo"
	"github.com/robfig/cron/v3"
	tb "gopkg.in/tucnak/telebot.v2"
)

const (
	baseURL = "https://meteoinfo.by/radar"

	monisoringDistance = 34  // pixel
	homeX              = 206 // pixel
	homeY              = 231 // pixel

	RGBDeviation = 5 // for each R, G and B level

	nowImageName  = "now.png"
	prevImageName = "prev.png"

	cronSchedule = "@every 10m"

	tableName = "ukbb-bot"
	AWSRegion = "us-east-1"
)

var (
	botToken string

	precipLow  = Pixel{155, 234, 143}
	precipMed  = Pixel{88, 255, 67}
	precipHigh = Pixel{70, 194, 120}

	convPrecipLow  = Pixel{70, 147, 248}
	convPrecipMed  = Pixel{12, 89, 255}
	convPrecipHigh = Pixel{97, 83, 192}

	storm70  = Pixel{255, 146, 163}
	storm90  = Pixel{255, 63, 53}
	storm100 = Pixel{194, 6, 17}

	hailLow  = Pixel{255, 234, 12}
	hailMed  = Pixel{255, 152, 17}
	hailHigh = Pixel{168, 76, 6}

	squallLow  = Pixel{221, 168, 254}
	squallMed  = Pixel{232, 90, 255}
	squallHigh = Pixel{190, 28, 255}
)

// dynamodbTable returns DynamoDB table object
func dynamodbTable() dynamo.Table {
	session, err := session.NewSession()
	if err != nil {
		log.Println("Can't open a new session", err)

	}
	db := dynamo.New(session, &aws.Config{Region: aws.String(AWSRegion)})
	return db.Table(tableName)
}

// ##### INIT #####
func init() {
	// Get environment variables and check errors
	botToken = os.Getenv("UKBB_BOT_TOKEN")
	if len(botToken) == 0 {
		log.Fatal("UKBB_BOT_TOKEN environment variable is not set. Exit.")
	}

	// Download first image bewfore bot and cron start
	downloadImage(imageURL())
}

// ##### MAIN #####
func main() {
	log.Println("Starting Bot...")

	// Create new bot entity
	b, err := tb.NewBot(tb.Settings{
		Token:  botToken,
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

	// Add periodic job for downloading new images and sending alerts
	c := cron.New()
	c.AddFunc(cronSchedule, func() {
		// Save previous image to file and download a new image
		copyNewToPrev()
		downloadImage(imageURL())

		// Check weather
		gettingWorse := isItGettingWorse()

		// If weather is bad
		if gettingWorse {
			log.Println("Weather is getting worse. Sending alerts.")

			// Open databse
			table := dynamodbTable()

			// Get all users
			var users []tb.User
			err = table.Scan().All(&users)
			if err != nil {
				log.Println("Can't get users from databse", err)
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

	// Start Bot and Cron
	c.Start()
	b.Start()

}
