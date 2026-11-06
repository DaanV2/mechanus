package screen

import (
	"context"
	"errors"
	"fmt"

	"github.com/DaanV2/mechanus/server/application"
	"github.com/DaanV2/mechanus/server/engine/authentication/roles"
	"github.com/DaanV2/mechanus/server/engine/devices"
	"github.com/DaanV2/mechanus/server/engine/scenes"
	"github.com/DaanV2/mechanus/server/engine/screens"
	"github.com/DaanV2/mechanus/server/engine/screens/state"
	"github.com/DaanV2/mechanus/server/infrastructure/logging"
	"github.com/DaanV2/mechanus/server/pkg/math/vectors"
	screensv1 "github.com/DaanV2/mechanus/server/proto/screens/v1"
)

var _ screens.SceneController = &Controller{}

// ControllerLogger logs the play actions screens send.
const ControllerLogger logging.Prefixed = "screen-control"

var (
	// ErrNotAllowed is returned when a connection may not perform a play action.
	ErrNotAllowed = errors.New("not allowed")
)

// Controller carries out the play actions screens send on the active scene.
//
// Until a GM role and per entity control exist, a token may be moved by any logged in user, the same
// as the scene control API. Paired devices can only move the camera: touch control is not built yet.
// Every action needs the screen to show a battle map of the active scene.
type Controller struct {
	scenes *application.SceneService
}

// NewController creates a controller acting on the given scene service.
func NewController(sceneService *application.SceneService) *Controller {
	return &Controller{scenes: sceneService}
}

// MoveEntity implements screens.SceneController.
func (c *Controller) MoveEntity(ctx context.Context, conn *screens.ConnectionInfo, screenID, entityID string, position vectors.Vector2D) error {
	if !isUser(conn) {
		return screens.Refuse(screensv1.ActionErrorCode_ACTION_ERROR_CODE_PERMISSION_DENIED,
			fmt.Errorf("moving tokens: %w, only logged in users can", ErrNotAllowed))
	}
	if err := c.requireBattleMap(ctx, screenID); err != nil {
		return err
	}

	err := c.scenes.MoveEntity(ctx, entityID, position)
	if err == nil || errors.Is(err, application.ErrDeviceUpdate) {
		ControllerLogger.From(ctx, "userId", conn.ID, "entityId", entityID).
			Debug("moved entity", "x", position.X, "y", position.Y, "device_errors", err)

		return nil
	}

	return refusal(err)
}

// SetCamera implements screens.SceneController.
func (c *Controller) SetCamera(ctx context.Context, _ *screens.ConnectionInfo, screenID string, camera state.Camera) error {
	err := c.scenes.SetCamera(ctx, screenID, scenes.Camera{Center: camera.Center, Zoom: camera.Zoom})
	if err == nil || errors.Is(err, application.ErrDeviceUpdate) {
		if err != nil {
			ControllerLogger.From(ctx, "screenId", screenID).Warn("camera not sent to every device", "error", err)
		}

		return nil
	}

	return refusal(err)
}

func (c *Controller) requireBattleMap(ctx context.Context, screenID string) error {
	ok, err := c.scenes.ShowsBattleMap(ctx, screenID)
	if err != nil {
		return err
	}
	if !ok {
		return screens.Refuse(screensv1.ActionErrorCode_ACTION_ERROR_CODE_FAILED_PRECONDITION,
			fmt.Errorf("screen %q: %w", screenID, application.ErrNoBattleMap))
	}

	return nil
}

// refusal turns the errors of an invalid action into a refusal, any other error stays internal.
func refusal(err error) error {
	switch {
	case errors.Is(err, scenes.ErrInvalidPosition), errors.Is(err, scenes.ErrInvalidCamera):
		return screens.Refuse(screensv1.ActionErrorCode_ACTION_ERROR_CODE_INVALID_ARGUMENT, err)
	case errors.Is(err, scenes.ErrEntityNotFound):
		return screens.Refuse(screensv1.ActionErrorCode_ACTION_ERROR_CODE_NOT_FOUND, err)
	case errors.Is(err, application.ErrNoBattleMap), errors.Is(err, application.ErrNoActiveScene):
		return screens.Refuse(screensv1.ActionErrorCode_ACTION_ERROR_CODE_FAILED_PRECONDITION, err)
	default:
		return err
	}
}

func isUser(conn *screens.ConnectionInfo) bool {
	if conn == nil || conn.Type != devices.DeviceTypeUser {
		return false
	}
	for _, r := range conn.Roles {
		if roles.Role(r).Inherits(roles.User) {
			return true
		}
	}

	return false
}
