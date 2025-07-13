package handlers

import (
	"net/http"

	"echohub/models"
)

func (App *WebApp) SignUp(w http.ResponseWriter, r *http.Request) {
	var NewUser models.User
	if decodeErr := decodeJson(r, &NewUser); decodeErr != nil {
		http.Error(w, "Invalid JSON format.", http.StatusBadRequest)
		return
	}

	if err := App.Users.ValidateUser(&NewUser, "newUser"); err != nil {
		encodeJson(w, http.StatusBadRequest, nil)
		return
	}

	if err := App.Users.InsertUser(NewUser); err != nil {
		encodeJson(w, http.StatusInternalServerError, nil)
		return
	}

	if err := encodeJson(w, http.StatusCreated, nil); err != nil {
		http.Error(w, "failed to encode object.", http.StatusInternalServerError)
		return
	}
}

func (App *WebApp) SignIn(w http.ResponseWriter, r *http.Request) {
	var User *models.User
	if decodeErr := decodeJson(r, &User); decodeErr != nil {
		http.Error(w, "Invalid JSON format", http.StatusBadRequest)
		return
	}
	if err := App.Users.ValidateUser(User, "User"); err != nil {
		encodeJson(w, http.StatusUnauthorized, nil)
		return
	}
	session, err := App.Sessions.GenerateNewSession(User.ID)
	if err != nil {
		encodeJson(w, http.StatusInternalServerError, err.Error())
		return
	}
	cokkie, err := App.Sessions.UpSertSession(session)
	if err != nil {
		encodeJson(w, http.StatusInternalServerError, err.Error())
		return
	}
	// TODO get user by id and encode it as response +token
	// and change getting that token logic on client side
	User, err = App.Users.GetUserByID(User.ID)
	if err != nil {
		encodeJson(w, http.StatusInternalServerError, err.Error())
		return
	}
	User.Token = session.Token
	http.SetCookie(w, &cokkie)

	if err := encodeJson(w, http.StatusOK, &User); err != nil {
		http.Error(w, "failed to encode object.", http.StatusInternalServerError)
		return
	}
}

func (App *WebApp) SignOut(w http.ResponseWriter, r *http.Request) {
	user, ok := r.Context().Value(contextKeyUser).(*models.User)
	if !ok {
		encodeJson(w, http.StatusUnauthorized, nil)
		return
	}

	if err := App.Sessions.DeleteSession(user.ID); err != nil {
		encodeJson(w, http.StatusInternalServerError, nil)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "session_id",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		MaxAge:   -1,
	})

	encodeJson(w, http.StatusOK, nil)
}

func (App *WebApp) LoggedUser(w http.ResponseWriter, r *http.Request) {
	user, ok := r.Context().Value(contextKeyUser).(*models.User)
	if !ok {
		encodeJson(w, http.StatusUnauthorized, nil)
		return
	}

	encodeJson(w, http.StatusOK, user)
}
