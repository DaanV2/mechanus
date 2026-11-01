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
	"github.com/DaanV2/mechanus/server/infrastructure/telemetry"
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
	tracingCfg := telemetry.GetConfig()
	authCfg := authentication.GetAuthConfig()

	// Setup OpenTelemetry telemetry
	otelmanager, err := telemetry.Setup(setupCtx, tracingCfg)
	if err != nil {
		return nil, err
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
	campaignRepo := repositories.NewCampaignRepository(db)
	jtiService := authentication.NewJTIService(db)
	storageProvider := storage.DBStorage[*authentication.KeyData](db)

	// Service and Managers
	keyManager, keyErr := authentication.NewKeyManager(storageProvider)
	if keyErr != nil {
		return nil, keyErr
	}
	jwtService := authentication.NewJWTService(jtiService, keyManager).WithOptions(authCfg.JWT)
	keyManager.WithRotation(authCfg.KeyRotation, jwtService.MaxTokenLifetime())
	userService := application.NewUserService(userRepo)
	campaignService := application.NewCampaignService(campaignRepo)
	screenManager := screens.NewScreenManager()

	// Servers
	websocketHandler := websocket.NewWebsocketHandler(screenManager, jwtService, websocketCfg)

	router, err := CreateRouter(RouterSetup{
		WebsocketHandler: websocketHandler,
		HealthChecker:    lfManager,
		ReadyChecker:     lfManager,
		Interceptors: []connect.Interceptor{
			telemetry.TraceGRPCMiddleware(tracingCfg),
			&grpc.LoggingInterceptor{},
			grpc.NewAuthenticator(jwtService),
		},
	}, RouterRPCS{
		Campaign: grpc.NewCampaignServiceHandler(campaignService, userService),
		Login:    grpc.NewLoginServiceHandler(userService, jwtService, jtiService),
		User:     grpc.NewUserServiceHandler(userService, jtiService, grpc.UserServiceOptions{SignupEnabled: authCfg.SignupEnabled}),
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
	lfManager.Add(screenManager, keyManager, jwtService, db, otelmanager, server.Server)

	return server, nil
}
