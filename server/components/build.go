package components

import (
	"context"

	"github.com/DaanV2/mechanus/server/application"
	"github.com/DaanV2/mechanus/server/engine/screens"
	"github.com/DaanV2/mechanus/server/infrastructure/authentication"
	"github.com/DaanV2/mechanus/server/infrastructure/persistence/repositories"
	"github.com/DaanV2/mechanus/server/infrastructure/storage"
	"github.com/DaanV2/mechanus/server/infrastructure/transport/grpc"
	"github.com/DaanV2/mechanus/server/infrastructure/transport/http"
	"github.com/DaanV2/mechanus/server/infrastructure/transport/websocket"
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
	user_repo := repositories.NewUserRepository(db)
	service := application.NewUserService(user_repo)
	jtiService := authentication.NewJTIService(db)
	componentManager := application.NewComponentManager()
	storageProvider := storage.DBStorage[*authentication.KeyData](db)
	keyManager, err := NewKeyManager(componentManager, storageProvider)
	if err != nil {
		return nil, err
	}
	jwtService := authentication.NewJWTService(jtiService, keyManager)
	loginService := grpc.NewLoginService(service, jwtService)
	userService := grpc.NewUserService(service)
	corsConfig := grpc.GetCORSConfig()
	corsHandler := grpc.NewCORSHandler(corsConfig)
	rpcs := grpc.RPCS{
		Login: loginService,
		User:  userService,
		JWT:   jwtService,
		CORS:  corsHandler,
	}
	webServices := http.WEBServices{
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
