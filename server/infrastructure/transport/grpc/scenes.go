package grpc

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"connectrpc.com/connect"
	"github.com/DaanV2/mechanus/server/application"
	"github.com/DaanV2/mechanus/server/engine/authentication/roles"
	"github.com/DaanV2/mechanus/server/engine/devices"
	"github.com/DaanV2/mechanus/server/engine/scenes"
	"github.com/DaanV2/mechanus/server/infrastructure/logging"
	"github.com/DaanV2/mechanus/server/infrastructure/scenestore"
	"github.com/DaanV2/mechanus/server/pkg/math/vectors"
	scenesv1 "github.com/DaanV2/mechanus/server/proto/scenes/v1"
	"github.com/DaanV2/mechanus/server/proto/scenes/v1/scenesv1connect"
)

var _ scenesv1connect.SceneServiceHandler = &SceneServiceHandler{}

const (
	SceneLogger logging.Prefixed = "grpc-scenes"
)

// DeviceStatuses reports the health / connection status of the devices in the room, see [application.DeviceMonitor].
type DeviceStatuses interface {
	Statuses(ctx context.Context) ([]devices.Status, error)
}

// SceneServiceHandler implements the scenes gRPC service, the GM's control over the active scene.
type SceneServiceHandler struct {
	scenes  *application.SceneService
	devices DeviceStatuses
}

// NewSceneServiceHandler creates a new SceneServiceHandler.
func NewSceneServiceHandler(sceneService *application.SceneService, deviceStatuses DeviceStatuses) *SceneServiceHandler {
	return &SceneServiceHandler{
		scenes:  sceneService,
		devices: deviceStatuses,
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
		SceneLogger.From(ctx).Warn("some scene configs could not be read", "error", err)
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
	logger := SceneLogger.From(ctx, "userId", jwt.Claims.User.ID, "sceneId", sceneID, "variantId", req.Msg.GetVariantId())

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

	logger := SceneLogger.From(ctx, "userId", jwt.Claims.User.ID, "variantId", req.Msg.GetVariantId())

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

	return &scenesv1.ActiveScene{Scene: toScene(cfg), VariantId: ref.VariantID, Entities: toEntities(s.scenes.ActiveEntities())}
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
	case errors.Is(err, scenes.ErrEntityNotFound):
		return connect.NewError(connect.CodeNotFound, err)
	case errors.Is(err, scenes.ErrEntityExists):
		return connect.NewError(connect.CodeAlreadyExists, err)
	case errors.Is(err, scenestore.ErrInvalidID), errors.Is(err, scenes.ErrInvalidEntity), errors.Is(err, scenes.ErrInvalidPosition):
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

// SetEntityHidden implements scenesv1connect.SceneServiceHandler.
func (s *SceneServiceHandler) SetEntityHidden(ctx context.Context, req *connect.Request[scenesv1.SetEntityHiddenRequest]) (*connect.Response[scenesv1.SetEntityHiddenResponse], error) {
	jwt, err := requireRole(ctx, roles.User)
	if err != nil {
		return nil, err
	}

	entityID := req.Msg.GetEntityId()
	logger := SceneLogger.From(ctx, "userId", jwt.Claims.User.ID, "entityId", entityID, "hidden", req.Msg.GetHidden())

	t, err := s.scenes.SetEntityHidden(ctx, entityID, req.Msg.GetHidden())
	deviceErrs, err := splitDeviceErrors(t, err)
	if err != nil {
		logger.Warn("could not hide / reveal entity", "error", err)

		return nil, sceneError(err)
	}
	logger.Info("changed entity visibility", "device_errors", len(deviceErrs))

	return connect.NewResponse(&scenesv1.SetEntityHiddenResponse{Active: s.active(), DeviceErrors: deviceErrs}), nil
}

// SpawnEntity implements scenesv1connect.SceneServiceHandler.
func (s *SceneServiceHandler) SpawnEntity(ctx context.Context, req *connect.Request[scenesv1.SpawnEntityRequest]) (*connect.Response[scenesv1.SpawnEntityResponse], error) {
	jwt, err := requireRole(ctx, roles.User)
	if err != nil {
		return nil, err
	}

	entity := scenes.Entity{
		ID:     strings.TrimSpace(req.Msg.GetEntityId()),
		Kind:   fromEntityKind(req.Msg.GetKind()),
		Name:   strings.TrimSpace(req.Msg.GetName()),
		Hidden: req.Msg.GetHidden(),
	}
	var position *vectors.Vector2D
	if p := req.Msg.GetPosition(); p != nil {
		position = &vectors.Vector2D{X: p.GetX(), Y: p.GetY()}
	}
	logger := SceneLogger.From(ctx, "userId", jwt.Claims.User.ID, "entityId", entity.ID)

	t, err := s.scenes.SpawnEntity(ctx, entity, position)
	deviceErrs, err := splitDeviceErrors(t, err)
	if err != nil {
		logger.Warn("could not spawn entity", "error", err)

		return nil, sceneError(err)
	}
	logger.Info("spawned entity", "device_errors", len(deviceErrs))

	return connect.NewResponse(&scenesv1.SpawnEntityResponse{Active: s.active(), DeviceErrors: deviceErrs}), nil
}

// RemoveEntity implements scenesv1connect.SceneServiceHandler.
func (s *SceneServiceHandler) RemoveEntity(ctx context.Context, req *connect.Request[scenesv1.RemoveEntityRequest]) (*connect.Response[scenesv1.RemoveEntityResponse], error) {
	jwt, err := requireRole(ctx, roles.User)
	if err != nil {
		return nil, err
	}

	entityID := req.Msg.GetEntityId()
	logger := SceneLogger.From(ctx, "userId", jwt.Claims.User.ID, "entityId", entityID)

	t, err := s.scenes.RemoveEntity(ctx, entityID)
	deviceErrs, err := splitDeviceErrors(t, err)
	if err != nil {
		logger.Warn("could not remove entity", "error", err)

		return nil, sceneError(err)
	}
	logger.Info("removed entity", "device_errors", len(deviceErrs))

	return connect.NewResponse(&scenesv1.RemoveEntityResponse{Active: s.active(), DeviceErrors: deviceErrs}), nil
}

// MoveEntity implements scenesv1connect.SceneServiceHandler.
func (s *SceneServiceHandler) MoveEntity(ctx context.Context, req *connect.Request[scenesv1.MoveEntityRequest]) (*connect.Response[scenesv1.MoveEntityResponse], error) {
	jwt, err := requireRole(ctx, roles.User)
	if err != nil {
		return nil, err
	}

	entityID := req.Msg.GetEntityId()
	p := req.Msg.GetPosition()
	logger := SceneLogger.From(ctx, "userId", jwt.Claims.User.ID, "entityId", entityID)
	if p == nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("%w: no position", scenes.ErrInvalidPosition))
	}

	var deviceErrs []string
	err = s.scenes.MoveEntity(ctx, entityID, vectors.Vector2D{X: p.GetX(), Y: p.GetY()})
	if errors.Is(err, application.ErrDeviceUpdate) {
		deviceErrs, err = []string{err.Error()}, nil
	}
	if err != nil {
		logger.Warn("could not move entity", "error", err)

		return nil, sceneError(err)
	}
	logger.Debug("moved entity", "x", p.GetX(), "y", p.GetY())

	return connect.NewResponse(&scenesv1.MoveEntityResponse{Active: s.active(), DeviceErrors: deviceErrs}), nil
}

