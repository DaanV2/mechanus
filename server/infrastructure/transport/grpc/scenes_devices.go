package grpc

import (
	"context"
	"fmt"

	"connectrpc.com/connect"
	"github.com/DaanV2/mechanus/server/engine/authentication/roles"
	"github.com/DaanV2/mechanus/server/engine/devices"
	scenesv1 "github.com/DaanV2/mechanus/server/proto/scenes/v1"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// ListDeviceStatuses implements scenesv1connect.SceneServiceHandler.
func (s *SceneServiceHandler) ListDeviceStatuses(ctx context.Context, _ *connect.Request[scenesv1.ListDeviceStatusesRequest]) (*connect.Response[scenesv1.ListDeviceStatusesResponse], error) {
	if _, err := requireRole(ctx, roles.User); err != nil {
		return nil, err
	}

	statuses, err := s.devices.Statuses(ctx)
	if err != nil {
		SceneLogger.From(ctx).Error("could not read the device statuses", "error", err)

		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("reading device statuses: %w", err))
	}

	result := make([]*scenesv1.DeviceStatus, 0, len(statuses))
	for i := range statuses {
		result = append(result, toDeviceStatus(&statuses[i]))
	}

	return connect.NewResponse(&scenesv1.ListDeviceStatusesResponse{Devices: result}), nil
}

func toDeviceStatus(status *devices.Status) *scenesv1.DeviceStatus {
	result := &scenesv1.DeviceStatus{
		DeviceId:       status.DeviceID,
		Name:           status.Name,
		Kind:           string(status.Kind),
		Integration:    status.Integration,
		Targets:        status.Targets,
		State:          toDeviceState(status.State),
		Error:          status.Error,
		LastApplyError: status.LastApplyError,
	}
	if !status.LastApply.IsZero() {
		result.LastApply = timestamppb.New(status.LastApply)
	}

	return result
}

func toDeviceState(state devices.State) scenesv1.DeviceState {
	switch state {
	case devices.StateOnline:
		return scenesv1.DeviceState_DEVICE_STATE_ONLINE
	case devices.StateOffline:
		return scenesv1.DeviceState_DEVICE_STATE_OFFLINE
	case devices.StateUnknown:
	}

	return scenesv1.DeviceState_DEVICE_STATE_UNSPECIFIED
}
