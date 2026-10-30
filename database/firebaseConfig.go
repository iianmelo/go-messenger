package database

import (
	"context"
	"log"
	"os"

	firebase "firebase.google.com/go"
	"firebase.google.com/go/auth"

	common "go.messenger/Common"
	"go.messenger/config"
	"google.golang.org/api/option"
)

// OptionalFirebaseAuth starts the Admin SDK only in Firebase mode.
// Local mode returns nil so the process can boot without a service account.
func OptionalFirebaseAuth() *auth.Client {
	if config.LocalMode() {
		return nil
	}
	return InitFirebaseAuth()
}

func InitFirebaseAuth() *auth.Client {
	serviceAccount, fileExi := os.LookupEnv("SERVICE_ACCOUNT_JSON")

	if !fileExi {
		log.Fatalf("Please provide valid firebbase auth credential json!")
	}

	opt := option.WithCredentialsFile(serviceAccount)
	// Initialize the Firebase app
	app, err := firebase.NewApp(context.Background(), nil, opt)

	if err != nil {
		log.Fatalf("Error initializing Firebase app: %v", err)
	}

	authClient, err := app.Auth(context.Background())

	if err != nil {
		panic(err)
	}

	common.FireAuth = authClient

	return authClient
}
