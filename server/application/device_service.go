package application

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"maps"
	"math/big"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/DaanV2/mechanus/server/engine/devices"
	"github.com/DaanV2/mechanus/server/engine/screens"
	"github.com/DaanV2/mechanus/server/infrastructure/authentication"
	"github.com/DaanV2/mechanus/server/infrastructure/lifecycle"
	"github.com/DaanV2/mechanus/server/infrastructure/logging"
	"github.com/DaanV2/mechanus/server/infrastructure/persistence"
	"github.com/DaanV2/mechanus/server/infrastructure/persistence/models"
	"github.com/DaanV2/mechanus/server/infrastructure/persistence/repositories"
)

var (
	// ErrDeviceIDInvalid is returned when a device reports no usable device ID.
	ErrDeviceIDInvalid = errors.New("device ID is missing or invalid")
	// ErrDeviceNameInvalid is returned when a device is given no usable name.
	ErrDeviceNameInvalid = errors.New("device name is missing or too long")
	// ErrDeviceTargetInvalid is returned when a device is bound to an unusable target.
	ErrDeviceTargetInvalid = errors.New("device target is invalid")
	// ErrDeviceTokenInvalid is returned when a device refresh token isn't accepted.
	ErrDeviceTokenInvalid = errors.New("device token is invalid, expired or revoked")
	// ErrPairingCodeInvalid is returned when a code doesn't belong to a device that is pairing right now.
	ErrPairingCodeInvalid = errors.New("pairing code is invalid or expired")
	// ErrTooManyPairingAttempts is returned when too many wrong codes were entered.
	ErrTooManyPairingAttempts = errors.New("too many wrong pairing codes, try again later")
	// ErrPairingBusy is returned when too many devices are pairing at once.
	ErrPairingBusy = errors.New("too many devices are pairing, try again later")
	// ErrPairingReplaced is returned to a pairing stream when the same device started pairing on another one.
	ErrPairingReplaced = errors.New("the device started pairing again")
	// ErrPairingStopped is returned when the server is shutting down.
	ErrPairingStopped = errors.New("the server is shutting down")
)

const maxDeviceNameLength = 64

var (
	_ lifecycle.AfterInitialize = &DeviceService{}
	_ lifecycle.BeforeShutdown  = &DeviceService{}
)

type (
	// PairingOptions limits how pairing codes can be used.
	PairingOptions struct {
		// How long a code stays valid before the device gets a new one
		CodeLifetime time.Duration
		// How many wrong codes someone can enter within GuessWindow. Reaching this replaces every open code,
		// as a wrong guess can't be tied to one code, and refuses more attempts until the window ends
		MaxWrongCodes int
		GuessWindow   time.Duration
		// How many devices can pair at once. The route needs no login, and every open code makes a guess more likely to hit
		MaxSessions int
	}

	// PairingEvent is what a device gets told while pairing: a code to show, or at the end its tokens.
	PairingEvent struct {
		Code      string
		ExpiresAt time.Time
		Tokens    *authentication.TokenPair
	}

	// DeviceInfo describes a device that is paired, connected or both.
	DeviceInfo struct {
		ID         string
		Name       string
		Hardware   string
		Target     string
		Paired     bool
		Connected  bool
		PairedAt   time.Time
		LastSeenAt time.Time
	}

	// DeviceService registers devices by pairing them, and manages their tokens.
	DeviceService struct {
		repo    *repositories.DeviceRepository
		jwts    *authentication.JWTService
		jtis    *authentication.JTIService
		screens *screens.ScreenManager
		options PairingOptions
		logger  logging.Enriched

		lock     sync.Mutex // Guards everything below, and the fields of the sessions
		byCode   map[string]*pairingSession
		byDevice map[string]*pairingSession
		wrong    map[string]*wrongCodes
		stopped  bool
	}

	pairingSession struct {
		deviceID string
		hardware devices.HardwareType
		code     string
		closed   bool
		renew    chan struct{}                  // Asks for a new code
		paired   chan *authentication.TokenPair // Holds the tokens once paired
		ended    chan error                     // Holds why the stream has to end
	}

	wrongCodes struct {
		count int
		first time.Time
	}

	// pairedRoom adds the bindings of paired devices to the authored room.
	pairedRoom struct {
		authored RoomConfigStore
		devices  *DeviceService
	}
)