func toEntityKind(kind scenes.EntityKind) scenesv1.EntityKind {
	switch kind {
	case scenes.EntityPlayer:
		return scenesv1.EntityKind_ENTITY_KIND_PLAYER
	case scenes.EntityNPC:
		return scenesv1.EntityKind_ENTITY_KIND_NPC
	case scenes.EntityEnemy:
		return scenesv1.EntityKind_ENTITY_KIND_ENEMY
	case scenes.EntityObject:
		return scenesv1.EntityKind_ENTITY_KIND_OBJECT
	default:
		return scenesv1.EntityKind_ENTITY_KIND_UNSPECIFIED
	}
}

// fromEntityKind returns the scene entity kind, empty (and so refused by a spawn) for an unspecified kind.
func fromEntityKind(kind scenesv1.EntityKind) scenes.EntityKind {
	switch kind {
	case scenesv1.EntityKind_ENTITY_KIND_PLAYER:
		return scenes.EntityPlayer
	case scenesv1.EntityKind_ENTITY_KIND_NPC:
		return scenes.EntityNPC
	case scenesv1.EntityKind_ENTITY_KIND_ENEMY:
		return scenes.EntityEnemy
	case scenesv1.EntityKind_ENTITY_KIND_OBJECT:
		return scenes.EntityObject
	case scenesv1.EntityKind_ENTITY_KIND_UNSPECIFIED:
	}

	return ""
}

func toEntities(entities []application.ActiveEntity) []*scenesv1.Entity {
	result := make([]*scenesv1.Entity, 0, len(entities))
	for _, e := range entities {
		pe := &scenesv1.Entity{Id: e.ID, Kind: toEntityKind(e.Kind), Name: e.Name, Hidden: e.Hidden}
		if e.Position != nil {
			pe.Position = &scenesv1.Position{X: e.Position.X, Y: e.Position.Y}
		}
		result = append(result, pe)
	}

	return result
}
