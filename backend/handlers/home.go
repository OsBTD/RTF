package handlers

import (
	"net/http"
	"time"

	"echohub/models"
)

// serve html file on home page
func (app *WebApp) Home(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "../frontend/index.html")
}

func (app *WebApp) GetCategories(w http.ResponseWriter, r *http.Request) {
	var category models.Category
	decodeJson(r, &category)
	switch category.Target {
	case "all":
		categories, err := app.Categories.GetAllCategories()
		if err != nil {
			encodeJson(w, http.StatusInternalServerError, err.Error())
			return
		}
		encodeJson(w, http.StatusOK, categories)
		return
	default:
		category, err := app.Categories.GetCategoryByID(category.ID)
		if err != nil {
			encodeJson(w, http.StatusInternalServerError, err.Error())
			return
		}
		encodeJson(w, http.StatusOK, category)
		return
	}
}

func (app *WebApp) NewComment(w http.ResponseWriter, r *http.Request) {
	user, ok := r.Context().Value(contextKeyUser).(*models.User)
	if !ok {
		encodeJson(w, http.StatusUnauthorized, nil)
		return
	}

	var comment models.Comment

	if err := decodeJson(r, &comment); err != nil {
		encodeJson(w, http.StatusBadRequest, nil)
		return
	}

	if err := models.ValidateComment(&comment); err != nil {
		encodeJson(w, http.StatusBadRequest, nil)
	}

	comment.UserID = user.ID

	if err := app.Comments.InsertComment(comment); err != nil {
		encodeJson(w, http.StatusInternalServerError, nil)
		return
	}

	encodeJson(w, http.StatusOK, comment)
}

func (app *WebApp) NewPost(w http.ResponseWriter, r *http.Request) {
	user, ok := r.Context().Value(contextKeyUser).(*models.User)
	if !ok {
		encodeJson(w, http.StatusUnauthorized, nil)
		return
	}

	var post models.Post

	decodeJson(r, &post)
	post.UserID = user.ID

	if err := models.ValidatePost(&post); err != nil {
		encodeJson(w, http.StatusBadRequest, nil)
	}

	if err := app.Posts.InsertPost(post); err != nil {
		encodeJson(w, http.StatusInternalServerError, nil)
	}

	encodeJson(w, http.StatusOK, nil)
}

func (app *WebApp) GetPosts(w http.ResponseWriter, r *http.Request) {
	user, ok := r.Context().Value(contextKeyUser).(*models.User)
	if !ok {
		encodeJson(w, http.StatusUnauthorized, nil)
		return
	}
	var filter *models.PostFilter
	decodeJson(r, &filter)

	posts, err, errCode := app.Posts.FilterPosts(filter, user.ID)
	if err != nil || errCode == http.StatusNoContent {
		encodeJson(w, errCode, nil)
		return
	}

	encodeJson(w, http.StatusOK, posts)
}

func (app *WebApp) GetPostComments(w http.ResponseWriter, r *http.Request) {
	var filter *models.CommentsFilter
	decodeJson(r, &filter)

	comments, err := app.Comments.GetComments(filter)
	if err != nil {
		encodeJson(w, http.StatusInternalServerError, nil)
		return
	}
	encodeJson(w, http.StatusOK, comments)
}

func (app *WebApp) Recent(w http.ResponseWriter, r *http.Request) {
	user, ok := r.Context().Value(contextKeyUser).(*models.User)
	if !ok {
		encodeJson(w, http.StatusUnauthorized, nil)
		return
	}

	users, err := app.Users.GetSortedUsersByConversation(user.ID)
	if err != nil {
		encodeJson(w, http.StatusInternalServerError, err.Error())
		return
	}
	encodeJson(w, http.StatusOK, users)
}

func (app *WebApp) Conversation(w http.ResponseWriter, r *http.Request) {
	var filter *models.MessagesFilter
	decodeJson(r, &filter)

	user, ok := r.Context().Value(contextKeyUser).(*models.User)
	if !ok || !filter.ConversationID.Valid {
		encodeJson(w, http.StatusUnauthorized, nil)
		return
	}
	chunkedMessages, err := app.Messages.GetMessages(user.ID, filter)
	if err != nil {
		encodeJson(w, http.StatusInternalServerError, err.Error())
		return
	}
	encodeJson(w, http.StatusOK, chunkedMessages)
}

func (app *WebApp) MarkSeen(w http.ResponseWriter, r *http.Request) {

	var mark models.MessagesFilter
	if err := decodeJson(r, &mark); err != nil {
		encodeJson(w, http.StatusBadRequest, nil)
		return
	}

	now := time.Now()
	_, err := app.Messages.DB.Exec(`
		UPDATE messages SET seen_at = ? 
		WHERE conversation_id = ?`, now, mark.ConversationID)
	if err != nil {
		encodeJson(w, http.StatusInternalServerError, nil)
		return
	}
	encodeJson(w, http.StatusOK, nil)
}