// DefaultPairingOptions returns the default limits on pairing codes.
func DefaultPairingOptions() PairingOptions {
	return PairingOptions{
		CodeLifetime:  5 * time.Minute,
		MaxWrongCodes: 5,
		GuessWindow:   time.Minute,
		MaxSessions:   64,
	}
}

// NewDeviceService creates a new DeviceService with the default pairing limits.
func NewDeviceService(repo *repositories.DeviceRepository, jwts *authentication.JWTService, jtis *authentication.JTIService, screenManager *screens.ScreenManager) *DeviceService {
	return &DeviceService{
		repo:     repo,
		jwts:     jwts,
		jtis:     jtis,
		screens:  screenManager,
		options:  DefaultPairingOptions(),
		logger:   logging.Enriched{}.WithPrefix("devices"),
		byCode:   map[string]*pairingSession{},
		byDevice: map[string]*pairingSession{},
		wrong:    map[string]*wrongCodes{},
	}
}

// WithPairingOptions replaces the limits on pairing codes.
func (s *DeviceService) WithPairingOptions(options PairingOptions) *DeviceService {
	s.options = options

	return s
}

// AfterInitialize implements lifecycle.AfterInitialize, it forgets the devices that have to pair again by now.
func (s *DeviceService) AfterInitialize(ctx context.Context) error {
	if err := s.repo.DeleteNotSeenSince(ctx, s.seenCutoff()); err != nil {
		return fmt.Errorf("cleaning up devices: %w", err)
	}

	return nil
}

// BeforeShutdown implements lifecycle.BeforeShutdown, it ends the open pairing streams so they don't hold up the server.
func (s *DeviceService) BeforeShutdown(_ context.Context) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.stopped = true
	for _, session := range s.byDevice {
		session.end(ErrPairingStopped)
	}

	return nil
}

// Pair runs the device's side of pairing until it is paired or ctx ends, which is when its code stops being valid.
// Every event is handed to send, an error of send ends pairing.
func (s *DeviceService) Pair(ctx context.Context, deviceID string, hardware devices.HardwareType, send func(PairingEvent) error) error {
	if !devices.ValidID(deviceID) {
		return ErrDeviceIDInvalid
	}

	session, err := s.openPairing(deviceID, hardware)
	if err != nil {
		return err
	}
	defer s.closePairing(ctx, session)

	for {
		event, err := s.newCode(session)
		if err != nil {
			return err
		}
		if err := send(event); err != nil {
			return err
		}

		tokens, err := s.waitForPairing(ctx, session)
		if err != nil {
			return err
		}
		if tokens == nil {
			continue
		}

		if err := send(PairingEvent{Tokens: tokens}); err != nil {
			// The admin saw it succeed, but the device never got its tokens: don't leave a device listed that isn't paired
			s.logger.From(ctx).Warn("device tokens could not be delivered, revoking", "deviceId", deviceID, "error", err)

			return errors.Join(err, s.revokeDetached(ctx, deviceID))
		}

		return nil
	}
}

// waitForPairing returns the tokens once paired, or nil when the device needs a new code.
func (s *DeviceService) waitForPairing(ctx context.Context, session *pairingSession) (*authentication.TokenPair, error) {
	timer := time.NewTimer(s.options.CodeLifetime)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case err := <-session.ended:
		return nil, err
	case tokens := <-session.paired:
		return tokens, nil
	case <-session.renew:
		return nil, nil //nolint:nilnil // Nothing went wrong, and there are no tokens yet
	case <-timer.C:
		return nil, nil //nolint:nilnil // Same as above
	}
}

