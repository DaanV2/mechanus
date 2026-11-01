package grpc

import (
	"context"
	"errors"
	"strings"

	"connectrpc.com/connect"
	"github.com/DaanV2/mechanus/server/application"
	"github.com/DaanV2/mechanus/server/engine/authentication/roles"
	"github.com/DaanV2/mechanus/server/engine/scenes"
	"github.com/DaanV2/mechanus/server/infrastructure/authentication"
	"github.com/DaanV2/mechanus/server/infrastructure/logging"
	"github.com/DaanV2/mechanus/server/infrastructure/scenestore"
	scenesv1 "github.com/DaanV2/mechanus/server/proto/scenes/v1"
	"github.com/DaanV2/mechanus/server/proto/scenes/v1/scenesv1connect"
)

var _ scenesv1connect.SceneServiceHandler = &SceneServiceHandler{}

// SceneServiceHandler implements the scenes gRPC service, the GM's control over the active scene.
type SceneServiceHandler struct {
	scenes *application.SceneService
	logger logging.Enriched
}

// NewSceneServiceHandler creates a new SceneServiceHandler.
func NewSceneServiceHandler(sceneService *application.SceneService) *SceneServiceHandler {
	return &SceneServiceHandler{
		scenes: sceneService,
		logger: logging.Enriched{}.WithPrefix("grpc-scenes"),
	}
}

// List implements scenesv1connect.SceneServiceHandler.
func (s *SceneServiceHandler) List(ctx context.Context, _ *connect.Request[scenesv1.ListScenesRequest]) (*connect.Response[scenesv1.ListScenesResponse], error) {
	if _, err := requireRole(ctx, roles.User); err != nil {
		return nil, err
	}

	cfgs, err := s.scenes.Scenes(ctx)
	if err != nil {
		// Broken scene files are skipped, the readable ones are still listed
		s.logger.From(ctx).Warn("some scene configs could not be read", "error", err)
	}

	result := make([]*scenesv1.Scene, 0, len(cfgs))
	for _, cfg := range cfgs {
		result = append(result, toScene(cfg))
	}

	return connect.NewResponse(&scenesv1.ListScenesResponse{Scenes: result}), nil
}

// Activate implements scenesv1connect.SceneServiceHandler.
func (s *SceneServiceHandler) Activate(ctx context.Context, req *connect.Request[scenesv1.ActivateSceneRequest]) (*connect.Response[scenesv1.ActivateSceneResponse], error) {
	jwt, err := requireRole(ctx, roles.User)
	if err != nil {
		return nil, err
	}

	sceneID := strings.TrimSpace(req.Msg.GetSceneId())
	logger := s.logger.With("userId", jwt.Claims.User.ID, "sceneId", sceneID, "variantId", req.Msg.GetVariantId()).From(ctx)

	t, err := s.scenes.Activate(ctx, sceneID, req.Msg.GetVariantId())
	deviceErrs, err := splitDeviceErrors(t, err)
	if err != nil {
		logger.Error("could not activate scene", "error", err)

		return nil, sceneError(err)
	}
	logger.Info("activated scene", "device_errors", len(deviceErrs))

	return connect.NewResponse(&scenesv1.ActivateSceneResponse{Active: s.active(), DeviceErrors: deviceErrs}), nil
}

// SetVariant implements scenesv1connect.SceneServiceHandler.
func (s *SceneServiceHandler) SetVariant(ctx context.Context, req *connect.Request[scenesv1.SetVariantRequest]) (*connect.Response[scenesv1.SetVariantResponse], error) {
	jwt, err := requireRole(ctx, roles.User)
	if err != nil {
		return nil, err
	}

	logger := s.logger.With("userId", jwt.Claims.User.ID, "variantId", req.Msg.GetVariantId()).From(ctx)

	t, err := s.scenes.SetVariant(ctx, req.Msg.GetVariantId())
	deviceErrs, err := splitDeviceErrors(t, err)
	if err != nil {
		logger.Error("could not switch variant", "error", err)

		return nil, sceneError(err)
	}
	logger.Info("switched variant", "device_errors", len(deviceErrs))

	return connect.NewResponse(&scenesv1.SetVariantResponse{Active: s.active(), DeviceErrors: deviceErrs}), nil
}

// GetActive implements scenesv1connect.SceneServiceHandler.
func (s *SceneServiceHandler) GetActive(ctx context.Context, _ *connect.Request[scenesv1.GetActiveSceneRequest]) (*connect.Response[scenesv1.GetActiveSceneResponse], error) {
	if _, err := requireRole(ctx, roles.User); err != nil {
		return nil, err
	}

	return connect.NewResponse(&scenesv1.GetActiveSceneResponse{Active: s.active()}), nil
}

func (s *SceneServiceHandler) active() *scenesv1.ActiveScene {
	cfg, ref := s.scenes.ActiveConfig()
	if cfg == nil {
		return nil
	}

	return &scenesv1.ActiveScene{Scene: toScene(cfg), VariantId: ref.VariantID}
}

// splitDeviceErrors separates device errors from failures of the change itself: when a transition is
// returned the change happened, and any error only means some devices could not be instructed.
func splitDeviceErrors(t *scenes.Transition, err error) ([]string, error) {
	if err == nil {
		return nil, nil
	}
	if t == nil {
		return nil, err
	}

	var result []string
	if joined, ok := err.(interface{ Unwrap() []error }); ok {
		for _, e := range joined.Unwrap() {
			result = append(result, e.Error())
		}
	} else {
		result = append(result, err.Error())
	}

	return result, nil
}

func sceneError(err error) error {
	switch {
	case errors.Is(err, scenes.ErrSceneNotFound), errors.Is(err, scenes.ErrVariantNotFound):
		return connect.NewError(connect.CodeNotFound, err)
	case errors.Is(err, scenestore.ErrInvalidID):
		return connect.NewError(connect.CodeInvalidArgument, err)
	case errors.Is(err, application.ErrNoActiveScene):
		return connect.NewError(connect.CodeFailedPrecondition, err)
	default:
		return connect.NewError(connect.CodeInternal, err)
	}
}

func toScene(cfg *scenes.Config) *scenesv1.Scene {
	variants := make([]*scenesv1.Variant, 0, len(cfg.Variants))
	for _, v := range cfg.Variants {
		variants = append(variants, &scenesv1.Variant{Id: v.ID, Name: v.Name, Tags: v.Tags})
	}

	return &scenesv1.Scene{
		Id:             cfg.ID,
		Name:           cfg.Name,
		Description:    cfg.Description,
		Variants:       variants,
		DefaultVariant: cfg.DefaultVariant,
	}
}

// requireRole returns the jwt of the request, or a connect error if the request is not authenticated with the given role.
func requireRole(ctx context.Context, role roles.Role) (authentication.JWTContext, error) {
	jwt, err := authentication.JWTFromContext(ctx)
	if err != nil {
		return jwt, connect.NewError(connect.CodeUnauthenticated, err)
	}

	if !jwt.IsAuthenicatedWithRole(role) {
		return jwt, connect.NewError(connect.CodePermissionDenied, errors.New("expecting atleast a "+role.String()+" role"))
	}

	return jwt, nil
}
