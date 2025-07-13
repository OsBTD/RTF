package handlers

import (
	"net/http"

	"echohub/models"
)

type WebApp struct {
	Users         *models.UserModel
	Categories    *models.CategoryModel
	Posts         *models.PostModel
	Comments      *models.CommentModel
	Conversations *models.ConversationModel
	Messages      *models.MessageModel
	Sessions      *models.SessionModel
	Hub WSHub
	Rl  *RateLimiter
}

func (app *WebApp) NewRouter() http.Handler {
	mux := http.NewServeMux()

	// Serve ONLY the "assets" directory under /assets/
	public := http.StripPrefix("/public/", http.FileServer(http.Dir("../frontend/public")))
	mux.Handle("/public/", public)

	// Serve the HTML file on root
	mux.HandleFunc("/", app.Home)
	mux.HandleFunc("POST /signup", app.SignUp)
	mux.HandleFunc("POST /signin", app.SignIn)
	mux.HandleFunc("DELETE /signout", app.SignOut)
	mux.HandleFunc("POST /me", app.LoggedUser)
	// mux.HandleFunc("/auth", app.Auth)


	mux.HandleFunc("POST /categories", app.GetCategories)
	mux.HandleFunc("POST /newpost", app.NewPost)
	mux.HandleFunc("POST /posts", app.GetPosts)
	mux.HandleFunc("POST /comments", app.GetPostComments)
	mux.HandleFunc("POST /newcomment", app.NewComment) // TODO to implement
	mux.HandleFunc("/ws", app.HTTPtoWS)
	mux.HandleFunc("POST /recent", app.Recent)
	mux.HandleFunc("POST /conversation", app.Conversation)
	mux.HandleFunc("POST /mark-seen", app.MarkSeen)


	// return app.Rl.RLMiddleware((mux))
	return app.Rl.RLMiddleware(app.AuthMiddleware(mux))
}
