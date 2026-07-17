import { cpSync, mkdirSync } from "node:fs";
import { join, resolve } from "node:path";
import { execSync } from "node:child_process";
import { fileURLToPath } from "node:url";

const scriptDir = resolve(fileURLToPath(new URL(".", import.meta.url)));
const rootDir = resolve(scriptDir, "..", "..");
const frontendDir = resolve(rootDir, "frontend");
const backendDir = resolve(rootDir, "backend");
const resourcesBackendDir = resolve(frontendDir, "src-tauri", "resources", "backend");

function targetTriple(platform, arch) {
  if (platform === "darwin" && arch === "arm64") return "aarch64-apple-darwin";
  if (platform === "darwin" && arch === "x64") return "x86_64-apple-darwin";
  if (platform === "linux" && arch === "x64") return "x86_64-unknown-linux-gnu";
  if (platform === "linux" && arch === "arm64") return "aarch64-unknown-linux-gnu";
  if (platform === "win32" && arch === "x64") return "x86_64-pc-windows-msvc";
  if (platform === "win32" && arch === "arm64") return "aarch64-pc-windows-msvc";

  throw new Error(`Unsupported platform/arch: ${platform}/${arch}`);
}

const triple = targetTriple(process.platform, process.arch);
const binName = process.platform === "win32" ? "resume-generator.exe" : "resume-generator";
const outputBinary = join(resourcesBackendDir, binName);

mkdirSync(resourcesBackendDir, { recursive: true });

execSync(`go build -o "${outputBinary}" .`, {
  cwd: backendDir,
  stdio: "inherit"
});

cpSync(join(backendDir, "templates"), join(resourcesBackendDir, "templates"), {
  recursive: true,
  force: true
});
cpSync(join(backendDir, "resume.json"), join(resourcesBackendDir, "resume.json"), {
  force: true
});

console.log(`Built backend sidecar (${triple}) at ${outputBinary}`);
