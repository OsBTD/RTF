package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	_ "github.com/mattn/go-sqlite3"

	"echohub/handlers"
	"echohub/models"
)

func main() {
	db, err := sql.Open("sqlite3", "./db/app.db")
	if err != nil {
		log.Fatalln(err)
	}

	webForum := handlers.WebApp{
		Users: &models.UserModel{
			DB: db,
		},
		Sessions: &models.SessionModel{
			DB: db,
		},
		Categories: &models.CategoryModel{
			DB: db,
		},
		Posts: &models.PostModel{
			DB: db,
		},
		Comments: &models.CommentModel{
			DB: db,
		},
		Conversations: &models.ConversationModel{
			DB: db,
		},
		Messages: &models.MessageModel{
			DB: db,
		},
		Hub: handlers.WSHub{
			Upgrader: websocket.Upgrader{
				CheckOrigin: func(r *http.Request) bool { return true },
			},
			Clients:   make(map[*websocket.Conn]*models.User),
			Broadcast: make(chan models.Message),
			Lock:      sync.Mutex{},
		},
		Rl: handlers.NewRateLimiter(20, time.Second),
	}

	port := ":" + os.Getenv("PORT")
	if port == ":" {
		port += "8080"
	}
	server := http.Server{
		Addr:    port,
		Handler: webForum.NewRouter(),
	}

	go webForum.BroadcastMessages()

	log.Println("server listening on http://localhost" + port)

	if err := server.ListenAndServe(); err != nil {
		log.Fatalln(err)
	}
}
