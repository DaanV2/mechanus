package util_test

import (
	"time"

	"github.com/DaanV2/mechanus/server/pkg/database/models"
	xrand "github.com/DaanV2/mechanus/server/pkg/extensions/rand"
)

func CreateUserID() string {
	return xrand.MustID(36)
}

func CreateUser() *models.User {
	id := CreateUserID()

	return &models.User{
		Model: models.Model{
			ID:        id,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		Name:         "username_" + id,
		Roles:        []string{"user"},
		Campaigns:    []*models.Campaign{},
		Characters:   []*models.Character{},
		PasswordHash: []byte("password12345"),
	}
}
