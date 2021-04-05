package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"time"

	"github.com/prologic/bitcask"
	"github.com/robfig/cron/v3"
	tb "gopkg.in/tucnak/telebot.v2"
)

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

	AlertCronSchedule    = "@every 9m30s"
	DownloadCronSchedule = "@every 10m"

	BaseURL = "https://meteoinfo.by/radar"
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
		_, err = b.Send(m.Sender, handleGetRadarImage(m), options)
		if err != nil {
			log.Println("Failed to respond on btnGetRadarImage:", err)
		}
	})

	c := cron.New()

	// Add periodic job for downloading new images
	c.AddFunc(DownloadCronSchedule, func() {
		// Copy it to prev.png and then download a new one
		input, err := ioutil.ReadFile(NowImageName)
		if err != nil {
			log.Println("Can't read from file", NowImageName)
		}
		err = ioutil.WriteFile(PrevImageName, input, 0644)
		if err != nil {
			log.Println("Can't write to file", PrevImageName)
		}

		DownloadImage(GetImageURL())
	})

	// Add periodic job for alerting
	c.AddFunc(AlertCronSchedule, func() {
		var userObj *tb.User

		log.Println("Running alert job.")

		// Check weather
		gettingWorse := isItGettingWorse()

		if gettingWorse {
			// Open databse
			db, _ := bitcask.Open(DBPath)
			defer db.Close()

			// Get all keys
			keys := db.Keys()
			for key := range keys {
				userBytearray, err := db.Get(key)
				if err != nil {
					log.Println("Can't get user object from database. ID:", string(key))
				}

				err = json.Unmarshal(userBytearray, userObj)
				if err != nil {
					log.Println("Can't Unmarshal user from byte array. ID:", string(key))
				}

				// Send message to a user
				_, err = b.Send(userObj, alertMessage)
				if err != nil {
					log.Println("Can't send message to", string(key))
				}
			}
		}
	})

	// Download first image bewfore bot and cron start
	DownloadImage(GetImageURL())

	c.Start()
	b.Start()

}