// openPairing registers a new session. A session of the same device gets replaced: its code stops working.
func (s *DeviceService) openPairing(deviceID string, hardware devices.HardwareType) (*pairingSession, error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	if s.stopped {
		return nil, ErrPairingStopped
	}

	old, replacing := s.byDevice[deviceID]
	if !replacing && len(s.byDevice) >= s.options.MaxSessions {
		return nil, ErrPairingBusy
	}
	if replacing {
		delete(s.byCode, old.code)
		old.code = ""
		old.closed = true
		old.end(ErrPairingReplaced)
	}

	session := &pairingSession{
		deviceID: deviceID,
		hardware: hardware,
		renew:    make(chan struct{}, 1),
		paired:   make(chan *authentication.TokenPair, 1),
		ended:    make(chan error, 1),
	}
	s.byDevice[deviceID] = session

	return session, nil
}

// closePairing forgets a session whose stream ended. Tokens it never got to deliver are revoked again.
func (s *DeviceService) closePairing(ctx context.Context, session *pairingSession) {
	s.lock.Lock()
	delete(s.byCode, session.code)
	session.code = ""
	session.closed = true
	if s.byDevice[session.deviceID] == session {
		delete(s.byDevice, session.deviceID)
	}
	s.lock.Unlock()

	select {
	case <-session.paired:
		if err := s.revokeDetached(ctx, session.deviceID); err != nil {
			s.logger.From(ctx).Error("could not revoke undelivered device tokens", "deviceId", session.deviceID, "error", err)
		}
	default:
	}
}

// revokeDetached revokes a device from a context that might have ended already, as the stream that was using it did.
func (s *DeviceService) revokeDetached(ctx context.Context, deviceID string) error {
	err := s.Revoke(context.WithoutCancel(ctx), deviceID)
	if errors.Is(err, devices.ErrDeviceNotFound) {
		return nil
	}

	return err
}

func (session *pairingSession) end(err error) {
	select {
	case session.ended <- err:
	default:
	}
}

func (session *pairingSession) requestCode() {
	select {
	case session.renew <- struct{}{}:
	default:
	}
}

// newCode replaces the code of the session, the old one stops working.
func (s *DeviceService) newCode(session *pairingSession) (PairingEvent, error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	delete(s.byCode, session.code)
	for {
		n, err := rand.Int(rand.Reader, big.NewInt(10_000))
		if err != nil {
			return PairingEvent{}, fmt.Errorf("generating pairing code: %w", err)
		}
		// Terminates: there are far fewer sessions (MaxSessions) than codes
		code := fmt.Sprintf("%04d", n.Int64())
		if _, taken := s.byCode[code]; taken {
			continue
		}

		session.code = code
		s.byCode[code] = session

		return PairingEvent{Code: code, ExpiresAt: time.Now().Add(s.options.CodeLifetime)}, nil
	}
}

// Confirm pairs the device showing the code on behalf of confirmedBy. Pairing again replaces the old tokens of the device.
func (s *DeviceService) Confirm(ctx context.Context, code, name, confirmedBy string) (*models.Device, error) {
	name = strings.TrimSpace(name)
	if name == "" || len(name) > maxDeviceNameLength {
		return nil, ErrDeviceNameInvalid
	}

	session, err := s.claimCode(code, confirmedBy)
	if err != nil {
		return nil, err
	}

	device, tokens, err := s.register(ctx, session, name, confirmedBy)
	if err != nil {
		session.requestCode() // The code is used up, without this the device keeps showing a dead code

		return nil, err
	}

	if !s.deliver(session, tokens) {
		// The device left in the meantime
		if err := s.revokeDetached(ctx, session.deviceID); err != nil {
			return nil, err
		}

		return nil, ErrPairingCodeInvalid
	}

	s.logger.From(ctx).Info("device paired", "deviceId", device.ID, "name", device.Name, "pairedBy", confirmedBy)

	return device, nil
}

// claimCode uses up the code, and counts it against confirmedBy when it is wrong.
func (s *DeviceService) claimCode(code, confirmedBy string) (*pairingSession, error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	wrong, ok := s.wrong[confirmedBy]
	if ok && time.Since(wrong.first) > s.options.GuessWindow {
		delete(s.wrong, confirmedBy)
		wrong, ok = nil, false
	}
	if ok && wrong.count >= s.options.MaxWrongCodes {
		return nil, ErrTooManyPairingAttempts
	}

	session, found := s.byCode[code]
	if !found {
		if !ok {
			wrong = &wrongCodes{first: time.Now()}
			s.wrong[confirmedBy] = wrong
		}
		wrong.count++
		if wrong.count >= s.options.MaxWrongCodes {
			s.invalidateCodes()
		}

		return nil, ErrPairingCodeInvalid
	}

	delete(s.byCode, code)
	session.code = ""
	delete(s.wrong, confirmedBy)

	return session, nil
}

