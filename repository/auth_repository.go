package repository

import (
	"errors"
	"fmt"
	"log"

	"github.com/hackdaemon2/instashop/model"
	"github.com/hackdaemon2/instashop/util"

	"github.com/jinzhu/gorm"
	"golang.org/x/crypto/bcrypt"
)

func FindUserBy(db *gorm.DB, id string, value string) (*model.User, error) {
	var existingUser model.User
	if err := db.Where(fmt.Sprintf("%s = ? AND is_deleted = false", id), value).First(&existingUser).Error; err != nil {
		return nil, err
	}
	return &existingUser, nil
}

func RegisterUser(db *gorm.DB, user *model.User) (*model.User, error) {
	if existingUser, err := FindUserBy(db, "email", user.Email); existingUser != nil || err != nil {
		if existingUser != nil {
			err = errors.New(fmt.Sprintf("user with email %s already exists", user.Email))
		}
		if !gorm.IsRecordNotFoundError(err) {
			return nil, err
		}
	}

	if err := db.Create(user).Error; err != nil {
		log.Println("unable to save user", err)
		return nil, err
	}

	log.Println("user saved successfully")
	return user, nil
}

func LoginUser(db *gorm.DB, email, password string) (*util.JwtData, error) {
	existingUser, err := FindUserBy(db, "email", email)
	if err != nil || existingUser == nil {
		if existingUser == nil {
			return nil, errors.New("invalid user credentials")
		}
		return nil, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(existingUser.Password), []byte(password)); err != nil {
		return nil, err
	}

	JwtData, err := util.GenerateJWT(existingUser.UserID, existingUser.Role)
	if err != nil {
		return nil, err
	}

	return &JwtData, nil
}
