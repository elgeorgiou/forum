package handlers

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"forum/internal/auth"
	"forum/internal/database"
	"forum/internal/session"
)

const (
	googleAuthorizeURL = "https://accounts.google.com/o/oauth2/v2/auth"
	googleTokenURL     = "https://oauth2.googleapis.com/token"
	googleUserInfoURL  = "https://openidconnect.googleapis.com/v1/userinfo"

	googleProvider    = "google"
	googleStateCookie = "google_oauth_state"
)

type GoogleOAuthHandler struct {
	db       *sql.DB
	sessions *session.Manager
	client   *http.Client
}

type googleTokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	Scope       string `json:"scope"`
	IDToken     string `json:"id_token"`
	Error       string `json:"error"`
}

type googleUser struct {
	Sub           string `json:"sub"`
	Email         string `json:"email"`
	EmailVerified bool   `json:"email_verified"`
	Name          string `json:"name"`
	GivenName     string `json:"given_name"`
	FamilyName    string `json:"family_name"`
}

func NewGoogleOAuthHandler(
	db *sql.DB,
	sessions *session.Manager,
) *GoogleOAuthHandler {
	return &GoogleOAuthHandler{
		db:       db,
		sessions: sessions,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (h *GoogleOAuthHandler) Login(
	w http.ResponseWriter,
	r *http.Request,
) {
	if r.Method != http.MethodGet {
		RenderErrorPage(
			w,
			http.StatusMethodNotAllowed,
		)
		return
	}

	clientID := os.Getenv("GOOGLE_CLIENT_ID")
	clientSecret := os.Getenv("GOOGLE_CLIENT_SECRET")

	if clientID == "" || clientSecret == "" {
		RenderErrorPage(
			w,
			http.StatusInternalServerError,
		)
		return
	}

	state, err := generateGoogleOAuthState()
	if err != nil {
		RenderErrorPage(
			w,
			http.StatusInternalServerError,
		)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     googleStateCookie,
		Value:    state,
		Path:     "/",
		MaxAge:   600,
		HttpOnly: true,
		Secure:   oauthSecureCookies(),
		SameSite: http.SameSiteLaxMode,
	})

	values := url.Values{}
	values.Set("client_id", clientID)
	values.Set("redirect_uri", googleRedirectURL(r))
	values.Set("response_type", "code")
	values.Set("scope", "openid email profile")
	values.Set("state", state)

	http.Redirect(
		w,
		r,
		googleAuthorizeURL+"?"+values.Encode(),
		http.StatusSeeOther,
	)
}

func (h *GoogleOAuthHandler) Callback(
	w http.ResponseWriter,
	r *http.Request,
) {
	if r.Method != http.MethodGet {
		RenderErrorPage(
			w,
			http.StatusMethodNotAllowed,
		)
		return
	}

	if r.URL.Query().Get("error") != "" {
		http.Redirect(
			w,
			r,
			"/auth?mode=login",
			http.StatusSeeOther,
		)
		return
	}

	if !validGoogleState(r) {
		RenderErrorPage(
			w,
			http.StatusBadRequest,
		)
		return
	}

	expireGoogleStateCookie(w)

	code := r.URL.Query().Get("code")
	if code == "" {
		RenderErrorPage(
			w,
			http.StatusBadRequest,
		)
		return
	}

	accessToken, err := h.exchangeGoogleCode(
		r,
		code,
	)
	if err != nil {
		RenderErrorPage(
			w,
			http.StatusInternalServerError,
		)
		return
	}

	googleUserData, err := h.getGoogleUser(
		accessToken,
	)
	if err != nil {
		RenderErrorPage(
			w,
			http.StatusInternalServerError,
		)
		return
	}

	if googleUserData.Sub == "" ||
		googleUserData.Email == "" ||
		!googleUserData.EmailVerified {
		RenderErrorPage(
			w,
			http.StatusBadRequest,
		)
		return
	}

	userID, err := h.findOrCreateGoogleUser(
		googleUserData,
	)
	if err != nil {
		RenderErrorPage(
			w,
			http.StatusInternalServerError,
		)
		return
	}

	if err := h.sessions.Create(w, userID); err != nil {
		RenderErrorPage(
			w,
			http.StatusInternalServerError,
		)
		return
	}

	http.Redirect(
		w,
		r,
		"/",
		http.StatusSeeOther,
	)
}

func (h *GoogleOAuthHandler) exchangeGoogleCode(
	r *http.Request,
	code string,
) (string, error) {
	values := url.Values{}

	values.Set(
		"client_id",
		os.Getenv("GOOGLE_CLIENT_ID"),
	)

	values.Set(
		"client_secret",
		os.Getenv("GOOGLE_CLIENT_SECRET"),
	)

	values.Set("code", code)
	values.Set("grant_type", "authorization_code")
	values.Set("redirect_uri", googleRedirectURL(r))

	request, err := http.NewRequest(
		http.MethodPost,
		googleTokenURL,
		strings.NewReader(values.Encode()),
	)
	if err != nil {
		return "", fmt.Errorf(
			"create google token request: %w",
			err,
		)
	}

	request.Header.Set(
		"Content-Type",
		"application/x-www-form-urlencoded",
	)

	response, err := h.client.Do(request)
	if err != nil {
		return "", fmt.Errorf(
			"request google token: %w",
			err,
		)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return "", fmt.Errorf(
			"google token status: %s",
			response.Status,
		)
	}

	var token googleTokenResponse

	if err := json.NewDecoder(
		response.Body,
	).Decode(&token); err != nil {
		return "", fmt.Errorf(
			"decode google token: %w",
			err,
		)
	}

	if token.Error != "" {
		return "", fmt.Errorf(
			"google token error: %s",
			token.Error,
		)
	}

	if token.AccessToken == "" {
		return "", errors.New(
			"google returned empty access token",
		)
	}

	return token.AccessToken, nil
}

func (h *GoogleOAuthHandler) getGoogleUser(
	accessToken string,
) (googleUser, error) {
	request, err := http.NewRequest(
		http.MethodGet,
		googleUserInfoURL,
		nil,
	)
	if err != nil {
		return googleUser{}, fmt.Errorf(
			"create google user request: %w",
			err,
		)
	}

	request.Header.Set(
		"Authorization",
		"Bearer "+accessToken,
	)

	response, err := h.client.Do(request)
	if err != nil {
		return googleUser{}, fmt.Errorf(
			"request google user: %w",
			err,
		)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return googleUser{}, fmt.Errorf(
			"google user status: %s",
			response.Status,
		)
	}

	var user googleUser

	if err := json.NewDecoder(
		response.Body,
	).Decode(&user); err != nil {
		return googleUser{}, fmt.Errorf(
			"decode google user: %w",
			err,
		)
	}

	return user, nil
}

func (h *GoogleOAuthHandler) findOrCreateGoogleUser(
	googleUserData googleUser,
) (int64, error) {
	account, err := database.GetOAuthAccount(
		h.db,
		googleProvider,
		googleUserData.Sub,
	)
	if err == nil {
		return account.UserID, nil
	}

	if !errors.Is(
		err,
		database.ErrOAuthAccountNotFound,
	) {
		return 0, err
	}

	email := strings.TrimSpace(
		googleUserData.Email,
	)

	existingUser, err := database.GetUserByEmail(
		h.db,
		email,
	)
	if err == nil {
		_, err = database.CreateOAuthAccount(
			h.db,
			existingUser.ID,
			googleProvider,
			googleUserData.Sub,
		)
		if err != nil {
			return 0, err
		}

		return existingUser.ID, nil
	}

	if !errors.Is(
		err,
		database.ErrUserNotFound,
	) {
		return 0, err
	}

	baseUsername := googleUsername(
		googleUserData,
	)

	username, err := h.availableGoogleUsername(
		baseUsername,
	)
	if err != nil {
		return 0, err
	}

	randomPassword, err := generateGoogleOAuthState()
	if err != nil {
		return 0, err
	}

	passwordHash, err := auth.HashPassword(
		randomPassword,
	)
	if err != nil {
		return 0, err
	}

	userID, err := database.CreateUser(
		h.db,
		username,
		email,
		passwordHash,
	)
	if err != nil {
		return 0, err
	}

	_, err = database.CreateOAuthAccount(
		h.db,
		userID,
		googleProvider,
		googleUserData.Sub,
	)
	if err != nil {
		return 0, err
	}

	return userID, nil
}

func (h *GoogleOAuthHandler) availableGoogleUsername(
	base string,
) (string, error) {
	base = strings.TrimSpace(base)

	if base == "" {
		base = "google-user"
	}

	username := base

	for suffix := 0; ; suffix++ {
		if suffix > 0 {
			username = fmt.Sprintf(
				"%s-%d",
				base,
				suffix,
			)
		}

		_, err := database.GetUserByUsername(
			h.db,
			username,
		)

		if errors.Is(
			err,
			database.ErrUserNotFound,
		) {
			return username, nil
		}

		if err != nil {
			return "", err
		}
	}
}

func googleUsername(
	user googleUser,
) string {
	name := strings.TrimSpace(
		user.GivenName,
	)

	if name == "" {
		name = strings.TrimSpace(
			user.Name,
		)
	}

	if name == "" {
		emailParts := strings.SplitN(
			user.Email,
			"@",
			2,
		)

		name = emailParts[0]
	}

	name = strings.ToLower(name)
	name = strings.ReplaceAll(name, " ", "-")

	if name == "" {
		return "google-user"
	}

	return name
}

func generateGoogleOAuthState() (string, error) {
	randomBytes := make([]byte, 32)

	if _, err := rand.Read(randomBytes); err != nil {
		return "", fmt.Errorf(
			"generate google oauth state: %w",
			err,
		)
	}

	return hex.EncodeToString(randomBytes), nil
}

func validGoogleState(
	r *http.Request,
) bool {
	state := r.URL.Query().Get("state")

	if state == "" {
		return false
	}

	cookie, err := r.Cookie(
		googleStateCookie,
	)
	if err != nil {
		return false
	}

	return cookie.Value != "" &&
		cookie.Value == state
}

func expireGoogleStateCookie(
	w http.ResponseWriter,
) {
	http.SetCookie(w, &http.Cookie{
		Name:     googleStateCookie,
		Value:    "",
		Path:     "/",
		Expires:  time.Unix(0, 0),
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   oauthSecureCookies(),
		SameSite: http.SameSiteLaxMode,
	})
}

func googleRedirectURL(
	r *http.Request,
) string {
	if configured := strings.TrimSpace(
		os.Getenv("GOOGLE_REDIRECT_URL"),
	); configured != "" {
		return configured
	}

	scheme := "http"

	if r.TLS != nil {
		scheme = "https"
	}

	return scheme +
		"://" +
		r.Host +
		"/auth/google/callback"
}
