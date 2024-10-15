package services

import (
	"bajetapp/utils"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

type AuthService struct {
	db                *mongo.Database
	googleOauthConfig *oauth2.Config
}

type GoogleUserInfo struct {
	ID            string `json:"id"`
	Email         string `json:"email"`
	VerifiedEmail bool   `json:"verified_email"`
	Name          string `json:"name"`
	GivenName     string `json:"given_name"`
	FamilyName    string `json:"family_name"`
	Picture       string `json:"picture"`
	Locale        string `json:"locale"`
}

func NewAuthService(db *mongo.Database, googleClientID, googleClientSecret, redirectURL string) *AuthService {
	return &AuthService{
		db: db,
		googleOauthConfig: &oauth2.Config{
			ClientID:     googleClientID,
			ClientSecret: googleClientSecret,
			RedirectURL:  redirectURL,
			Scopes: []string{
				"https://www.googleapis.com/auth/userinfo.profile",
				"https://www.googleapis.com/auth/userinfo.email",
			},
			Endpoint: google.Endpoint,
		},
	}
}

func (as *AuthService) GoogleLogin(c echo.Context) (string, error) {
	// Implement login logic
	oauthStateString, err := utils.GenerateRandomString(32)
	if err != nil {
		return "", fmt.Errorf("failed to generate OAuth state string: %w", err)
	}

	// Store the OAuth state string in the session
	sess, _ := session.Get("session", c)
	sess.Values["oauthStateString"] = oauthStateString
	if err := sess.Save(c.Request(), c.Response()); err != nil {
		return "", fmt.Errorf("failed to save session: %w", err)
	}

	url := as.googleOauthConfig.AuthCodeURL(oauthStateString)
	return url, nil
}

func (as *AuthService) ProcessGoogleCallback(c echo.Context, stateString string, code string) (GoogleUserInfo, error) {
	// Retrieve the OAuth state string from the session
	sess, _ := session.Get("session", c)
	oauthStateString, ok := sess.Values["oauthStateString"].(string)
	if !ok || stateString != oauthStateString {
		return GoogleUserInfo{}, fmt.Errorf("invalid OAuth state string")
	}
	token, err := as.googleOauthConfig.Exchange(context.Background(), code)
	if err != nil {
		return GoogleUserInfo{}, fmt.Errorf("failed to exchange token: %w", err)
	}

	client := as.googleOauthConfig.Client(context.Background(), token)
	userInfoResp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		return GoogleUserInfo{}, fmt.Errorf("failed to get user info: %w", err)
	}
	defer userInfoResp.Body.Close()

	if userInfoResp.StatusCode != http.StatusOK {
		return GoogleUserInfo{}, fmt.Errorf("failed to get user info: status code %d", userInfoResp.StatusCode)
	}

	userInfoByte, err := io.ReadAll(userInfoResp.Body)
	if err != nil {
		return GoogleUserInfo{}, fmt.Errorf("failed to read user info: %w", err)
	}

	var userInfo GoogleUserInfo
	if err := json.Unmarshal(userInfoByte, &userInfo); err != nil {
		return GoogleUserInfo{}, fmt.Errorf("failed to unmarshal user info: %v", err)
	}

	// Store the user info in the session
	sess, _ = session.Get("session", c)
	sess.Values["user_info"] = string(userInfoByte)
	err = sess.Save(c.Request(), c.Response())
	if err != nil {
		return GoogleUserInfo{}, fmt.Errorf("failed to save session: %w", err)
	}

	// store the user info in the database if not exists, check by userinfo email. if already exist update the user info
	_, err = as.db.Collection("users").UpdateOne(
		c.Request().Context(),
		bson.M{"email": userInfo.Email},
		bson.M{"$set": userInfo},
		options.Update().SetUpsert(true),
	)
	if err != nil {
		return GoogleUserInfo{}, fmt.Errorf("failed to upsert user info: %w", err)
	}
	return userInfo, nil
}

func (as *AuthService) Logout() {

}
