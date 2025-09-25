package components

import (
	"context"

	"github.com/DaanV2/mechanus/server/infrastructure/transport/websocket"
	"github.com/DaanV2/mechanus/server/internal/grpc"
	"github.com/DaanV2/mechanus/server/internal/web"
	"github.com/DaanV2/mechanus/server/mechanus/screens"
	"github.com/DaanV2/mechanus/server/pkg/application"
	"github.com/DaanV2/mechanus/server/pkg/authentication"
	grpc_handlers "github.com/DaanV2/mechanus/server/pkg/grpc/handlers"
	"github.com/DaanV2/mechanus/server/pkg/grpc/rpcs/rpcs_users"
	user_service "github.com/DaanV2/mechanus/server/pkg/services/users"
	"github.com/DaanV2/mechanus/server/pkg/storage"
)

// BuildServer setup all the necessary components for an api server
func BuildServer(setupCtx context.Context) (*Server, error) {
	v, err := GetDatabaseOptions()
	if err != nil {
		return nil, err
	}
	db, err := SetupDatabase(setupCtx, v...)
	if err != nil {
		return nil, err
	}
	service := user_service.NewService(db)
	jtiService := authentication.NewJTIService(db)
	componentManager := application.NewComponentManager()
	storageProvider := storage.DBStorage[*authentication.KeyData](db)
	keyManager, err := NewKeyManager(componentManager, storageProvider)
	if err != nil {
		return nil, err
	}
	jwtService := authentication.NewJWTService(jtiService, keyManager)
	loginService := rpcs_users.NewLoginService(service, jwtService)
	userService := rpcs_users.NewUserService(service)
	corsConfig := grpc_handlers.GetCORSConfig()
	corsHandler := grpc_handlers.NewCORSHandler(corsConfig)
	rpcs := grpc.RPCS{
		Login: loginService,
		User:  userService,
		JWT:   jwtService,
		CORS:  corsHandler,
	}
	webServices := web.WEBServices{
		Components: componentManager,
	}
	screenManager := screens.NewScreenManager()
	websocketConfig := websocket.GetWebsocketConfig()
	websocketService := websocket.NewWebsocketHandler(screenManager, jwtService, websocketConfig)

	serversManager, err := createServerManager(setupCtx, rpcs, websocketService, webServices)
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
