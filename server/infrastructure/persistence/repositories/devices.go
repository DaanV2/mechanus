package repositories

import (
	"context"
	"time"

	"github.com/DaanV2/mechanus/server/infrastructure/logging"
	"github.com/DaanV2/mechanus/server/infrastructure/persistence"
	"github.com/DaanV2/mechanus/server/infrastructure/persistence/models"
)

// DeviceRepository provides database operations for paired devices.
type DeviceRepository struct {
	db     *persistence.DB
	logger logging.Enriched
}

// NewDeviceRepository creates a new device repository with the provided database.
func NewDeviceRepository(db *persistence.DB) *DeviceRepository {
	return &DeviceRepository{
		db:     db,
		logger: logging.Enriched{}.WithPrefix("devices"),
	}
}

// Save creates the device, or updates it when a device with its ID exists.
func (repo *DeviceRepository) Save(ctx context.Context, device *models.Device) error {
	repo.logger.With("deviceId", device.ID).From(ctx).Debug("saving device")

	return repo.db.WithContext(ctx).Save(device).Error
}

// Get looks up the device by the given id.
func (repo *DeviceRepository) Get(ctx context.Context, id string) (*models.Device, error) {
	repo.logger.With("deviceId", id).From(ctx).Debug("getting device")

	var device models.Device
	tx := repo.db.WithContext(ctx).First(&device, "id = ?", id)
	if tx.Error != nil {
		return nil, tx.Error
	}

	return &device, nil
}

// ListSeenSince returns the devices that paired or refreshed their tokens after the given time, oldest pairing first.
func (repo *DeviceRepository) ListSeenSince(ctx context.Context, since time.Time) ([]*models.Device, error) {
	repo.logger.From(ctx).Debug("listing devices")

	var devices []*models.Device
	tx := repo.db.WithContext(ctx).
		Where("last_seen_at > ?", since).
		Order("paired_at").
		Find(&devices)

	return devices, tx.Error
}

// SetLastSeen records that the device was seen at the given time.
func (repo *DeviceRepository) SetLastSeen(ctx context.Context, id string, at time.Time) error {
	return repo.db.WithContext(ctx).
		Model(&models.Device{}).
		Where("id = ?", id).
		Update("last_seen_at", at).Error
}

// SetTarget binds the device to the target, returns false when there is no such device.
func (repo *DeviceRepository) SetTarget(ctx context.Context, id, target string) (bool, error) {
	repo.logger.With("deviceId", id, "target", target).From(ctx).Debug("binding device")

	tx := repo.db.WithContext(ctx).
		Model(&models.Device{}).
		Where("id = ?", id).
		Update("target", target)

	return tx.RowsAffected == 1, tx.Error
}

// Delete permanently removes the device, so it can pair again with the same ID.
func (repo *DeviceRepository) Delete(ctx context.Context, id string) error {
	repo.logger.With("deviceId", id).From(ctx).Debug("deleting device")

	return repo.db.WithContext(ctx).Unscoped().Delete(&models.Device{}, "id = ?", id).Error
}

// DeleteNotSeenSince permanently removes the devices last seen before the given time.
func (repo *DeviceRepository) DeleteNotSeenSince(ctx context.Context, since time.Time) error {
	return repo.db.WithContext(ctx).Unscoped().Delete(&models.Device{}, "last_seen_at <= ?", since).Error
}
