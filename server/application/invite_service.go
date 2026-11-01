package application

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/DaanV2/mechanus/server/engine/authentication/roles"
	"github.com/DaanV2/mechanus/server/infrastructure/persistence"
	"github.com/DaanV2/mechanus/server/infrastructure/persistence/models"
	"github.com/DaanV2/mechanus/server/infrastructure/persistence/repositories"
)

var (
	// ErrInviteInvalid is returned when an invite doesn't exist, expired, was used or was revoked.
	// The reasons are not told apart, so the answer doesn't help guessing invites.
	ErrInviteInvalid = errors.New("invite is invalid, expired or already used")
)

// MaxInviteDuration is the longest an invite can stay valid.
const MaxInviteDuration = 30 * 24 * time.Hour

// InviteService manages single use invites that let someone create their own account.
type InviteService struct {
	repo            *repositories.InviteRepository
	users           *UserService
	defaultDuration time.Duration
}

// NewInviteService creates a new InviteService, invites created without a duration stay valid for defaultDuration.
func NewInviteService(repo *repositories.InviteRepository, users *UserService, defaultDuration time.Duration) *InviteService {
	return &InviteService{repo, users, min(defaultDuration, MaxInviteDuration)}
}

// Create makes a new invite on behalf of createdBy, valid for the given duration.
// A duration of 0 or less uses the default, durations above [MaxInviteDuration] are capped.
func (s *InviteService) Create(ctx context.Context, createdBy string, validFor time.Duration) (*models.Invite, error) {
	if validFor <= 0 {
		validFor = s.defaultDuration
	}
	validFor = min(validFor, MaxInviteDuration)

	invite := &models.Invite{
		CreatedBy: createdBy,
		ExpiresAt: time.Now().Add(validFor),
	}
	if err := s.repo.Create(ctx, invite); err != nil {
		return nil, fmt.Errorf("creating invite: %w", err)
	}

	return invite, nil
}

// ListUsable returns the invites that can still be used.
func (s *InviteService) ListUsable(ctx context.Context) ([]*models.Invite, error) {
	return s.repo.ListUsable(ctx, time.Now())
}

// Get returns the invite when it can still be used, [ErrInviteInvalid] otherwise.
func (s *InviteService) Get(ctx context.Context, id string) (*models.Invite, error) {
	if id == "" {
		return nil, ErrInviteInvalid
	}

	invite, err := s.repo.Get(ctx, id)
	if err != nil {
		if persistence.IsNotExist(err) {
			return nil, ErrInviteInvalid
		}

		return nil, fmt.Errorf("getting invite: %w", err)
	}
	if !invite.Usable(time.Now()) {
		return nil, ErrInviteInvalid
	}

	return invite, nil
}

// Revoke makes the invite unusable, returns [ErrInviteInvalid] when there is no such invite.
func (s *InviteService) Revoke(ctx context.Context, id string) error {
	ok, err := s.repo.Revoke(ctx, id)
	if err != nil {
		return fmt.Errorf("revoking invite: %w", err)
	}
	if !ok {
		return ErrInviteInvalid
	}

	return nil
}

// Accept creates an account with the user role from the invite, and uses up the invite.
// The invite is claimed first, so two people can't use the same invite at once;
// when the account can't be created (eg: the username is taken) the invite can be used again.
func (s *InviteService) Accept(ctx context.Context, id, username string, password []byte) (*models.User, error) {
	if id == "" {
		return nil, ErrInviteInvalid
	}

	claimed, err := s.repo.Claim(ctx, id, time.Now())
	if err != nil {
		return nil, fmt.Errorf("claiming invite: %w", err)
	}
	if !claimed {
		return nil, ErrInviteInvalid
	}

	user := &models.User{
		Username:     username,
		PasswordHash: password,
		Roles:        []string{roles.User.String()},
		Campaigns:    []*models.Campaign{},
	}
	if err := s.users.Create(ctx, user); err != nil {
		if rerr := s.repo.Release(ctx, id); rerr != nil {
			err = errors.Join(err, fmt.Errorf("releasing invite: %w", rerr))
		}

		return nil, err
	}

	if err := s.repo.SetUsedBy(ctx, id, user.ID); err != nil {
		return user, fmt.Errorf("recording invite use: %w", err)
	}

	return user, nil
}
