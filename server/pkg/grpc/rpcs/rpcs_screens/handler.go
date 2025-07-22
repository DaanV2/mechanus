package rpcs_screens

import (
	"context"
	"errors"
	"strings"

	"connectrpc.com/connect"
	"github.com/DaanV2/mechanus/server/mechanus/scenes"
	"github.com/DaanV2/mechanus/server/pkg/authenication"
	"github.com/DaanV2/mechanus/server/pkg/authenication/roles"
	screensv1 "github.com/DaanV2/mechanus/server/pkg/grpc/gen/screens/v1"
	"github.com/DaanV2/mechanus/server/pkg/grpc/gen/screens/v1/screensv1connect"
)

var _ screensv1connect.ScreensServiceHandler = &ScreenService{}

type ScreenService struct {
	manager *scenes.Manager
}

func NewScreenService(manager *scenes.Manager) *ScreenService {
	return &ScreenService{
		manager,
	}
}

// ListenUpdate implements screensv1connect.ScreensServiceHandler.
func (s *ScreenService) ListenUpdate(ctx context.Context, req *connect.Request[screensv1.ScreenListenRequest], str *connect.ServerStream[screensv1.ScreenUpdate]) error {
	auth := authenicateRequest(ctx, req.Msg)
	if !auth.valid {
		err := errors.New("unauthorized access to screen " + auth.id + " with role " + auth.role)

		return connect.NewError(connect.CodePermissionDenied, err)
	}

	return nil
}

type screenAuthenication struct {
	role  string
	valid bool
	id    string
}

func authenicateRequest(ctx context.Context, req *screensv1.ScreenListenRequest) screenAuthenication {
	info := screenAuthenication{
		role:  strings.ToLower(req.GetRole().String()),
		valid: false,
		id:    req.GetId(),
	}

	switch req.GetRole() {
	case screensv1.ScreenRole_SCREEN_ROLE_ADMIN:
		info.valid = authenication.IsAuthenicatedWithRole(ctx, roles.Admin)
	case screensv1.ScreenRole_SCREEN_ROLE_OPERATOR:
		info.valid = authenication.IsAuthenicatedWithRole(ctx, roles.Operator)
	case screensv1.ScreenRole_SCREEN_ROLE_PLAYER:
		info.valid = authenication.IsAuthenicatedWithRole(ctx, roles.User)
	case screensv1.ScreenRole_SCREEN_ROLE_VIEWER:
		info.valid = authenication.IsAuthenicatedWithRole(ctx, roles.Viewer)
	case screensv1.ScreenRole_SCREEN_ROLE_UNKNOWN_UNSPECIFIED, screensv1.ScreenRole_SCREEN_ROLE_DEVICE:
		info.valid = true // Unknown/Device roles do not require authentication
	}

	return info
}
