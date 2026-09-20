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
	"strconv"
	"strings"
	"time"

	"forum/internal/auth"
	"forum/internal/database"
	"forum/internal/session"
)

const (
	githubAuthorizeURL = "https://github.com/login/oauth/authorize"
	githubTokenURL     = "https://github.com/login/oauth/access_token"
	githubUserURL      = "https://api.github.com/user"
	githubEmailsURL    = "https://api.github.com/user/emails"

	githubProvider    = "github"
	githubStateCookie = "github_oauth_state"
)

type GitHubOAuthHandler struct {
	db       *sql.DB
	sessions *session.Manager
	client   *http.Client
}

type githubTokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	Scope       string `json:"scope"`
	Error       string `json:"error"`
}

type githubUser struct {
	ID    int64  `json:"id"`
	Login string `json:"login"`
	Email string `json:"email"`
}

type githubEmail struct {
	Email    string `json:"email"`
	Primary  bool   `json:"primary"`
	Verified bool   `json:"verified"`
}

func NewGitHubOAuthHandler(
	db *sql.DB,
	sessions *session.Manager,
) *GitHubOAuthHandler {
	return &GitHubOAuthHandler{
		db:       db,
		sessions: sessions,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (h *GitHubOAuthHandler) Login(
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

	clientID := os.Getenv("GITHUB_CLIENT_ID")
	clientSecret := os.Getenv("GITHUB_CLIENT_SECRET")

	if clientID == "" || clientSecret == "" {
		RenderErrorPage(
			w,
			http.StatusInternalServerError,
		)
		return
	}

	state, err := generateOAuthState()
	if err != nil {
		RenderErrorPage(
			w,
			http.StatusInternalServerError,
		)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     githubStateCookie,
		Value:    state,
		Path:     "/",
		MaxAge:   600,
		HttpOnly: true,
		Secure:   oauthSecureCookies(),
		SameSite: http.SameSiteLaxMode,
	})

	values := url.Values{}
	values.Set("client_id", clientID)
	values.Set("redirect_uri", githubRedirectURL(r))
	values.Set("scope", "user:email")
	values.Set("state", state)

	http.Redirect(
		w,
		r,
		githubAuthorizeURL+"?"+values.Encode(),
		http.StatusSeeOther,
	)
}

func (h *GitHubOAuthHandler) Callback(
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

	if !validGitHubState(r) {
		RenderErrorPage(
			w,
			http.StatusBadRequest,
		)
		return
	}

	expireGitHubStateCookie(w)

	code := r.URL.Query().Get("code")
	if code == "" {
		RenderErrorPage(
			w,
			http.StatusBadRequest,
		)
		return
	}

	accessToken, err := h.exchangeCode(
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

	githubUserData, err := h.getGitHubUser(
		accessToken,
	)
	if err != nil {
		RenderErrorPage(
			w,
			http.StatusInternalServerError,
		)
		return
	}

	email := strings.TrimSpace(githubUserData.Email)

	if email == "" {
		email, err = h.getGitHubPrimaryEmail(
			accessToken,
		)
		if err != nil {
			RenderErrorPage(
				w,
				http.StatusBadRequest,
			)
			return
		}
	}

	userID, err := h.findOrCreateUser(
		githubUserData,
		email,
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

func (h *GitHubOAuthHandler) exchangeCode(
	r *http.Request,
	code string,
) (string, error) {
	values := url.Values{}

	values.Set(
		"client_id",
		os.Getenv("GITHUB_CLIENT_ID"),
	)

	values.Set(
		"client_secret",
		os.Getenv("GITHUB_CLIENT_SECRET"),
	)

	values.Set("code", code)
	values.Set("redirect_uri", githubRedirectURL(r))

	request, err := http.NewRequest(
		http.MethodPost,
		githubTokenURL,
		strings.NewReader(values.Encode()),
	)
	if err != nil {
		return "", fmt.Errorf(
			"create github token request: %w",
			err,
		)
	}

	request.Header.Set(
		"Content-Type",
		"application/x-www-form-urlencoded",
	)

	request.Header.Set(
		"Accept",
		"application/json",
	)

	response, err := h.client.Do(request)
	if err != nil {
		return "", fmt.Errorf(
			"request github token: %w",
			err,
		)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return "", fmt.Errorf(
			"github token status: %s",
			response.Status,
		)
	}

	var token githubTokenResponse

	if err := json.NewDecoder(
		response.Body,
	).Decode(&token); err != nil {
		return "", fmt.Errorf(
			"decode github token: %w",
			err,
		)
	}

	if token.Error != "" {
		return "", fmt.Errorf(
			"github token error: %s",
			token.Error,
		)
	}

	if token.AccessToken == "" {
		return "", errors.New(
			"github returned empty access token",
		)
	}

	return token.AccessToken, nil
}

func (h *GitHubOAuthHandler) getGitHubUser(
	accessToken string,
) (githubUser, error) {
	request, err := githubAPIRequest(
		githubUserURL,
		accessToken,
	)
	if err != nil {
		return githubUser{}, err
	}

	response, err := h.client.Do(request)
	if err != nil {
		return githubUser{}, fmt.Errorf(
			"request github user: %w",
			err,
		)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return githubUser{}, fmt.Errorf(
			"github user status: %s",
			response.Status,
		)
	}

	var user githubUser

	if err := json.NewDecoder(
		response.Body,
	).Decode(&user); err != nil {
		return githubUser{}, fmt.Errorf(
			"decode github user: %w",
			err,
		)
	}

	if user.ID == 0 ||
		strings.TrimSpace(user.Login) == "" {
		return githubUser{}, errors.New(
			"github returned incomplete user data",
		)
	}

	return user, nil
}

func (h *GitHubOAuthHandler) getGitHubPrimaryEmail(
	accessToken string,
) (string, error) {
	request, err := githubAPIRequest(
		githubEmailsURL,
		accessToken,
	)
	if err != nil {
		return "", err
	}

	response, err := h.client.Do(request)
	if err != nil {
		return "", fmt.Errorf(
			"request github emails: %w",
			err,
		)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return "", fmt.Errorf(
			"github emails status: %s",
			response.Status,
		)
	}

	var emails []githubEmail

	if err := json.NewDecoder(
		response.Body,
	).Decode(&emails); err != nil {
		return "", fmt.Errorf(
			"decode github emails: %w",
			err,
		)
	}

	for _, email := range emails {
		if email.Primary &&
			email.Verified &&
			strings.TrimSpace(email.Email) != "" {
			return email.Email, nil
		}
	}

	for _, email := range emails {
		if email.Verified &&
			strings.TrimSpace(email.Email) != "" {
			return email.Email, nil
		}
	}

	return "", errors.New(
		"no verified github email",
	)
}

func (h *GitHubOAuthHandler) findOrCreateUser(
	githubUserData githubUser,
	email string,
) (int64, error) {
	providerUserID := strconv.FormatInt(
		githubUserData.ID,
		10,
	)

	account, err := database.GetOAuthAccount(
		h.db,
		githubProvider,
		providerUserID,
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

	existingUser, err := database.GetUserByEmail(
		h.db,
		email,
	)
	if err == nil {
		_, err = database.CreateOAuthAccount(
			h.db,
			existingUser.ID,
			githubProvider,
			providerUserID,
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

	username, err := h.availableUsername(
		githubUserData.Login,
	)
	if err != nil {
		return 0, err
	}

	randomPassword, err := generateOAuthState()
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
		githubProvider,
		providerUserID,
	)
	if err != nil {
		return 0, err
	}

	return userID, nil
}

func (h *GitHubOAuthHandler) availableUsername(
	base string,
) (string, error) {
	base = strings.TrimSpace(base)

	if base == "" {
		base = "github-user"
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

func githubAPIRequest(
	endpoint string,
	accessToken string,
) (*http.Request, error) {
	request, err := http.NewRequest(
		http.MethodGet,
		endpoint,
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"create github api request: %w",
			err,
		)
	}

	request.Header.Set(
		"Authorization",
		"Bearer "+accessToken,
	)

	request.Header.Set(
		"Accept",
		"application/vnd.github+json",
	)

	return request, nil
}

func generateOAuthState() (string, error) {
	randomBytes := make([]byte, 32)

	if _, err := rand.Read(randomBytes); err != nil {
		return "", fmt.Errorf(
			"generate oauth state: %w",
			err,
		)
	}

	return hex.EncodeToString(randomBytes), nil
}

func validGitHubState(
	r *http.Request,
) bool {
	state := r.URL.Query().Get("state")

	if state == "" {
		return false
	}

	cookie, err := r.Cookie(
		githubStateCookie,
	)
	if err != nil {
		return false
	}

	return cookie.Value != "" &&
		cookie.Value == state
}

func expireGitHubStateCookie(
	w http.ResponseWriter,
) {
	http.SetCookie(w, &http.Cookie{
		Name:     githubStateCookie,
		Value:    "",
		Path:     "/",
		Expires:  time.Unix(0, 0),
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   oauthSecureCookies(),
		SameSite: http.SameSiteLaxMode,
	})
}

func githubRedirectURL(
	r *http.Request,
) string {
	if configured := strings.TrimSpace(
		os.Getenv("GITHUB_REDIRECT_URL"),
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
		"/auth/github/callback"
}

func oauthSecureCookies() bool {
	return os.Getenv(
		"FORUM_SECURE_COOKIES",
	) == "true"
}
