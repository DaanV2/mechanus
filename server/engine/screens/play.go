package screens

import (
	"context"
	"errors"
	"fmt"

	"github.com/DaanV2/mechanus/server/engine/screens/state"
	"github.com/DaanV2/mechanus/server/pkg/math/vectors"
	screensv1 "github.com/DaanV2/mechanus/server/proto/screens/v1"
)

// SceneController carries out the play actions screens send, such as moving a token or the camera.
// The screen ID is the device ID of the screen the connection looks at. A refused action returns a
// [RefusedError], which is sent back to the client; any other error is reported as an internal error.
// Neither closes the connection.
type SceneController interface {
	MoveEntity(ctx context.Context, conn *ConnectionInfo, screenID, entityID string, position vectors.Vector2D) error
	SetCamera(ctx context.Context, conn *ConnectionInfo, screenID string, camera state.Camera) error
}

// RefusedError is an action that was refused, with the reason the client is told.
type RefusedError struct {
	Code screensv1.ActionErrorCode
	Err  error
}

// Refuse returns a [RefusedError].
func Refuse(code screensv1.ActionErrorCode, err error) error {
	return &RefusedError{Code: code, Err: err}
}

func (e *RefusedError) Error() string {
	return fmt.Sprintf("refused (%s): %v", e.Code, e.Err)
}

func (e *RefusedError) Unwrap() error {
	return e.Err
}

// errNoController is reported when a play action arrives on a server without a scene controller.
var errNoController = errors.New("play actions are not available")

// HandleMoveEntity hands a token move to the scene controller.
func (s *ScreenHandler) HandleMoveEntity(ctx context.Context, listener *ScreenConn, msg *screensv1.MoveEntity) []*screensv1.ServerMessage {
	ctrl := s.control()
	if ctrl == nil {
		return actionError(Refuse(screensv1.ActionErrorCode_ACTION_ERROR_CODE_FAILED_PRECONDITION, errNoController))
	}
	pos := msg.GetPosition()
	err := ctrl.MoveEntity(ctx, listener.info, s.id, msg.GetEntityId(), vectors.Vector2D{X: pos.GetX(), Y: pos.GetY()})

	return s.actionResult(ctx, "move entity", err)
}

// HandleSetCamera hands a camera move to the scene controller.
func (s *ScreenHandler) HandleSetCamera(ctx context.Context, listener *ScreenConn, msg *screensv1.SetCamera) []*screensv1.ServerMessage {
	ctrl := s.control()
	if ctrl == nil {
		return actionError(Refuse(screensv1.ActionErrorCode_ACTION_ERROR_CODE_FAILED_PRECONDITION, errNoController))
	}
	c := msg.GetCamera()
	camera := state.Camera{
		Center: vectors.Vector2D{X: c.GetCenter().GetX(), Y: c.GetCenter().GetY()},
		Zoom:   c.GetZoom(),
	}
	err := ctrl.SetCamera(ctx, listener.info, s.id, camera)

	return s.actionResult(ctx, "set camera", err)
}

func (s *ScreenHandler) actionResult(ctx context.Context, action string, err error) []*screensv1.ServerMessage {
	if err == nil {
		return nil
	}

	var refused *RefusedError
	if errors.As(err, &refused) {
		s.logger.From(ctx).Debug("action refused", "action", action, "error", err)
	} else {
		s.logger.From(ctx).Error("action failed", "action", action, "error", err)
	}

	return actionError(err)
}

func actionError(err error) []*screensv1.ServerMessage {
	code := screensv1.ActionErrorCode_ACTION_ERROR_CODE_INTERNAL
	message := "internal error"
	var refused *RefusedError
	if errors.As(err, &refused) {
		code = refused.Code
		message = refused.Err.Error()
	}

	return []*screensv1.ServerMessage{{
		Action: &screensv1.ServerMessage_ActionError{
			ActionError: &screensv1.ActionError{Code: code, Message: message},
		},
	}}
}

// MoveEntity moves an entity's token on the screen and sends it to every listener. Returns false,
// and sends nothing, when the screen doesn't draw the entity.
func (s *ScreenHandler) MoveEntity(ctx context.Context, entityID string, position vectors.Vector2D) bool {
	s.stateLock.Lock()
	var update *screensv1.EntityUpdate
	for _, layer := range s.state.Layers {
		if e, ok := layer.Entities[entityID]; ok {
			e.Position = position
			update = &screensv1.EntityUpdate{LayerId: layer.ID, Entity: e.ToProto()}

			break
		}
	}
	s.stateLock.Unlock()

	if update == nil {
		return false
	}
	s.Broadcast(ctx, &screensv1.ServerMessage{
		Action: &screensv1.ServerMessage_EntityUpdate{EntityUpdate: update},
	})

	return true
}

// SetCamera moves the part of the view the screen shows and sends it to every listener.
func (s *ScreenHandler) SetCamera(ctx context.Context, camera state.Camera) {
	s.stateLock.Lock()
	s.state.Camera = &camera
	msg := camera.ToProto()
	s.stateLock.Unlock()

	s.Broadcast(ctx, &screensv1.ServerMessage{
		Action: &screensv1.ServerMessage_CameraUpdate{CameraUpdate: msg},
	})
}

func (s *ScreenHandler) control() SceneController {
	if s.controller == nil {
		return nil
	}

	return s.controller()
}
