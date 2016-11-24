package main

import (
	"io"
	"net/url"
	"regexp"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"encoding/json"
	"errors"
	"github.com/PuerkitoBio/goquery"
	"strconv"
)

func processAuth() {
	formData := url.Values{
		"login": {Configs.Login},
		"password": {Configs.Password},
		"referer": {"http://www.xvideos.com/"},
		"log": {""},
	}
	resp, err := httpClient.PostForm(authURL, formData)
	defer resp.Body.Close()
	checkErrAndExit(err)

}

func processVideoDownloadByUrl(url string) {
	downloadNodeInfoUrl, videoId, _ := getVideoURLwithNodeInfo(url)
	if isVideoExist(videoId) == false {
		urlFinal, _ := getVideoDirectDLURL(downloadNodeInfoUrl)
		err := downloadFromUrl(urlFinal, videoId)
		if err == nil {
			addToReadyList(videoId)
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
	videoId := parseRes[0][1]
	if videoId == "" {
		return "", "", errors.New("Cant get Video ID from url")
	}
	resultURL := fmt.Sprintf("http://www.xvideos.com/video-download/%s/", videoId)
	return resultURL, videoId, nil
}

func getVideoDirectDLURL(url string) (string, error) {

	response, err := httpClient.Get(url)
	if err != nil {
		log.Println("Error while getting direct video url", url, "-", err)
		return "", err
	}
	defer response.Body.Close()

	var jsonData DLInfo
	b, _ := ioutil.ReadAll(response.Body)
	err = json.Unmarshal(b, &jsonData)
	if err != nil {
		return "", err
	}
	if jsonData.Logged != true {
		processAuth()
		return getVideoDirectDLURL(url)
	}
	return jsonData.Url, nil
}

func downloadFromUrl(url string, fileName string) error {
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
	for ; ; {
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
		elementUrl  := s.Text()
		r := regexp.MustCompile(`\/video([0-9]+)\/`)
		parseRes := r.FindAllStringSubmatch(elementUrl, -1)
		links = append(links, parseRes[0][0])
	})
	return links, nil
}