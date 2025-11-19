package components

import (
	"context"

	"connectrpc.com/connect"
	"github.com/DaanV2/mechanus/server/application"
	"github.com/DaanV2/mechanus/server/engine/screens"
	"github.com/DaanV2/mechanus/server/infrastructure/authentication"
	"github.com/DaanV2/mechanus/server/infrastructure/lifecycle"
	"github.com/DaanV2/mechanus/server/infrastructure/persistence/repositories"
	"github.com/DaanV2/mechanus/server/infrastructure/servers"
	"github.com/DaanV2/mechanus/server/infrastructure/storage"
	"github.com/DaanV2/mechanus/server/infrastructure/tracing"
	"github.com/DaanV2/mechanus/server/infrastructure/transport/cors"
	"github.com/DaanV2/mechanus/server/infrastructure/transport/grpc"
	"github.com/DaanV2/mechanus/server/infrastructure/transport/websocket"
)

// BuildServer setup all the necessary components for an api server
func BuildServer(setupCtx context.Context) (*ServerComponents, error) {
	lfManager := lifecycle.NewManager()

	// Configs
	corsCfg := cors.GetCORSConfig()
	serverCfg := servers.GetServerConfig()
	websocketCfg := websocket.GetWebsocketConfig()
	tracingCfg := tracing.GetConfig()

	// Setup OpenTelemetry tracing
	tracerProvider, err := tracing.SetupTracing(setupCtx, tracingCfg)
	if err != nil {
		return nil, err
	}

	// Setup OpenTelemetry logging
	logProvider, err := tracing.SetupLogging(setupCtx, tracingCfg)
	if err != nil {
		return nil, err
	}

	tracingManager := tracing.NewManager(tracerProvider, logProvider)

	// Connect charm logger with OTEL log exporter
	// This must be done after OTEL log provider is set up
	if tracingCfg.Enabled {
		tracing.SetupLoggerBridge()
	}

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
	screenManager := screens.NewScreenManager()

	// Servers
	websocketHandler := websocket.NewWebsocketHandler(screenManager, jwtService, websocketCfg)

	router, err := CreateRouter(RouterSetup{
		WebsocketHandler: websocketHandler,
		HealthChecker:    lfManager,
		ReadyChecker:     lfManager,
		Interceptors: []connect.Interceptor{
			tracing.TraceGRPCMiddleware(tracingCfg),
			&grpc.LoggingInterceptor{},
			grpc.NewAuthenticator(jwtService),
		},
	}, RouterRPCS{
		Login: grpc.NewLoginServiceHandler(userService, jwtService),
		User:  grpc.NewUserServiceHandler(userService),
	},
		RouterConfig{
			CORS:    corsCfg,
			Server:  serverCfg,
			Tracing: tracingCfg,
		})
	if err != nil {
		return nil, err
	}

	server := &ServerComponents{
		Server:     CreateServer(router, serverCfg.Config),
		Users:      userService,
		DB:         db,
		Components: lfManager,
	}
	lfManager.Add(screenManager, keyManager, db, tracingManager, server.Server)

	return server, nil
}
