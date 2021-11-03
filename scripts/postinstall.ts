import { exec } from "child_process";

import * as fs from "fs";

import path from "path";

import mkdirp from "mkdirp";

import axios from "axios";

import decompress from "decompress";
// Mapping from Node's `process.arch` to Golang's `$GOARCH`
const ARCH_MAPPING: { [name: string]: string } = {
  ia32: "386",
  x64: "x86_64",
  arm: "arm",
};


// Mapping between Node's `process.platform` to Golang's
const PLATFORM_MAPPING: { [name: string]: string } = {
  darwin: "darwin",
  linux: "linux",
  win32: "windows",
  freebsd: "freebsd",
};

async function getInstallationPath() {
  // `npm bin` will output the path where binary files should be installed

  const value = await execShellCommand("npm bin -g");

  let dir: string | undefined = undefined;
  if (!value || value.length === 0) {
    // We couldn't infer path from `npm bin`. Let's try to get it from
    // Environment letiables set by NPM when it runs.
    // npm_config_prefix points to NPM's installation directory where `bin` folder is available
    // Ex: /Users/foo/.nvm/versions/node/v4.3.0
    let env = process.env;
    if (env && env.npm_config_prefix) {
      dir = path.join(env.npm_config_prefix, "bin");
    }
  } else {
    dir = value.trim();
  }
  if (dir) {
    await mkdirp(dir);
    return dir;
  } else {
    throw new Error("Could not getInstallationPath");
  }
}

async function verifyAndPlaceBinary(binName: string, binPath: string, callback: ErrCallback) {
  if (!fs.existsSync(path.join(binPath, binName)))
    return callback(
      new Error("Downloaded binary does not contain the binary specified in configuration - " +
        binName)
    );

  // Get installation path for executables under node
  const installationPath = await getInstallationPath();
  // Copy the executable to the path
  fs.rename(
    path.join(binPath, binName),
    path.join(installationPath, binName),
    (err) => {
      if (!err) {
        console.info("Installed cli successfully");
        callback(null);
      } else {
        callback(err);
      }
    }
  );
}

function validateConfiguration(packageJson) {
  if (!packageJson.version) {
    return "'version' property must be specified";
  }

  if (!packageJson.goBinary || typeof packageJson.goBinary !== "object") {
    return "'goBinary' property must be defined and be an object";
  }

  if (!packageJson.goBinary.name) {
    return "'name' property is necessary";
  }

  if (!packageJson.goBinary.path) {
    return "'path' property is necessary";
  }
}

function parsePackageJson() {
  if (!(process.arch in ARCH_MAPPING)) {
    throw new Error(
      "Installation is not supported for this architecture: " + process.arch
    );
  }

  if (!(process.platform in PLATFORM_MAPPING)) {
    throw new Error(
      "Installation is not supported for this platform: " + process.platform
    );
  }

  let packageJsonPath = path.join(".", "package.json");
  if (!fs.existsSync(packageJsonPath)) {
    throw new Error(
      "Unable to find package.json. " +
        "Please run this script at root of the package you want to be installed"
    );
  }

  let packageJson = JSON.parse(fs.readFileSync(packageJsonPath, "utf-8"));
  let error = validateConfiguration(packageJson);
  if (error && error.length > 0) {
    throw new Error("Invalid package.json: " + error);
  }

  // We have validated the config. It exists in all its glory
  let binName: string = packageJson.goBinary.name;
  let binPath: string = packageJson.goBinary.path;
  let version: string = packageJson.version;
  if (version[0] === "v") version = version.substr(1); // strip the 'v' if necessary v0.0.1 => 0.0.1

  // Binary name on Windows has .exe suffix
  if (process.platform === "win32") {
    binName += ".exe";
  }

  return {
    binName: binName,
    binPath: binPath,
    version: version,
  };
}

type ErrCallback = (_err: Error | null) => void;

async function downloadAndInstall(url: string, outPath: string, binName:string) {
  const tempPath = fs.mkdtempSync("depp");

  const { data } = await axios({
    method: "get",
    url: url,
    responseType: "stream",
  });

  const tarballPath = `${tempPath}/tarball.tar.gz`;

  const outFile = fs.createWriteStream(tarballPath);

  data.pipe(outFile);

  return new Promise<void>((resolve) => {
    outFile.on("unpipe", async () => {
      console.log("Unpacked tarball");
      await decompress(tarballPath, `${tempPath}/tar`);
      fs.renameSync(
        path.resolve(`${tempPath}/tar/${binName}`),
        path.resolve(outPath)
      );
      console.log("Wrote binary to out path", outPath);

      fs.rmdirSync(tempPath, { recursive: true });

      resolve();
    });
  });
}

/**
 * Reads the configuration from application's package.json,
 * validates properties, copied the binary from the package and stores at
 * ./bin in the package's root. NPM already has support to install binary files
 * specific locations when invoked with "npm install -g"
 *
 *  See: https://docs.npmjs.com/files/package.json#bin
 */
const INVALID_INPUT = "Invalid inputs";
async function install(callback: ErrCallback) {
  let opts = parsePackageJson();
  if (!opts) return callback(new Error(INVALID_INPUT));
  mkdirp.sync(opts.binPath);
  console.info(
    `Copying the relevant binary for your platform ${process.platform}`
  );
  const platform = PLATFORM_MAPPING[process.platform];
  const arch = ARCH_MAPPING[process.arch];

  const url = `https://github.com/CryogenicPlanet/depp/releases/download/v${opts.version}/depp_${opts.version}_${platform}_${arch}.tar.gz`;

  console.log("Downloading binary from", url);

  await downloadAndInstall(url, `${opts.binPath}/${opts.binName}`, opts.binName);

  //   await execShellCommand(`cp ${src} ${opts.binPath}/${opts.binName}`);
  //   await execShellCommand(`cp ${src} ${opts.binPath}/${opts.binName}`);
  await verifyAndPlaceBinary(opts.binName, opts.binPath, callback);
}

async function uninstall(callback: ErrCallback) {
  let opts = parsePackageJson();
  try {
    const installationPath = await getInstallationPath();
    fs.unlink(path.join(installationPath, opts.binName), (err) => {
      if (err) {
        return callback(err);
      }
    });
  } catch (ex) {
    // Ignore errors when deleting the file.
  }
  console.info("Uninstalled cli successfully");
  return callback(null);
}

// Parse command line arguments and call the right method
let actions = {
  install: install,
  uninstall: uninstall,
};
/**
 * Executes a shell command and return it as a Promise.
 * @param cmd {string}
 * @return {Promise<string>}
 */
function execShellCommand(cmd: string) {
  return new Promise<string>((resolve, reject) => {
    exec(cmd, (error, stdout, stderr) => {
      if (error) {
        console.warn(error);
        reject(error);
      }
      resolve(stdout ? stdout : stderr);
    });
  });
}

const argv = process.argv;
if (argv && argv.length > 2) {
  let cmd = process.argv[2];
  if (!actions[cmd]) {
    console.log(
      "Invalid command. `install` and `uninstall` are the only supported commands"
    );
    process.exit(1);
  }

  actions[cmd](function (err: Error | null) {
    if (err) {
      console.error(err);
      process.exit(1);
    } else {
      process.exit(0);
    }
  });
}
