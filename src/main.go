package main

import (
	"bufio"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"
	"sync"

	"github.com/gin-gonic/gin"
)

var (
	DIARY_PATH = "diary"                                  // default diary dir
	MARK       = time.Date(1, 1, 1, 1, 1, 1, 1, time.UTC) // time benchmark
	wg 			 sync.WaitGroup
)

// diary file
type TTText struct {
	FilePath string
	Title    string
	Time     string
	Content  string
}

// go template render to html
var diarys []TTText = nil

func main() {
	// create router engine
	e := gin.Default()

	// load html template file
	e.LoadHTMLFiles("index.html")

	e.GET("/", Handle)

	// update page data
	go func() {
		// 300s update
		for {
			time.Sleep(15 * time.Second)
			Refresh()
		}
	}()

	e.Run(":80")
}

func Handle(ctx *gin.Context) {
	wg.Add(1)
	go func(ctx *gin.Context) {
		defer wg.Done()

		// render diary list to template
		if diarys != nil {
			ctx.HTML(http.StatusOK, "index.html", gin.H{
				"diarys": diarys,
			})
		} else {
			go Load()
		}
	}(ctx)

	wg.Wait()
}


// load all diary file
func Load() {
	wg.Add(1)
	defer wg.Done()

	// dir, err := os.ReadDir(DIARY_PATH)
	dir, err := os.ReadDir(os.Getenv("DIARY_PATH"))

	if err != nil {
		// os.Mkdir(DIARY_PATH, 0666)
		os.Mkdir(os.Getenv("DIARY_PATH"), 0666)
		log.Fatalln("DIARY_PATH system environment variable not defined")
	}

	for _, entry := range dir {
		// open diary file
		// p := filepath.Join(DIARY_PATH, entry.Name())
		p := filepath.Join(os.Getenv("DIARY_PATH"), entry.Name())

		// read a diary file
		go ReadDiary(p)
	}
}

func Refresh() {
	// dir, _ := os.ReadDir(DIARY_PATH)
	dir, _ := os.ReadDir(os.Getenv("DIARY_PATH"))

	var p string
	// var fn string
	for _, entry := range dir {
		p = filepath.Join(DIARY_PATH, entry.Name())
		// p := filepath.Join(os.Getenv("DIARY_PATH"), entry.Name())

		// get diary file create time
		s, _ := os.Stat(p)
		t := s.ModTime()

		// update new file time
		if MARK.Before(t) {
			MARK = t
		}
	}

	// append new diary to page
	if !IsExist(p) {
		ReadDiary(p)
		log.Println("append a latest diary")
	}
}

// is new file in biary list?
func IsExist(fn string) bool {
	for _, d := range diarys {
		if fn == d.FilePath {
			return true
		}
	}
	return false
}

// read diary
func ReadDiary(p string) {
	wg.Add(1)
	defer wg.Done()

	f, _ := os.Open(p)

	// read over then close
	defer f.Close()

	// create read stream
	s := bufio.NewScanner(f)
	var diary TTText

	// read diary file body
	for i := 0; s.Scan(); i++ {
		switch i {
		case 0:
			diary.Title = s.Text()
		case 1:
			diary.Time = s.Text()
		default:
			diary.Content += s.Text()
		}
	}

	// append to diary list
	diary.FilePath = f.Name()
	diarys = append(diarys, diary)
}
