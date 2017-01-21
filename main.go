package main

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/cookiejar"
	"os"
	"time"

	"github.com/patrickmn/go-cache"
	"gopkg.in/cheggaaa/pb.v1"
	"gopkg.in/yaml.v2"
)

var (
	version, build, buildDate string
)

var applicationExitFunction = func(code int) { os.Exit(code) }

func abstractExitFunction(exit int) {
	applicationExitFunction(exit)
}

func init() {
	log.Printf("Version:    [%s]\nBuild:      [%s]\nBuild Date: [%s]\n", version, build, buildDate)
	flag.StringVar(&configPath, "config", "./config.yml", "Path to config.yml")
	flag.Parse()
}

func initConfigs() {
	Cache = cache.New(cache.NoExpiration, cache.NoExpiration)

	data, err := ioutil.ReadFile(configPath)
	checkErrAndExit(err)

	err = yaml.Unmarshal(data, &Configs)
	checkErrAndExit(err)

	cookieJar, err := cookiejar.New(nil)
	checkErrAndExit(err)
	httpClient = &http.Client{
		Jar: cookieJar,
	}
}

func checkErrAndExit(err error) {
	if err != nil {
		log.Println(err)
		time.Sleep(10 * time.Millisecond)
		abstractExitFunction(1)
	}
}

func checkFolderExistsOrCreate() {
	_, err := os.Stat(Configs.DownloadPath)
	if os.IsNotExist(err) {
		err = os.MkdirAll(Configs.DownloadPath, 755)
		checkErrAndExit(err)
	}
}

func main() {
	initConfigs()
	log.Println(printObject(Configs))
	checkFolderExistsOrCreate()

	initDBConnection()
	defer sqlDB.Close()

	initDB()
	updateCacheDataFromDB()

	processAuth()
	log.Println("Starting playlist getter")
	videoList, err := getVideoListFromPlayList()
	checkErrAndExit(err)
	log.Println("Processing playlist finished")
	log.Println(videoList)

	// create bar
	pbar := pb.New(len(videoList))
	pbar.SetRefreshRate(time.Second)
	pbar.ShowPercent = true
	pbar.ShowBar = true
	pbar.ShowCounters = true
	pbar.ShowTimeLeft = false
	pbar.ShowSpeed = false
	pbar.SetWidth(80)
	pbar.SetMaxWidth(180)
	pbar.Start()

	for i := len(videoList) - 1; i >= 0; i-- {
		processVideoDownloadByURL(videoList[i])
		pbar.Increment()
	}

	pbar.FinishPrint("Download finished!")
}

func printObject(v interface{}) string {
	res2B, _ := json.Marshal(v)
	return string(res2B)
}
