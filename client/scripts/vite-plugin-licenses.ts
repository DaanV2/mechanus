// Vite plugin that writes the licences of every npm package bundled into the client to `licenses/npm.json`.
// Only packages whose code ends up in the browser bundle are listed, not build tools.
// The About page reads this file next to `licenses/go.json`, which the server build writes.
import { readdirSync, readFileSync } from 'node:fs';
import { join } from 'node:path';
import type { Plugin } from 'vite';

/** One third-party dependency; same shape as the Go list from `server/scripts/licenses`. */
interface LicenseEntry {
  name: string;
  version: string;
  license: string;
  url?: string;
  text?: string;
}

interface PackageJson {
  name?: string;
  version?: string;
  license?: string | { type?: string };
  homepage?: string;
  repository?: string | { url?: string };
}

const NODE_MODULES = '/node_modules/';

/** Returns the root folder of the npm package that a module ID lives in, or undefined for project files. */
function packageDir(id: string): string | undefined {
  const path = id.replace(/^\0/, '').split('?')[0].replaceAll('\\', '/');
  const index = path.lastIndexOf(NODE_MODULES);
  if (index < 0) return undefined;

  const rest = path.slice(index + NODE_MODULES.length).split('/');
  const nameParts = rest[0].startsWith('@') ? 2 : 1;
  if (rest.length <= nameParts) return undefined;

  return path.slice(0, index + NODE_MODULES.length) + rest.slice(0, nameParts).join('/');
}

function readEntry(dir: string): LicenseEntry | undefined {
  let pkg: PackageJson;
  try {
    pkg = JSON.parse(readFileSync(join(dir, 'package.json'), 'utf8')) as PackageJson;
  } catch {
    return undefined;
  }
  if (!pkg.name) return undefined;

  const license = typeof pkg.license === 'string' ? pkg.license : pkg.license?.type;
  const repository = typeof pkg.repository === 'string' ? pkg.repository : pkg.repository?.url;

  const texts: string[] = [];
  for (const file of readdirSync(dir).sort()) {
    if (!/^(licen[cs]e|copying|notice)/i.test(file)) continue;
    try {
      texts.push(readFileSync(join(dir, file), 'utf8').trim());
    } catch {
      // a folder or unreadable file, skip it
    }
  }

  return {
    name: pkg.name,
    version: pkg.version ?? '',
    license: license ?? 'Unknown',
    url: pkg.homepage ?? repository?.replace(/^git\+/, '').replace(/\.git$/, ''),
    text: texts.length > 0 ? texts.join('\n\n') : undefined
  };
}

export function licenses(): Plugin {
  return {
    name: 'mechanus-licenses',
    apply: 'build',
    generateBundle() {
      // SvelteKit also builds a server bundle for prerendering; only the browser bundle matters here.
      if (this.environment.config.consumer !== 'client') return;

      const entries = new Map<string, LicenseEntry>();
      for (const id of this.getModuleIds()) {
        const dir = packageDir(id);
        if (!dir || entries.has(dir)) continue;

        const entry = readEntry(dir);
        if (entry) entries.set(dir, entry);
      }

      // The same package can be installed at two paths (nested node_modules); keep one per name@version.
      const unique = new Map<string, LicenseEntry>();
      for (const entry of entries.values()) unique.set(`${entry.name}@${entry.version}`, entry);

      const sorted = [...unique.values()].sort((a, b) => a.name.localeCompare(b.name));
      this.emitFile({
        type: 'asset',
        fileName: 'licenses/npm.json',
        source: JSON.stringify(sorted, null, 2)
      });
    }
  };
}
