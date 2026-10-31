// Package scenes contains the scene domain model.
//
// Configuration and runtime state are kept strictly apart:
//
//   - [Config] is what the GM prepared: the entities present at the start, the base [Instructions]
//     for every output, and the [Variant]s (night, rain, combat, ...). It is loaded from file
//     storage and never modified at runtime.
//   - [State] is what happened during play: the active variant, entities revealed, removed or
//     spawned, and live instruction overrides. It is saved to the database when a scene is loaded
//     or exited, and restored the next time the scene is loaded.
//   - [Scene] is a loaded scene, a Config together with its State, and computes the effective
//     instructions and entities from both.
//
// Instructions are addressed to logical targets, which a devices.Room resolves to physical
// devices. [NewTransition] compares two sets of instructions so only what changed is sent.
//
// See docs/architecture/scenes.md for the full model.
package scenes
