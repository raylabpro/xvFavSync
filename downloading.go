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

func processVideoDownloadByURL(url string) {
	downloadNodeInfoURL, videoID, _ := getVideoURLwithNodeInfo(url)
	if isVideoExist(videoID) == false {
		urlFinal, _ := getVideoDirectDLURL(downloadNodeInfoURL)
		err := downloadFromURL(urlFinal, videoID)
		if err == nil {
			addToReadyList(videoID)
		}
	}
}

func getVideoURLwithNodeInfo(url string) (string, string, error) {
	r := regexp.MustCompile(`video([0-9]+)`)
	parseRes := r.FindAllStringSubmatch(url, -1)
	if len(parseRes) <= 0 {
		return "", "", errors.New("Video ID Not found. Wrong URL")
	}
	if len(parseRes[0]) <= 0 {
		return "", "", errors.New("Video ID Not found. Wrong URL")
	}
	videoID := parseRes[0][1]
	if videoID == "" {
		return "", "", errors.New("Cant get Video ID from url")
	}
	resultURL := fmt.Sprintf("http://www.xvideos.com/video-download/%s/", videoID)
	return resultURL, videoID, nil
}

func getVideoDirectDLURL(url string) (string, error) {

	response, err := httpClient.Get(url)
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

func downloadFromURL(url string, fileName string) error {
	filePath := Configs.DownloadPath + fileName + ".mp4"
	//log.Println("Downloading", "to", filePath)

	response, err := httpClient.Get(url)
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
	links := []string{}

	//log.Println("Processing main playlist page", PlayListURL)
	videos, err := processLinksFromPage(PlayListURL)
	if err != nil {
		log.Fatalln(err)
		return links, err
	}
	for _, x := range videos {
		links = append(links, x)
	}
	//log.Println("Processed main playlist page", PlayListURL, "already links in db", len(links))

	counter := 0
	for {
		counter = counter + 1
		//log.Println("Processing additional", counter, "playlist page", PlayListURL + "/" + strconv.Itoa(counter))
		videos, err := processLinksFromPage(PlayListURL + "/" + strconv.Itoa(counter))
		if err != nil {
			break
		}
		for _, x := range videos {
			links = append(links, x)
		}
		//log.Println("Processed additional", counter, "playlist page", PlayListURL + "/" + strconv.Itoa(counter), "already links in db", len(links))
	}
	return links, nil
}

func processLinksFromPage(url string) ([]string, error) {
	links := []string{}
	response, err := httpClient.Get(url)
	if err != nil {
		return links, errors.New("Error while getting " + url + " - " + err.Error())
	}
	if response.StatusCode != 200 {
		return links, errors.New("Error while getting " + url + " code not 200")
	}
	doc, err := goquery.NewDocumentFromResponse(response)
	if err != nil {
		log.Println(err)
		return links, err
	}

	videoTable := doc.Find(".thumb-inside")
	if videoTable.Length() <= 0 {
		return links, errors.New("No elements on page")
	}
	videoTable.Each(func(i int, s *goquery.Selection) {
		elementURL := s.Text()
		r := regexp.MustCompile(`\/video([0-9]+)\/`)
		parseRes := r.FindAllStringSubmatch(elementURL, -1)
		links = append(links, parseRes[0][0])
	})
	return links, nil
}
