import { existsSync, readFileSync, readdirSync } from "node:fs";
import path from "node:path";
import process from "node:process";
import { pathToFileURL } from "node:url";

const repoRoot = process.cwd();
const packageJsonPath = path.join(repoRoot, "package.json");
const packageJson = JSON.parse(readFileSync(packageJsonPath, "utf8"));
const packageName = packageJson.name;
const exportsMap = packageJson.exports ?? {};
const distRoot = path.join(repoRoot, "gen", "dist");

if (!existsSync(distRoot)) {
  throw new Error(`built package output missing at ${distRoot}; run npm run build first`);
}

const actualExports = Object.keys(exportsMap).filter((key) => key !== "./package.json");
const expectedExports = walkBuiltModules(distRoot)
  .map((relativePath) => `./${relativePath.replace(/\\/g, "/").replace(/\.js$/, "")}`)
  .sort();

const missingExports = expectedExports.filter((key) => !actualExports.includes(key));
const extraExports = actualExports.filter((key) => !expectedExports.includes(key));

if (missingExports.length || extraExports.length) {
  throw new Error(
    [
      "package.json exports drifted from generated dist modules",
      missingExports.length ? `missing exports: ${missingExports.join(", ")}` : null,
      extraExports.length ? `extra exports: ${extraExports.join(", ")}` : null,
    ]
      .filter(Boolean)
      .join("\n"),
  );
}

for (const exportKey of actualExports) {
  const exportValue = exportsMap[exportKey];
  if (typeof exportValue !== "object" || exportValue === null) {
    throw new Error(`export ${exportKey} must use the { types, import } object shape`);
  }

  for (const field of ["types", "import"]) {
    const target = exportValue[field];
    if (typeof target !== "string") {
      throw new Error(`export ${exportKey} is missing ${field}`);
    }

    const targetPath = path.join(repoRoot, target);
    if (!existsSync(targetPath)) {
      throw new Error(`export ${exportKey} points to missing ${field} target ${target}`);
    }
  }

  const importSpecifier = `${packageName}/${exportKey.slice(2)}`;
  const imported = await import(pathToFileURL(path.join(repoRoot, exportsMap[exportKey].import)).href);
  if (Object.keys(imported).length === 0) {
    throw new Error(`export ${importSpecifier} loaded but had no named exports`);
  }
}

console.log(`validated ${actualExports.length} package exports`);

function walkBuiltModules(rootDir) {
  const results = [];
  const stack = [""];

  while (stack.length) {
    const currentRelativeDir = stack.pop();
    const currentAbsoluteDir = path.join(rootDir, currentRelativeDir);

    for (const entry of readdirSync(currentAbsoluteDir, { withFileTypes: true })) {
      const relativePath = path.join(currentRelativeDir, entry.name);
      if (entry.isDirectory()) {
        stack.push(relativePath);
        continue;
      }
      if (entry.isFile() && relativePath.endsWith("_pb.js")) {
        results.push(relativePath);
      }
    }
  }

  return results;
}