// invalidateCodes makes every device get a new code, whatever got guessed so far is now worthless.
func (s *DeviceService) invalidateCodes() {
	for code, session := range s.byCode {
		delete(s.byCode, code)
		session.code = ""
		session.requestCode()
	}
}

// register stores the device and starts its session, revoking the session it had when pairing again.
func (s *DeviceService) register(ctx context.Context, session *pairingSession, name, confirmedBy string) (*models.Device, *authentication.TokenPair, error) {
	previous := ""
	device, err := s.repo.Get(ctx, session.deviceID)
	switch {
	case err == nil:
		previous = device.JTI // A device pairing again keeps its target
	case persistence.IsNotExist(err):
		device = &models.Device{}
		device.ID = session.deviceID
	default:
		return nil, nil, fmt.Errorf("getting device: %w", err)
	}

	tokens, err := s.jwts.CreateDeviceSession(ctx, session.deviceID, name)
	if err != nil {
		return nil, nil, fmt.Errorf("creating device session: %w", err)
	}

	now := time.Now()
	device.Name = name
	device.Hardware = string(session.hardware)
	device.JTI = tokens.JTI
	device.PairedBy = confirmedBy
	device.PairedAt = now
	device.LastSeenAt = now
	if err := s.repo.Save(ctx, device); err != nil {
		return nil, nil, fmt.Errorf("saving device: %w", err)
	}

	if previous != "" {
		if _, err := s.jtis.Revoke(ctx, previous); err != nil {
			return nil, nil, fmt.Errorf("revoking previous device session: %w", err)
		}
	}

	return device, tokens, nil
}

// deliver hands the tokens to the stream of the session, returns false when that stream is gone.
func (s *DeviceService) deliver(session *pairingSession, tokens *authentication.TokenPair) bool {
	s.lock.Lock()
	defer s.lock.Unlock()

	if session.closed {
		return false
	}
	session.paired <- tokens // Buffered, and a session is only claimed once as claiming removes its code

	return true
}

// Refresh exchanges a device refresh token for new device tokens, and counts as the device being used.
func (s *DeviceService) Refresh(ctx context.Context, token string) (*authentication.TokenPair, error) {
	t, err := s.jwts.Validate(ctx, token)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrDeviceTokenInvalid, err)
	}
	claims, ok := authentication.GetClaims(t.Claims)
	if !ok || claims.Scope != authentication.ScopeDeviceRefresh {
		return nil, ErrDeviceTokenInvalid
	}

	device, err := s.repo.Get(ctx, claims.User.ID)
	if err != nil {
		if persistence.IsNotExist(err) {
			return nil, ErrDeviceTokenInvalid
		}

		return nil, fmt.Errorf("getting device: %w", err)
	}
	if device.JTI != claims.ID {
		return nil, ErrDeviceTokenInvalid
	}

	tokens, err := s.jwts.RefreshDeviceSession(ctx, device.ID, device.Name, device.JTI)
	if err != nil {
		return nil, fmt.Errorf("refreshing device session: %w", err)
	}
	if err := s.repo.SetLastSeen(ctx, device.ID, time.Now()); err != nil {
		return nil, fmt.Errorf("updating device: %w", err)
	}

	return tokens, nil
}

