package main

import (
	"crypto/tls"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
)

// imageURL returns URL to the most recent radar image available
func imageURL() string {
	// Disable HTTPS certificate check. WORKAROUND
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	client := &http.Client{Transport: tr}

	// Create request
	req, _ := http.NewRequest("GET", baseURL+"?q=UKBB", nil)

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

	// Return empty string if no image reference found
	if imgURLPath == "" {
		return ""
	}

	return baseURL + imgURLPath
}

// downloadImage downloads image to NowImageName
func downloadImage(url string) {
	// Disable HTTPS certificate check. WORKAROUND
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	log.Println("Downloading image:", url)

	// Create client
	client := &http.Client{Transport: tr}

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
	err = ioutil.WriteFile(nowImageName, respBody, 0644)
	if err != nil {
		log.Println("Can't save image from URL", url)
	}
}

// copyNewToPrev will copy file with NowImageName to PrevImageName
func copyNewToPrev() {
	input, err := ioutil.ReadFile(nowImageName)

	if err != nil {
		log.Println("Can't read from file", nowImageName)
	}

	err = ioutil.WriteFile(prevImageName, input, 0644)
	if err != nil {
		log.Println("Can't write to file", prevImageName)
	}
}
