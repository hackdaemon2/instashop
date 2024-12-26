package config

import (
	"fmt"
	"log"
	"os"

	_ "github.com/go-sql-driver/mysql"
	"github.com/hackdaemon2/instashop/model"
	"github.com/jinzhu/gorm"
	"github.com/joho/godotenv"
	"golang.org/x/crypto/bcrypt"
)

var DB *gorm.DB

func LoadEnv() {
	if err := godotenv.Load(".env"); err != nil {
		log.Fatal("Error loading .env file")
	}
}

func GetEnv(key string) string {
	value := os.Getenv(key)
	if value == "" {
		log.Printf("Environment variable %s not set", key)
	}
	return value
}

// ConnectDatabase initializes the database and creates an admin user
// which will be the default user used to test the implementation
func ConnectDatabase() {
	var err error

	dbURL := fmt.Sprintf(
		"%s:%s@tcp(%s:%s)/%s?charset=utf8&parseTime=True&loc=Local", // MySQL connection format
		GetEnv("DB_USER"),
		GetEnv("DB_PASSWORD"),
		GetEnv("DB_HOST"),
		GetEnv("DB_PORT"),
		GetEnv("DB_NAME"),
	)

	// Establish database connection
	DB, err = gorm.Open("mysql", dbURL)
	if err != nil {
		log.Fatal("Error connecting to the database: ", err)
	}

	log.Println("Database connection established.")

	// Auto migrate the DB models (create tables and update schema if needed)
	if err := DB.AutoMigrate(&model.User{}, &model.Order{}, &model.Product{}).Error; err != nil {
		log.Fatalf("Error auto-migrating DB models: %v", err)
	}

	createNewAdminUser(err)
}

// Helper function to create a new admin user
func createNewAdminUser(err error) { // Check if the user "John Doe" exists
	var user model.User
	err = DB.Where("email = ?", GetEnv("ADMIN_EMAIL")).First(&user).Error

	if err != nil && !gorm.IsRecordNotFoundError(err) {
		// Some other error occurred while querying the database
		log.Fatalf("Error fetching user from the database: %v", err)
	} else if gorm.IsRecordNotFoundError(err) { // If user doesn't exist, create the user
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(GetEnv("ADMIN_PASSWORD")), bcrypt.DefaultCost)
		if err != nil {
			log.Fatalf("Error hashing user password: %v", err) // Return error if password hashing fails
		}

		// Create "John Doe" with admin role
		user = createNewUserModel(string(hashedPassword))

		// Create the user in the database
		if err := DB.Create(&user).Error; err != nil {
			log.Fatalf("Error creating user: %v", err)
		}

		log.Printf("User '%s %s' created with admin role.\n", GetEnv("ADMIN_FIRST_NAME"), GetEnv("ADMIN_LAST_NAME"))
	} else {
		log.Printf("User '%s %s' already exists.\n", GetEnv("ADMIN_FIRST_NAME"), GetEnv("ADMIN_LAST_NAME"))
	}
}

func createNewUserModel(hashedPassword string) model.User {
	return model.User{
		FirstName: GetEnv("ADMIN_FIRST_NAME"),
		LastName:  GetEnv("ADMIN_LAST_NAME"),
		Email:     GetEnv("ADMIN_EMAIL"),
		Password:  hashedPassword,
		Currency:  "NGN",
		Role:      model.AdminRole, // Assigning admin role
	}
}
