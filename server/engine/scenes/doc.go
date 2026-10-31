// Package scenes contains the scene domain model.
//
// A [Scene] is a place or moment in the game: it records which entities are present and what every
// output in the room should be doing. A scene can have [Variant]s (night, rain, combat, ...) that
// override parts of its [Instructions]. Instructions are addressed to logical targets, which a
// devices.Room resolves to physical devices.
//
// Moving between scenes or variants is planned with [Plan], which returns a [Transition] holding
// only the instructions that changed, so devices are not re-instructed needlessly.
//
// See docs/architecture/scenes.md for the full model.
package scenes
