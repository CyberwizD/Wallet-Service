package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

// GoogleUser represents the subset of Google profile info we care about.
type GoogleUser struct {
	Email string `json:"email"`
	Name  string `json:"name"`
}

// NewGoogleOAuth builds the oauth2 config for Google sign-in.
func NewGoogleOAuth(clientID, clientSecret, redirectURL string) *oauth2.Config {
	return &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  redirectURL,
		Scopes:       []string{"openid", "email", "profile"},
		Endpoint:     google.Endpoint,
	}
}

// ExchangeCode exchanges an auth code for a token.
func ExchangeCode(ctx context.Context, cfg *oauth2.Config, code string) (*oauth2.Token, error) {
	return cfg.Exchange(ctx, code)
}

// FetchGoogleUser pulls the profile for the given token.
func FetchGoogleUser(ctx context.Context, token *oauth2.Token) (*GoogleUser, error) {
	client := oauth2.NewClient(ctx, oauth2.StaticTokenSource(token))
	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("google userinfo returned status %d", resp.StatusCode)
	}
	var gu GoogleUser
	if err := json.NewDecoder(resp.Body).Decode(&gu); err != nil {
		return nil, err
	}
	return &gu, nil
}
