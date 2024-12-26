package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/joho/godotenv"
)

func changeWorkingDirectory(t *testing.T, dir string) {
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("Failed to change working directory to %s: %v", dir, err)
	}
}

func getProjectRoot(t *testing.T) string {
	projectRoot, err := filepath.Abs("..")
	if err != nil {
		t.Fatalf("Failed to get project root: %v", err)
	}
	return projectRoot
}

func TestConnectDatabase(t *testing.T) {
	// Change the working directory to the project root
	projectRoot := getProjectRoot(t)
	changeWorkingDirectory(t, projectRoot)

	// Load the .env file
	if err := godotenv.Load(".env"); err != nil {
		t.Fatalf("Failed to load .env file from %s: %v", ".env", err)
	}

	// Connect to the database
	ConnectDatabase()

	if DB == nil {
		t.Fatal("Database connection is nil")
	}
}