// List returns the paired devices, followed by the devices connected without being paired (through a logged in user).
func (s *DeviceService) List(ctx context.Context) ([]*DeviceInfo, error) {
	paired, err := s.paired(ctx)
	if err != nil {
		return nil, err
	}

	connected := map[string]devices.HardwareType{}
	for _, conn := range s.screens.Connections() {
		if conn.DeviceID != "" {
			connected[conn.DeviceID] = conn.Hardware
		}
	}

	result := make([]*DeviceInfo, 0, len(paired)+len(connected))
	for _, device := range paired {
		_, isConnected := connected[device.ID]
		delete(connected, device.ID)
		result = append(result, &DeviceInfo{
			ID:         device.ID,
			Name:       device.Name,
			Hardware:   device.Hardware,
			Target:     device.Target,
			Paired:     true,
			Connected:  isConnected,
			PairedAt:   device.PairedAt,
			LastSeenAt: device.LastSeenAt,
		})
	}

	for _, id := range slices.Sorted(maps.Keys(connected)) {
		result = append(result, &DeviceInfo{ID: id, Hardware: string(connected[id]), Connected: true})
	}

	return result, nil
}

// Revoke un-pairs the device: its tokens stop working and its connections close right away, and it has to pair again.
func (s *DeviceService) Revoke(ctx context.Context, id string) error {
	device, err := s.repo.Get(ctx, id)
	if err != nil {
		if persistence.IsNotExist(err) {
			return devices.ErrDeviceNotFound
		}

		return fmt.Errorf("getting device: %w", err)
	}

	if _, err := s.jtis.Revoke(ctx, device.JTI); err != nil {
		return fmt.Errorf("revoking device session: %w", err)
	}
	if err := s.repo.Delete(ctx, id); err != nil {
		return fmt.Errorf("deleting device: %w", err)
	}

	// Don't wait for the periodic session check of the connection, the screen goes back to pairing right away
	s.screens.CloseDevice(id, "device revoked")

	return nil
}

// Bind binds the paired device to a target of the room, an empty target removes the binding.
// It takes effect the next time the scene instructs its devices.
func (s *DeviceService) Bind(ctx context.Context, id, target string) (*models.Device, error) {
	target = strings.TrimSpace(target)
	if target != "" && !devices.ValidID(target) {
		return nil, ErrDeviceTargetInvalid
	}

	ok, err := s.repo.SetTarget(ctx, id, target)
	if err != nil {
		return nil, fmt.Errorf("binding device: %w", err)
	}
	if !ok {
		return nil, devices.ErrDeviceNotFound
	}

	return s.repo.Get(ctx, id)
}

// Room wraps the store of the authored room, so the room also binds the paired devices to their targets.
// The authored room itself stays read only.
func (s *DeviceService) Room(authored RoomConfigStore) RoomConfigStore {
	return &pairedRoom{authored: authored, devices: s}
}

// CurrentRoom implements RoomConfigStore.
func (r *pairedRoom) CurrentRoom(ctx context.Context) (*devices.Room, error) {
	authored, err := r.authored.CurrentRoom(ctx)
	if err != nil {
		return nil, err
	}

	paired, err := r.devices.paired(ctx)
	if err != nil {
		return nil, err
	}

	room := devices.NewRoom(authored.ID, authored.Name)
	maps.Copy(room.Devices, authored.Devices)
	for target, ids := range authored.Bindings {
		room.Bindings[target] = slices.Clone(ids)
	}

	for _, device := range paired {
		if device.Target == "" {
			continue
		}
		if _, known := room.Devices[device.ID]; !known {
			room.AddDevice(&devices.Device{
				ID:          device.ID,
				Name:        device.Name,
				Kind:        devices.KindScreen,
				Integration: devices.IntegrationScreen,
			})
		}
		if err := room.Bind(device.Target, device.ID); err != nil {
			return nil, fmt.Errorf("binding device %q: %w", device.ID, err)
		}
	}

	return room, nil
}

// paired returns the devices that are still paired: one that wasn't used for the lifetime of its refresh token has to pair again.
func (s *DeviceService) paired(ctx context.Context) ([]*models.Device, error) {
	result, err := s.repo.ListSeenSince(ctx, s.seenCutoff())
	if err != nil {
		return nil, fmt.Errorf("listing devices: %w", err)
	}

	return result, nil
}

func (s *DeviceService) seenCutoff() time.Time {
	return time.Now().Add(-s.jwts.Options().DeviceRefreshTokenDuration)
}
