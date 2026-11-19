// Package scenestore implements the scene storage ports of the application layer.
//
// Configuration and state are stored separately:
//   - [ConfigStore] reads authored scene and room configuration from config files
//     (<app config dir>/scenes/<scene id>.yaml and <app config dir>/rooms/<room id>.yaml, YAML or JSON
//     with a version, see [configfile]), and reloads them when they change.
//   - [StateStore] keeps the runtime state of each scene in the database (key-value table,
//     keys state/scene/<scene id>).
package scenestore
