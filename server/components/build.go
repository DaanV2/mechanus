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
	"github.com/DaanV2/mechanus/server/infrastructure/tracing"
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
	tracingConf := tracing.GetConfig()

	// Setup OpenTelemetry tracing
	tracerProvider, err := tracing.SetupTracing(setupCtx, tracingConf)
	if err != nil {
		return nil, err
	}
	tracingManager := tracing.NewManager(tracerProvider)

	// Storage
	v, dbErr := GetDatabaseOptions()
	if dbErr != nil {
		return nil, dbErr
	}

	db, dbErr := SetupDatabase(setupCtx, v...)
	if dbErr != nil {
		return nil, dbErr
	}

	userRepo := repositories.NewUserRepository(db)
	jtiService := authentication.NewJTIService(db)
	storageProvider := storage.DBStorage[*authentication.KeyData](db)

	// Service and Managers
	keyManager, keyErr := authentication.NewKeyManager(storageProvider)
	if keyErr != nil {
		return nil, keyErr
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
	mdnsServer, mdnsErr := CreateMDNSServer(setupCtx, mdnsConf)
	if mdnsErr != nil {
		return nil, mdnsErr
	}

	serverManager := &servers.Manager{}
	serverManager.Register(
		CreateAPIServer(apiConf, websocketHandler, rpcs, tracingConf),
		CreateWebServer(webConf, lfManager, lfManager, tracingConf),
		mdnsServer,
	)
	server := &Server{
		Manager:    serverManager,
		Users:      userService,
		DB:         db,
		Components: lfManager,
	}
	lfManager.Add(screenManager, keyManager, db, tracingManager)

	return server, nil
}
