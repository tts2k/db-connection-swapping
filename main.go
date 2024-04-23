package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/olahol/melody"
)

func writeFirstConfigFile() {
	config := Config{
		DBName: "test.db",
	}
	writeConfigFile(&config)
}

func main() {
	writeFirstConfigFile()
	err := InitDB()
	if err != nil {
		fmt.Println(err)
		return
	}
	defer CloseDB()

	m := melody.New()

	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		m.HandleRequest(w, r)
	})

	http.HandleFunc("/createdb", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		err := SwapDB()
		if err != nil {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			log.Fatalln("swap:", err)
		}
	})

	quit := make(chan bool)
	m.HandleConnect(func(s *melody.Session) {
		log.Println("connect")
		tickler := time.NewTicker(1 * time.Second)
		go func() {
			for {
				select {
				case <-tickler.C:
					title := GetName()
					s.Write([]byte(title))
				case <-quit:
					log.Println("quit")
					return
				}
			}
		}()
	})

	m.HandleDisconnect(func(_ *melody.Session) {
		log.Println("closed")
		close(quit)
	})

	fmt.Println("Serving on port 5000")
	http.ListenAndServe(":5000", nil)
}
