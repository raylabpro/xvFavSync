package main

import (
	"net/http"

	"github.com/boltdb/bolt"
	"github.com/patrickmn/go-cache"
)

//Cache ...
var Cache *cache.Cache

type conf struct {
	Login        string `yaml:"login"`
	Password     string `yaml:"password"`
	PlaylistLink string `yaml:"playlist_link"`
	DownloadPath string `yaml:"download_path"`
	UserAgent    string `yaml:"user_agent"`
}

//Configs ...
var Configs conf

var configPath string

var httpClient *http.Client

var sqlDB *bolt.DB

type dlInfo struct {
	Logged bool   `json:"LOGGED"`
	URL    string `json:"URL"`
	URLLow string `json:"URL_LOW"`
}

const authURL = "http://upload.xvideos.com/account/"
