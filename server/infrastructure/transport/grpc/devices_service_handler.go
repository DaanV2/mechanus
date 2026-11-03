package grpc

import (
	"context"
	"errors"

	"connectrpc.com/connect"
	"github.com/DaanV2/mechanus/server/application"
	"github.com/DaanV2/mechanus/server/engine/authentication/roles"
	"github.com/DaanV2/mechanus/server/engine/devices"
	"github.com/DaanV2/mechanus/server/infrastructure/authentication"
	"github.com/DaanV2/mechanus/server/infrastructure/logging"
	"github.com/DaanV2/mechanus/server/infrastructure/persistence/models"
	devicesv1 "github.com/DaanV2/mechanus/server/proto/devices/v1"
	"github.com/DaanV2/mechanus/server/proto/devices/v1/devicesv1connect"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var _ devicesv1connect.DeviceServiceHandler = &DeviceServiceHandler{}

// DeviceServicePolicies returns the policy of every procedure of the device service.
func DeviceServicePolicies() map[string]authentication.Policy {
	admin := authentication.PolicyRole(roles.Admin)

	return map[string]authentication.Policy{
		devicesv1connect.DeviceServicePairProcedure:    authentication.PolicyAnonymous,
		devicesv1connect.DeviceServiceRefreshProcedure: authentication.PolicyAnonymous,
		devicesv1connect.DeviceServiceConfirmProcedure: admin,
		devicesv1connect.DeviceServiceListProcedure:    admin,
		devicesv1connect.DeviceServiceRevokeProcedure:  admin,
		devicesv1connect.DeviceServiceBindProcedure:    admin,
	}
}

// DeviceServiceHandler lets devices pair and keep their tokens fresh, and admins manage the devices.
// Who may call what is decided by [DeviceServicePolicies].
type DeviceServiceHandler struct {
	devices *application.DeviceService
	logger  logging.Enriched
}

// NewDeviceServiceHandler creates a new DeviceServiceHandler.
func NewDeviceServiceHandler(deviceService *application.DeviceService) *DeviceServiceHandler {
	return &DeviceServiceHandler{
		devices: deviceService,
		logger:  logging.Enriched{}.WithPrefix("grpc-devices"),
	}
}

// Pair implements devicesv1connect.DeviceServiceHandler.
func (h *DeviceServiceHandler) Pair(ctx context.Context, req *connect.Request[devicesv1.PairRequest], stream *connect.ServerStream[devicesv1.PairResponse]) error {
	deviceID := req.Header().Get(authentication.DeviceIDHeader)
	hardware := devices.ParseHardwareType(req.Header().Get(authentication.DeviceTypeHeader))

	err := h.devices.Pair(ctx, deviceID, hardware, func(event application.PairingEvent) error {
		return stream.Send(toPairResponse(event))
	})

	return h.toError(ctx, err)
}

// Refresh implements devicesv1connect.DeviceServiceHandler.
func (h *DeviceServiceHandler) Refresh(ctx context.Context, req *connect.Request[devicesv1.RefreshDeviceRequest]) (*connect.Response[devicesv1.RefreshDeviceResponse], error) {
	tokens, err := h.devices.Refresh(ctx, req.Msg.GetToken())
	if err != nil {
		return nil, h.toError(ctx, err)
	}

	return connect.NewResponse(&devicesv1.RefreshDeviceResponse{Tokens: toDeviceTokens(tokens)}), nil
}

// Confirm implements devicesv1connect.DeviceServiceHandler.
func (h *DeviceServiceHandler) Confirm(ctx context.Context, req *connect.Request[devicesv1.ConfirmPairingRequest]) (*connect.Response[devicesv1.ConfirmPairingResponse], error) {
	identity := authentication.IdentityFromContext(ctx)

	device, err := h.devices.Confirm(ctx, req.Msg.GetCode(), req.Msg.GetName(), identity.ID)
	if err != nil {
		return nil, h.toError(ctx, err)
	}

	return connect.NewResponse(&devicesv1.ConfirmPairingResponse{Device: toDevice(device)}), nil
}

// List implements devicesv1connect.DeviceServiceHandler.
func (h *DeviceServiceHandler) List(ctx context.Context, _ *connect.Request[devicesv1.ListDevicesRequest]) (*connect.Response[devicesv1.ListDevicesResponse], error) {
	list, err := h.devices.List(ctx)
	if err != nil {
		return nil, h.toError(ctx, err)
	}

	result := make([]*devicesv1.Device, 0, len(list))
	for _, info := range list {
		device := &devicesv1.Device{
			Id:        info.ID,
			Name:      info.Name,
			Hardware:  info.Hardware,
			Target:    info.Target,
			Paired:    info.Paired,
			Connected: info.Connected,
		}
		if info.Paired {
			device.PairedAt = timestamppb.New(info.PairedAt)
			device.LastSeenAt = timestamppb.New(info.LastSeenAt)
		}
		result = append(result, device)
	}

	return connect.NewResponse(&devicesv1.ListDevicesResponse{Devices: result}), nil
}

// Revoke implements devicesv1connect.DeviceServiceHandler.
func (h *DeviceServiceHandler) Revoke(ctx context.Context, req *connect.Request[devicesv1.RevokeDeviceRequest]) (*connect.Response[devicesv1.RevokeDeviceResponse], error) {
	if err := h.devices.Revoke(ctx, req.Msg.GetId()); err != nil {
		return nil, h.toError(ctx, err)
	}

	h.logger.From(ctx).Info("device revoked", "deviceId", req.Msg.GetId())

	return connect.NewResponse(&devicesv1.RevokeDeviceResponse{}), nil
}

// Bind implements devicesv1connect.DeviceServiceHandler.
func (h *DeviceServiceHandler) Bind(ctx context.Context, req *connect.Request[devicesv1.BindDeviceRequest]) (*connect.Response[devicesv1.BindDeviceResponse], error) {
	device, err := h.devices.Bind(ctx, req.Msg.GetId(), req.Msg.GetTarget())
	if err != nil {
		return nil, h.toError(ctx, err)
	}

	return connect.NewResponse(&devicesv1.BindDeviceResponse{Device: toDevice(device)}), nil
}

// toError maps the errors of the device service to connect errors.
func (h *DeviceServiceHandler) toError(ctx context.Context, err error) error {
	switch {
	case err == nil:
		return nil
	case errors.Is(err, context.Canceled), errors.Is(err, context.DeadlineExceeded):
		return connect.NewError(connect.CodeCanceled, err) // The device left, nothing to report
	case errors.Is(err, application.ErrDeviceIDInvalid),
		errors.Is(err, application.ErrDeviceNameInvalid),
		errors.Is(err, application.ErrDeviceTargetInvalid):
		return connect.NewError(connect.CodeInvalidArgument, err)
	case errors.Is(err, application.ErrPairingCodeInvalid):
		return connect.NewError(connect.CodeFailedPrecondition, application.ErrPairingCodeInvalid)
	case errors.Is(err, application.ErrTooManyPairingAttempts), errors.Is(err, application.ErrPairingBusy):
		return connect.NewError(connect.CodeResourceExhausted, err)
	case errors.Is(err, application.ErrPairingReplaced):
		return connect.NewError(connect.CodeAborted, err)
	case errors.Is(err, application.ErrPairingStopped):
		return connect.NewError(connect.CodeUnavailable, err)
	case errors.Is(err, application.ErrDeviceTokenInvalid):
		return connect.NewError(connect.CodeUnauthenticated, application.ErrDeviceTokenInvalid)
	case errors.Is(err, devices.ErrDeviceNotFound):
		return connect.NewError(connect.CodeNotFound, err)
	}

	h.logger.From(ctx).Error("device service error", "error", err)

	return connect.NewError(connect.CodeInternal, err)
}

func toPairResponse(event application.PairingEvent) *devicesv1.PairResponse {
	if event.Tokens != nil {
		return &devicesv1.PairResponse{Event: &devicesv1.PairResponse_Tokens{Tokens: toDeviceTokens(event.Tokens)}}
	}

	return &devicesv1.PairResponse{Event: &devicesv1.PairResponse_Code{Code: &devicesv1.PairingCode{
		Code:      event.Code,
		ExpiresAt: timestamppb.New(event.ExpiresAt),
	}}}
}

func toDeviceTokens(tokens *authentication.TokenPair) *devicesv1.DeviceTokens {
	return &devicesv1.DeviceTokens{
		Token:        tokens.Access,
		Type:         authentication.DeviceTokenType,
		RefreshToken: tokens.Refresh,
	}
}

func toDevice(device *models.Device) *devicesv1.Device {
	return &devicesv1.Device{
		Id:         device.ID,
		Name:       device.Name,
		Hardware:   device.Hardware,
		Target:     device.Target,
		Paired:     true,
		PairedAt:   timestamppb.New(device.PairedAt),
		LastSeenAt: timestamppb.New(device.LastSeenAt),
	}
}
