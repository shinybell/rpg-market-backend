package auth

import (
	"context"
	"log"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/auth"
	"google.golang.org/api/option"
)

type Client struct {
	app        *firebase.App
	authClient *auth.Client
}

var globalClient *Client

// NewClient creates a new Firebase client
func NewClient(credentialsPath string) (*Client, error) {
	ctx := context.Background()

	opt := option.WithCredentialsFile(credentialsPath)
	app, err := firebase.NewApp(ctx, nil, opt)
	if err != nil {
		return nil, err
	}

	authClient, err := app.Auth(ctx)
	if err != nil {
		return nil, err
	}

	log.Println("Firebase Admin SDK initialized successfully")

	client := &Client{
		app:        app,
		authClient: authClient,
	}

	globalClient = client
	return client, nil
}

// GetAuthClient returns the Firebase Auth client
func (c *Client) GetAuthClient() *auth.Client {
	return c.authClient
}

// VerifyIDToken verifies a Firebase ID token
func (c *Client) VerifyIDToken(ctx context.Context, idToken string) (*auth.Token, error) {
	return c.authClient.VerifyIDToken(ctx, idToken)
}

// Global accessor (for convenience in middleware)
func GetClient() *Client {
	return globalClient
}
