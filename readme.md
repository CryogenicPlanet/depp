# Depp - A fast unused and duplicate package checker

![](https://user-images.githubusercontent.com/10355479/139758905-7f911615-84d0-46c6-805a-06f8eafaf633.png)

## Installation

```
npm install -g depp-installer (will try to get npm install -g depp later)
```

## Usage

Just run `depp` in your project folder and it will do the rest. Keep in mind it will likely fail without setting some externals

**Note if you want it to work with JS** please use `-j` or `--js` by default it will do only `.ts|.tsx` files

All options
```
➜ depp --help
NAME:
   depp - Find un used packages fast

USAGE:
   depp [global options] command [command options] [arguments...]

COMMANDS:
   clean    Cleans all output files
   show     Shows previous report
   help, h  Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --dev, -d                             Enable dev dependencies (default: false)
   --js, -j                              Enable js source files (default: false)
   --path value, -p value                Overwrite root directory
   --log, -l                             Will write logs to .depcheck.log (default: false)
   --source value, -s value              Overwrite default sources
   --report, -r                          Generate report file (default: false)
   --show-versions, -v                   Show conflicting versions (default: false)
   --write-output-files, -w              This will write the esbuild output files. (default: false)
   --externals value, -e value           Pass custom externals using this flag
   --ignore-namespace value, --in value  Pass namespace (@monorepo) to be ignored
   --help, -h                            show help (default: false)

```

### Example Advanced usage

This is an example of advanced usage of the script with `externals` and `ignore-namespace`

```
depp -v -j -e mobx -e magic-sdk -e domain -e @daybrush/utils -e yjs -e constants -e ws  -e perf_hooks -in @editor -in @server   --report
```


## Why use this

1. It is using `esbuild` and `go` so it is quite a bit faster than most other tools
2. Most tools that I could find at least, didn't not support monorepos. This does and is built for monorepos

## Caveats 

This is not been extensively tested and might have some short comings, it may not identify every unused package but will definitely do a decent first pass


## Acknowledgement


This is built upon the excellent work down by [@evanw](https://github.com/evanw/) on `esbuild` and uses `esbuild` under the hood