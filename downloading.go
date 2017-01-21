package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/url"
	"os"
	"regexp"
	"strconv"

	"net/http"

	"github.com/PuerkitoBio/goquery"
)

func processAuth() {
	formData := url.Values{
		"login":    {Configs.Login},
		"password": {Configs.Password},
		"referer":  {"http://www.xvideos.com/"},
		"log":      {""},
	}
	resp, err := httpClient.PostForm(authURL, formData)
	checkErrAndExit(err)
	defer resp.Body.Close()
}

func processVideoDownloadByURL(videoID string) {
	downloadNodeInfoURL := fmt.Sprintf("http://www.xvideos.com/video-download/%s/", videoID)
	if isVideoExist(videoID) == false {
		urlFinal, _ := getVideoDirectDLURL(downloadNodeInfoURL)
		err := downloadFromURL(urlFinal, videoID)
		if err == nil {
			addToReadyList(videoID)
		}
	}
}

func getVideoDirectDLURL(url string) (string, error) {
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("User-Agent", Configs.UserAgent)

	response, err := httpClient.Do(req)
	if err != nil {
		log.Println("Error while getting direct video url", url, "-", err)
		return "", err
	}
	defer response.Body.Close()

	var jsonData dlInfo
	b, _ := ioutil.ReadAll(response.Body)
	err = json.Unmarshal(b, &jsonData)
	if err != nil {
		return "", err
	}

	if jsonData.Logged != true {
		processAuth()
		return getVideoDirectDLURL(url)
	}
	return jsonData.URL, nil
}

func downloadFromURL(url string, videoID string) error {
	filePath := fmt.Sprintf("%v%v.mp4", Configs.DownloadPath, videoID)
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("User-Agent", Configs.UserAgent)
	response, err := httpClient.Do(req)
	if err != nil {
		log.Println("Error while downloading", url, "-", err)
		return err
	}
	defer response.Body.Close()

	output, err := os.Create(filePath)
	if err != nil {
		log.Println("Error while creating file", filePath, "-", err)
		return err
	}
	defer output.Close()

	_, err = io.Copy(output, response.Body)
	if err != nil {
		log.Println("Error while downloading", url, "-", err)
		return err
	}
	//log.Println(n, "bytes downloaded.")
	return nil
}

func getVideoListFromPlayList() ([]string, error) {
	PlayListURL := Configs.PlaylistLink
	videoIDs := []string{}

	//log.Println("Processing main playlist page", PlayListURL)
	videos, err := processLinksFromPage(PlayListURL)
	if err != nil {
		log.Fatalln(err)
		return videoIDs, err
	}
	for _, x := range videos {
		videoIDs = append(videoIDs, x)
	}

	counter := 0
	for {
		counter = counter + 1
		//log.Println("Processing additional", counter, "playlist page", PlayListURL + "/" + strconv.Itoa(counter))
		videos, err := processLinksFromPage(PlayListURL + "/" + strconv.Itoa(counter))
		if err != nil {
			break
		}
		for _, x := range videos {
			videoIDs = append(videoIDs, x)
		}
		//log.Println("Processed additional", counter, "playlist page", PlayListURL + "/" + strconv.Itoa(counter), "already links in db", len(links))
	}
	return videoIDs, nil
}

func processLinksFromPage(url string) ([]string, error) {
	videoIDs := []string{}

	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("User-Agent", Configs.UserAgent)

	response, err := httpClient.Do(req)
	if err != nil {
		return videoIDs, errors.New("Error while getting " + url + " - " + err.Error())
	}
	if response.StatusCode != 200 {
		return videoIDs, errors.New("Error while getting " + url + " code not 200 but " + string(response.StatusCode))
	}
	doc, err := goquery.NewDocumentFromResponse(response)
	if err != nil {
		log.Println(err)
		return videoIDs, err
	}

	videoTable := doc.Find("div.mozaique>div")
	if videoTable.Length() <= 0 {
		return videoIDs, errors.New("No elements on page")
	}
	r := regexp.MustCompile(`video_([0-9]+)`)

	videoTable.Each(func(i int, s *goquery.Selection) {
		vId, _ := s.Attr("id")
		parseRes := r.FindAllStringSubmatch(vId, -1)
		videoIDs = append(videoIDs, parseRes[0][1])
	})
	return videoIDs, nil
}
