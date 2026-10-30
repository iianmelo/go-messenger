package controllers_test

import (
	"log"
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	if os.Getenv("RUN_FIREBASE_TESTS") != "1" {
		log.Print("skipping tests/integration; set RUN_FIREBASE_TESTS=1 to call Firebase")
		os.Exit(0)
	}
	os.Exit(m.Run())
}
