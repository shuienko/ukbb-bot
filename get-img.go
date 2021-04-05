package main

import (
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
)

func GetImageURL() string {
	client := &http.Client{}

	// Create request
	req, _ := http.NewRequest("GET", BaseURL+"?q=UKBB", nil)

	parseFormErr := req.ParseForm()
	if parseFormErr != nil {
		log.Println(parseFormErr)
	}

	// Fetch Request
	resp, err := client.Do(req)

	if err != nil {
		log.Println("Failure : ", err)
	}

	// Read Response Body
	respBody, _ := ioutil.ReadAll(resp.Body)

	// Find image refereince in HTML body
	re := regexp.MustCompile(`/UKBB/UKBB_[0-9]+.png`)
	imgURLPath := re.FindString(string(respBody))

	return BaseURL + imgURLPath
}

func DownloadImage(url string) {
	log.Println("Downloading image:", url)

	// Create client
	client := &http.Client{}

	// Create request
	req, _ := http.NewRequest("GET", url, nil)

	// Fetch Request
	resp, err := client.Do(req)

	if err != nil {
		log.Println("Failure : ", err)
	}

	// Read Response Body
	respBody, _ := ioutil.ReadAll(resp.Body)

	// Save to file
	err = ioutil.WriteFile(NowImageName, respBody, 0644)
	if err != nil {
		log.Println("Can't save image from URL", url)
	}
}
