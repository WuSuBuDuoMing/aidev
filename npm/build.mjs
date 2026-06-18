#!/usr/bin/env node
/**
 * npm build script — cross-compiles NeoCode CLI for all platforms
 * and creates platform-specific npm packages.
 */

import { execSync } from 'child_process';
import { mkdirSync, writeFileSync, existsSync, copyFileSync } from 'fs';
import { join } from 'path';

const VERSION = process.env.VERSION || '2.3.0';
const ROOT = process.cwd();

const targets = [
  { os: 'darwin',  arch: 'arm64',  pkg: '@neocode/cli-darwin-arm64'  },
  { os: 'darwin',  arch: 'amd64',  pkg: '@neocode/cli-darwin-x64'    },
  { os: 'linux',   arch: 'arm64',  pkg: '@neocode/cli-linux-arm64'   },
  { os: 'linux',   arch: 'amd64',  pkg: '@neocode/cli-linux-x64'     },
  { os: 'windows', arch: 'arm64',  pkg: '@neocode/cli-win32-arm64'   },
  { os: 'windows', arch: 'amd64',  pkg: '@neocode/cli-win32-x64'     },
];

console.log(`Building NeoCode ${VERSION} for ${targets.length} targets...`);

for (const target of targets) {
  const ext = target.os === 'windows' ? '.exe' : '';
  const binName = `neocode${ext}`;
  const pkgDir = join(ROOT, 'platforms', target.pkg.replace('/', '-'));
  const binDir = join(pkgDir, 'bin');

  mkdirSync(binDir, { recursive: true });

  console.log(`  Building ${target.os}/${target.arch}...`);

  execSync(
    `CGO_ENABLED=0 GOOS=${target.os} GOARCH=${target.arch} go build -ldflags "-s -w -X main.Version=${VERSION}" -o "${join(binDir, binName)}" ../cmd/neocode`,
    { cwd: ROOT, stdio: 'pipe' }
  );

  writeFileSync(join(pkgDir, 'package.json'), JSON.stringify({
    name: target.pkg,
    version: VERSION,
    description: `NeoCode CLI binary for ${target.os}-${target.arch}`,
    os: [target.os === 'windows' ? 'win32' : target.os],
    cpu: [target.arch === 'amd64' ? 'x64' : target.arch],
    files: ['bin/'],
    license: 'MIT',
  }, null, 2));
}

// Create meta-package
const metaPkgDir = join(ROOT, 'neocode');
mkdirSync(metaPkgDir, { recursive: true });
mkdirSync(join(metaPkgDir, 'bin'), { recursive: true });

writeFileSync(join(metaPkgDir, 'bin', 'neocode.js'), `#!/usr/bin/env node
const { execFileSync } = require('child_process');
const { join } = require('path');
const os = require('os');

const platform = os.platform();
const arch = os.arch();

const mapping = {
  'darwin-arm64':  '@neocode/cli-darwin-arm64',
  'darwin-x64':    '@neocode/cli-darwin-x64',
  'linux-arm64':   '@neocode/cli-linux-arm64',
  'linux-x64':     '@neocode/cli-linux-x64',
  'win32-arm64':   '@neocode/cli-win32-arm64',
  'win32-x64':     '@neocode/cli-win32-x64',
};

const key = platform + '-' + arch;
const pkg = mapping[key];
if (!pkg) {
  console.error('Unsupported platform: ' + key);
  process.exit(1);
}

const ext = platform === 'win32' ? '.exe' : '';
let binPath;
try {
  binPath = require.resolve(pkg + '/bin/neocode' + ext);
} catch {
  console.error('Package not found: ' + pkg);
  console.error('Install it with: npm install ' + pkg);
  process.exit(1);
}

try {
  execFileSync(binPath, process.argv.slice(2), { stdio: 'inherit' });
} catch (e) {
  process.exit(e.status || 1);
}
`);

const optionalDeps = {};
for (const target of targets) {
  optionalDeps[target.pkg] = VERSION;
}

writeFileSync(join(metaPkgDir, 'package.json'), JSON.stringify({
  name: '@neocode/cli',
  version: VERSION,
  description: 'NeoCode — AI coding agent for Chinese and international models',
  bin: { neocode: 'bin/neocode.js' },
  optionalDependencies: optionalDeps,
  keywords: ['ai', 'coding', 'agent', 'deepseek', 'claude', 'gpt', 'gemini', 'mimo'],
  author: 'NeoCode',
  license: 'MIT',
  homepage: 'https://neocode.dev',
  repository: 'github:user/neocode',
}, null, 2));

console.log('\\nDone! Meta-package created at npm/neocode/');
