package main

import (
	"context"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/gmail/v1"
	"google.golang.org/api/option"
)

type App struct {
	GmailService *gmail.Service
}

func NewApp(clientID, clientSecret, refreshToken string) (*App, error) {
	config := &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Endpoint:     google.Endpoint,
		Scopes:       []string{gmail.GmailReadonlyScope},
	}

	token := &oauth2.Token{
		RefreshToken: refreshToken,
	}

	client := config.Client(context.Background(), token)

	svc, err := gmail.NewService(context.Background(), option.WithHTTPClient(client))
	if err != nil {
		return nil, err
	}

	return &App{GmailService: svc}, nil
}
