package models

import (
	"time"

	"gorm.io/gorm"
)

// Device is a paired device: a TV, tablet or other screen that got its credential by pairing.
// The ID is the stable ID the device generated itself. Revoking a device deletes it, pairing again starts over.
type Device struct {
	Model
	// The name the admin gave it when pairing
	Name string
	// What the device reports to run on, metadata only
	Hardware string
	// The target of the room it is bound to, if any
	Target string
	// The session (JTI) of its tokens, the old session is revoked when a device pairs again
	JTI string
	// The admin that confirmed the pairing
	PairedBy string
	PairedAt time.Time
	// When the device last paired or refreshed its tokens
	LastSeenAt time.Time `gorm:"index"`
}

// BeforeCreate replaces the hook of [Model], the ID of a device is not generated but reported by the device.
func (d *Device) BeforeCreate(_ *gorm.DB) error {
	return nil
}
