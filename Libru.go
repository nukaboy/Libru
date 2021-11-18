package main

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"time"

	badger "github.com/dgraph-io/badger/v3"
)

//Settings struct
type Settings struct {
	Folders   []Folder `json:"folders"`
	CheckTime int      `json:"checkTime"`
}

//Settings for folder to scan
type Folder struct {
	Path      string `json:"path"`
	Recursive bool   `json:"recursive"`
}

//Database file root
type Database struct {
	Entries []Entry `json:"entries"`
}

//Database entry
type Entry struct {
	Path string `json:"path"`
	Hash string `json:"hash"`
	Text string `json:"text"`
}

func checkFile(path string) {
	match, _ := regexp.MatchString("\\.txt$", path)
	if match {
		f, err := os.Open(path)
		if err != nil {
			log.Fatal(err)
		}
		defer f.Close()

		h := sha256.New()
		if _, err := io.Copy(h, f); err != nil {
			log.Fatal(err)
		}

	}
}

func checkFolder(folder Folder) {
	err := filepath.Walk(folder.Path,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			go checkFile(path)
			return nil
		})
	if err != nil {
		log.Println(err)
	}
}

func main() {
	// Open settings file and read contents
	jsonFile, err := os.Open("settings.json")
	if err != nil {
		fmt.Println(err)
	}
	defer jsonFile.Close()
	byteValue, _ := ioutil.ReadAll(jsonFile)
	var settings Settings
	json.Unmarshal(byteValue, &settings)

	//Open database
	dir, _ := os.Getwd()
	db, err := badger.Open(badger.DefaultOptions(dir + "/badger"))
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	for {
		for _, f := range settings.Folders {
			go checkFolder(f)
		}
		time.Sleep(time.Duration(settings.CheckTime) * time.Second)
	}
}
