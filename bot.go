package main

import (
	"log"
	"os"
	"time"

	tb "gopkg.in/tucnak/telebot.v2"
)

const (
	InfoDist     = 34 // pix
	WarningDist  = 23 // pix
	CriticalDist = 11 // pix

	ScalePixKm   = 1.14 // pix/km
	HomeX        = 206
	HomeY        = 231
	RGBDeviation = 5 // for each R, G, B level

	NowImageName  = "now.png"
	PrevImageName = "prev.png"
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

/*
TODO:
- Parsre website http page and download png image.
- For each saved user(location) process nearby pixels (10-20 km range?). If something is there - alert.
- Probably, should split users and start processing within goroutines. Just to speed up.
- Need to have image cache. 10m min expiration. Just a local disk cache


URLs:
https://stackoverflow.com/questions/33186783/get-a-pixel-array-from-from-golang-image-image

*/

// ##### INIT #####
func init() {
	// Get environment variables and check errors
	BotToken = os.Getenv("KWABOT_BOT_TOKEN")
	if len(BotToken) == 0 {
		log.Fatal("KWABOT_BOT_TOKEN environment variable is not set. Exit.")
	}
}

func main() {
	log.Println("Start Bot")

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

	b.Start()

}
