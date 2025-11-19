package checks

import (
	"context"

	"github.com/DaanV2/mechanus/server/components"
	"github.com/DaanV2/mechanus/server/infrastructure/config"
	"github.com/DaanV2/mechanus/server/infrastructure/persistence/models"
	"github.com/DaanV2/mechanus/server/pkg/extensions/xgorm"
	"github.com/DaanV2/mechanus/server/pkg/extensions/xstrings"
	"github.com/charmbracelet/log"
	"github.com/google/uuid"
	"github.com/lib/pq"
)

var (
	// InitializeConfig is the configuration object for server initialization settings.
	InitializeConfig = config.New("initialize")
	// AdminUser is the configuration key for the admin username.
	AdminUser = InitializeConfig.String("initialize.admin.username", "", "The admin username to use when initializing")
	// AdminPassword is the configuration key for the admin password.
	AdminPassword = InitializeConfig.String("initialize.admin.password", "", "The admin password to use when initializing")
)

// InitializeServer runs the necessary checks to verify if the server has been properly initialized.
func InitializeServer(ctx context.Context, server *components.ServerComponents) {
	ctx = xgorm.WithPrefix(ctx, "checks")

	// Is there an admin account?
	users, err := server.Users.Find(ctx, &models.User{Roles: pq.StringArray{"admin"}})
	if err != nil {
		log.Fatal("error attempting to check if there is an admin account", "error", err)
	}
	if len(users) == 0 {
		password := xstrings.FirstNotEmpty(AdminPassword.Value(), uuid.NewString())
		admin := models.User{
			Username:     xstrings.FirstNotEmpty(AdminUser.Value(), "admin"),
			PasswordHash: []byte(password),
			Roles:        pq.StringArray{"admin"},
		}
		if err := server.Users.Create(ctx, &admin); err != nil {
			log.Fatal("couldn't create admin account", "error", err)
		}

		log.Warnf("!!!! Create admin account, will only output this once !!!!\nusername: %s\npassword: %s", admin.Username, password)
	}
}
