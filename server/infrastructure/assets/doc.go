// Package assets resolves the asset references in scene configuration to URLs screens can load, and
// serves those assets over HTTP.
//
// Assets live in the app config folder:
//   - scene assets in the scene's own folder, <app config dir>/scenes/<scene id>/..., referenced by a
//     path relative to that folder (maps/cave.webp);
//   - shared assets in <app config dir>/shared/..., referenced with the shared:/ prefix
//     (shared:/monsters/goblins/goblin.png).
//
// Resolved URLs carry a content hash (?v=<hash>), so screens can cache them for as long as the file
// does not change.
package assets
