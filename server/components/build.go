package components

import (
	"context"

	"github.com/DaanV2/mechanus/server/application"
	"github.com/DaanV2/mechanus/server/engine/screens"
	"github.com/DaanV2/mechanus/server/infrastructure/authentication"
	"github.com/DaanV2/mechanus/server/infrastructure/lifecycle"
	"github.com/DaanV2/mechanus/server/infrastructure/persistence/repositories"
	"github.com/DaanV2/mechanus/server/infrastructure/servers"
	"github.com/DaanV2/mechanus/server/infrastructure/storage"
	"github.com/DaanV2/mechanus/server/infrastructure/transport/grpc"
	"github.com/DaanV2/mechanus/server/infrastructure/transport/http"
	"github.com/DaanV2/mechanus/server/infrastructure/transport/mdns"
	"github.com/DaanV2/mechanus/server/infrastructure/transport/websocket"
)

// BuildServer setup all the necessary components for an api server
func BuildServer(setupCtx context.Context) (*Server, error) {
	lfManager := lifecycle.NewManager()

	// Configs
	corsConf := grpc.GetCORSConfig()
	apiConf := grpc.GetAPIServerConfig()
	webConf := http.GetWebConfig()
	mdnsConf := mdns.GetServerConfig(webConf.Port)
	websocketConf := websocket.GetWebsocketConfig()

	// Storage
	v, err := GetDatabaseOptions()
	if err != nil {
		return nil, err
	}

	db, err := SetupDatabase(setupCtx, v...)
	if err != nil {
		return nil, err
	}

	userRepo := repositories.NewUserRepository(db)
	jtiService := authentication.NewJTIService(db)
	storageProvider := storage.DBStorage[*authentication.KeyData](db)

	// Service and Managers
	keyManager, err := authentication.NewKeyManager(storageProvider)
	if err != nil {
		return nil, err
	}
	jwtService := authentication.NewJWTService(jtiService, keyManager)
	userService := application.NewUserService(userRepo)
	rpcs := grpc.RPCS{
		Login: grpc.NewLoginServiceHandler(userService, jwtService),
		User:  grpc.NewUserServiceHandler(userService),
		JWT:   jwtService,
		CORS:  grpc.NewCORSHandler(corsConf),
	}
	screenManager := screens.NewScreenManager()

	// Servers
	websocketHandler := websocket.NewWebsocketHandler(screenManager, jwtService, websocketConf)
	mdnsServer, err := CreateMDNSServer(setupCtx, mdnsConf)
	if err != nil {
		return nil, err
	}

	serverManager := &servers.Manager{}
	serverManager.Register(
		CreateAPIServer(apiConf, websocketHandler, rpcs),
		CreateWebServer(webConf, lfManager, lfManager),
		mdnsServer,
	)
	server := &Server{
		Manager:    serverManager,
		Users:      userService,
		DB:         db,
		Components: lfManager,
	}
	lfManager.Add(screenManager, keyManager, db)

	return server, nil
}
