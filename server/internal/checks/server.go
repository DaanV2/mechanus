package checks

import (
	"context"

	"github.com/DaanV2/mechanus/server/internal/components"
	"github.com/DaanV2/mechanus/server/pkg/config"
	"github.com/DaanV2/mechanus/server/pkg/database/models"
	xgorm "github.com/DaanV2/mechanus/server/pkg/extensions/gorm"
	xstrings "github.com/DaanV2/mechanus/server/pkg/extensions/strings"
	"github.com/charmbracelet/log"
	"github.com/google/uuid"
	"github.com/lib/pq"
)

var (
	InitializeConfig = config.New("initialize")
	AdminUser        = InitializeConfig.String("initialize.admin.username", "", "The admin username to use when initializing")
	AdminPassword    = InitializeConfig.String("initialize.admin.password", "", "The admin password to use when initializing")
)

func InitializeServer(ctx context.Context, server *components.Server) {
	ctx = xgorm.WithPrefix(ctx, "checks")

	// Is there an admin account?
	users, err := server.Users.Find(ctx, &models.User{Roles: pq.StringArray{"admin"}})
	if err != nil {
		log.Fatal("error attempting to check if there is an admin account", "error", err)
	}
	if len(users) == 0 {
		password := xstrings.FirstNotEmpty(AdminPassword.Value(), uuid.NewString())
		admin := models.User{
			Name:         xstrings.FirstNotEmpty(AdminUser.Value(), "admin"),
			PasswordHash: []byte(password),
			Roles:        pq.StringArray{"admin"},
		}
		if err := server.Users.Create(ctx, &admin); err != nil {
			log.Fatal("couldn't create admin account", "error", err)
		}

		log.Warnf("!!!! Create admin account, will only output this once !!!!\nusername: %s\npassword: %s", admin.Name, password)
	}
}
