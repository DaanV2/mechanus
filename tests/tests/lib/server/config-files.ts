import { execFileSync } from "node:child_process";

/** The container the test server runs in, see the docker-compose files in this folder. */
const CONTAINER = process.env.MECHANUS_CONTAINER ?? "mechanus-server";
/** The server's app config dir inside the container. */
const CONFIG_DIR = "/app/.config";

/**
 * Whether config files can be written for the test server: it must run in the docker container of the
 * compose files, with a writable scenes folder.
 */
export function canWriteConfigFiles(): boolean {
  try {
    execFileSync(
      "docker",
      ["exec", CONTAINER, "test", "-w", `${CONFIG_DIR}/scenes`],
      { stdio: "ignore" },
    );

    return true;
  } catch {
    return false;
  }
}

/**
 * Writes a config file in the server's config dir, as a GM editing it would. The file is written inside the
 * container, so the server sees the change on every platform. `file` is relative to the config dir.
 */
export function writeConfigFile(file: string, content: string): void {
  execFileSync(
    "docker",
    ["exec", "-i", CONTAINER, "sh", "-c", `cat > "${CONFIG_DIR}/${file}"`],
    {
      input: content,
    },
  );
}

/** Removes a config file from the server's config dir, if it exists. */
export function removeConfigFile(file: string): void {
  execFileSync("docker", [
    "exec",
    CONTAINER,
    "rm",
    "-f",
    `${CONFIG_DIR}/${file}`,
  ]);
}
