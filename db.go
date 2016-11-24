package main

import (
	"github.com/patrickmn/go-cache"
	"github.com/boltdb/bolt"
	"time"
)

const bucketName = "videos"

func addToCache(videoID string) {
	Cache.Add(videoID, 1, cache.NoExpiration)
}

func initDBConnection() {
	db, err := bolt.Open("videos.db", 0644, &bolt.Options{Timeout: 1 * time.Second})
	checkErrAndExit(err)
	sqlDB = db
}

func initDB()  {
	sqlDB.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(bucketName))
		checkErrAndExit(err)
		return nil
	})
}

func addToDB(videoID string) {
	err := sqlDB.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucketName))
		return b.Put([]byte(videoID), []byte("1"))
	})
	checkErrAndExit(err)
}

func isVideoExist(videoID string) bool {
	_, cExist := Cache.Get(videoID)
	if cExist {
		return true
	}
	return isVideoExistInDB(videoID)
}

func isVideoExistInDB(videoID string) bool {
	isExists := false
	sqlDB.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucketName))
		v := b.Get([]byte(videoID))
		if v == []byte("1") {
			isExists = true
		}
		return nil
	})
	return isExists
}


func addToReadyList(videoID string) {
	addToCache(videoID)
	addToDB(videoID)
}

func updateCacheDataFromDB() {
	sqlDB.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucketName))
		b.ForEach(func(k, v []byte) error {
			addToCache(string(k))
			return nil
		})
		return nil
	})
}