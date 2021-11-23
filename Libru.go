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

	"github.com/disintegration/imaging"
	"github.com/ledongthuc/pdf"
	gosseract "github.com/otiai10/gosseract"
)

//Settings struct
type Settings struct {
	Folders   []Folder `json:"folders"`
	CheckTime int      `json:"checkTime"`
	DBDir     string   `json:"database"`
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

var settings Settings

func readPdfText(path string) (string, error) {
	f, r, err := pdf.Open(path)
	defer func() {
		_ = f.Close()
	}()
	if err != nil {
		return "", err
	}
	totalPage := r.NumPage()

	for pageIndex := 1; pageIndex <= totalPage; pageIndex++ {
		p := r.Page(pageIndex)
		if p.V.IsNull() {
			continue
		}

		rows, _ := p.GetTextByRow()
		for _, row := range rows {
			println(">>>> row: ", row.Position)
			for _, word := range row.Content {
				fmt.Println(word.S)
			}
		}
	}
	return "", nil
}

func preprocess(image string) {
	// Open a test image.
	src, err := imaging.Open(image, imaging.AutoOrientation(true))
	if err != nil {
		log.Fatalf("failed to open image: %v", err)
	}

	src = imaging.Resize(src, 4000, 0, imaging.Lanczos)
	src = imaging.Grayscale(src)
	src = imaging.AdjustContrast(src, 20)
	src = imaging.Sharpen(src, 2)

	err = imaging.Save(src, "tmpimg.png")
	if err != nil {
		log.Fatalf("failed to save image: %v", err)
	}
}

func checkFile(path string) {
	//Open database
	//	dir := settings.DBDir
	//	db, err := badger.Open(badger.DefaultOptions(dir))
	//	if err != nil {
	//		log.Fatal(err)
	//	}
	//	defer db.Close()
	matchImage, _ := regexp.MatchString("\\.(jpg|jpeg|png)$", path)
	matchPDF, _ := regexp.MatchString("\\.pdf$", path)
	if matchImage && false {
		f, err := os.Open(path)
		if err != nil {
			log.Fatal(err)
		}
		defer f.Close()

		h := sha256.New()
		if _, err := io.Copy(h, f); err != nil {
			log.Fatal(err)
		}

		client := gosseract.NewClient()
		defer client.Close()
		preprocess(path)
		client.SetImage("tmpimg.png")
		text, _ := client.Text()
		fmt.Println(text)

	} else if matchPDF {
		fmt.Println("----------------------")
		fmt.Println(readPdfText(path))
	}
}

func checkFolder(folder Folder) {
	err := filepath.Walk(folder.Path,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			checkFile(path)
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
	json.Unmarshal(byteValue, &settings)

	for {
		for _, f := range settings.Folders {
			checkFolder(f)
		}
		time.Sleep(time.Duration(settings.CheckTime) * time.Second)
	}
}
