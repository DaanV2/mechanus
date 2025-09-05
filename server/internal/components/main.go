package components

import (
	"context"

	"github.com/DaanV2/mechanus/server/internal/grpc"
	"github.com/DaanV2/mechanus/server/internal/web"
	"github.com/DaanV2/mechanus/server/mechanus/scenes"
	"github.com/DaanV2/mechanus/server/pkg/application"
	"github.com/DaanV2/mechanus/server/pkg/authenication"
	grpc_handlers "github.com/DaanV2/mechanus/server/pkg/grpc/handlers"
	"github.com/DaanV2/mechanus/server/pkg/grpc/rpcs/rpcs_screens"
	"github.com/DaanV2/mechanus/server/pkg/grpc/rpcs/rpcs_users"
	user_service "github.com/DaanV2/mechanus/server/pkg/services/users"
	"github.com/DaanV2/mechanus/server/pkg/storage"
)

// BuildServer setup all the nessecary components for an api server
func BuildServer(ctx context.Context) (*Server, error) {
	v, err := GetDatabaseOptions()
	if err != nil {
		return nil, err
	}
	db, err := SetupDatabase(v...)
	if err != nil {
		return nil, err
	}
	service := user_service.NewService(db)
	jtiService := authenication.NewJTIService(db)
	componentManager := application.NewComponentManager()
	storageProvider := storage.DBStorage[*authenication.KeyData](db)
	keyManager, err := NewKeyManager(componentManager, storageProvider)
	if err != nil {
		return nil, err
	}
	jwtService := authenication.NewJWTService(jtiService, keyManager)
	loginService := rpcs_users.NewLoginService(service, jwtService)
	userService := rpcs_users.NewUserService(service)
	manager := scenes.NewManager()
	screenService := rpcs_screens.NewScreenService(manager)
	corsConfig := grpc_handlers.GetCORSConfig()
	corsHandler := grpc_handlers.NewCORSHandler(corsConfig)
	rpcs := grpc.RPCS{
		Login:  loginService,
		User:   userService,
		Screen: screenService,
		JWT:    jwtService,
		CORS:   corsHandler,
	}
	webServices := web.WEBServices{
		Components: componentManager,
	}
	serversManager, err := createServerManager(ctx, rpcs, webServices)
	if err != nil {
		return nil, err
	}
	server := &Server{
		Manager:    serversManager,
		Users:      service,
		DB:         db,
		Components: componentManager,
	}
	return server, nil
}
