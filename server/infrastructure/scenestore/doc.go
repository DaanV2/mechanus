// Package scenestore implements the scene storage ports of the application layer.
//
// Configuration and state are stored separately:
//   - [ConfigStore] reads authored scene and room configuration from file storage
//     (<app config dir>/scenes/<scene id>.dat and <app config dir>/rooms/<room id>.dat, JSON).
//   - [StateStore] keeps the runtime state of each scene in the database (key-value table,
//     keys state/scene/<scene id>).
package scenestore
