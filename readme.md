# Depp - A fast unused and duplicate package checker [![Go Reference](https://pkg.go.dev/badge/github.com/CryogenicPlanet/depp.svg)](https://pkg.go.dev/github.com/CryogenicPlanet/depp)

![](https://user-images.githubusercontent.com/10355479/139758905-7f911615-84d0-46c6-805a-06f8eafaf633.png)

## Installation

```bash
## NPM
npm install -g depp-installer 
# (will try to get npm install -g depp later)

## Go
go install github.com/cryogenicplanet/depp@latest

```

## Usage

Just run `depp` in your project folder and it will do the rest. Keep in mind it will likely fail without setting some externals

**Note if you want it to work with JS** please use `-j` or `--js` by default it will do only `.ts|.tsx` files

All options
```bash
➜ depp --help  
NAME:
   depp - Find un used packages fast

USAGE:
   depp [global options] command [command options] [arguments...]

COMMANDS:
   clean      Cleans all output files
   show       Shows previous report
   deploy, d  Automatically deploy your report to netlify
   config     A command to handle config
   init       Initialize project
   help, h    Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --dev, -d                              Enable dev dependencies (default: false)
   --js, -j                               Enable js source files (default: false)
   --path value, -p value                 Overwrite root directory
   --log, -l                              Will write logs to .depcheck.log (default: false)
   --source value, -s value               Overwrite default sources
   --report, -r                           Generate report file (default: false)
   --show-versions, -v                    Show conflicting versions (default: false)
   --write-output-files, -w               This will write the esbuild output files. (default: false)
   --externals value, -e value            Pass custom externals using this flag
   --ignore-namespaces value, --in value  Pass namespace (@monorepo) to be ignored
   --no-open, --no                        Flag to prevent auto opening report in browser (default: false)
   --save-config, --sc                    Flag to automatically save config from other flags (default: false)
   --ci                                   Run in github actions ci mode (default: false)
   --deploy value                         Will automatically deploy report to netlify
   --help, -h                             show help (default: false)
```

## Example Advanced usage

This is an example of advanced usage of the script with `externals` and `ignore-namespace`

```
depp -v -j -e mobx -e magic-sdk -e domain -e @daybrush/utils -e yjs -e constants -e ws  -e perf_hooks -in @editor -in @server   --report
```


## Configuration

You can save your `depp` config and not have to run it with flags every time, the config is saved in `.depp/config.json` but can be created from the cli

```bash
# Initialize config
depp init 

➜ depp --help                                                
NAME:
   depp - Find un used packages fast

USAGE:
   depp [global options] command [command options] [arguments...]

COMMANDS:
   clean      Cleans all output files
   show       Shows previous report
   deploy, d  Automatically deploy your report to netlify
   config     A command to handle config
   init       Initialize project
   help, h    Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --dev, -d                              Enable dev dependencies (default: false)
   --js, -j                               Enable js source files (default: false)
   --path value, -p value                 Overwrite root directory
   --log, -l                              Will write logs to .depcheck.log (default: false)
   --source value, -s value               Overwrite default sources
   --report, -r                           Generate report file (default: false)
   --show-versions, -v                    Show conflicting versions (default: false)
   --write-output-files, -w               This will write the esbuild output files. (default: false)
   --externals value, -e value            Pass custom externals using this flag
   --ignore-namespaces value, --in value  Pass namespace (@monorepo) to be ignored
   --no-open, --no                        Flag to prevent auto opening report in browser (default: false)
   --save-config, --sc                    Flag to automatically save config from other flags (default: false)
   --ci                                   Run in github actions ci mode (default: false)
   --deploy value                         Will automatically deploy report to netlify
   --browser                              Will use esbuild browser platform (by default it uses node platform) (default: false)
   --ignore-path value, --ip value        A glob pattern of files to be ignored
```

## CI

Currently only supports Github actions out of the box.

In mode, `depp` will automatically comment on the PR with its report. It will look like [this](https://github.com/CryogenicPlanet/cryogenicplanet.github.io/issues/49#issuecomment-961496544)

It can also deploy the report to [netlify](netlify.com) but requires a `NETLIFY_TOKEN` which you can get [here](https://app.netlify.com/user/applications#personal-access-tokens)

```yaml
name: Dependency CI

on:
  pull_request:


jobs:
  release-go:
    runs-on: ubuntu-latest
    steps:
      - name: Checkout
        uses: actions/checkout@v2
        with:
          fetch-depth: 0
      - name: Set up Go
        uses: actions/setup-go@v2
        with:
          go-version: 1.17
      - name: Install Depp 
        run: go install github.com/cryogenicplanet/depp@latest
      - name: Run Depp
        run: depp --ci
        env:
            # NETLIFY_TOKEN: ${{secrets.NETLIFY_TOKEN}}
            # Optional if you want report urls or not
            # You can get a netlify pat here https://app.netlify.com/user/applications#personal-access-tokens
```

## Example Outputs

1. [Markdown](./static/markdownReport.md)
2. [Html](https://cryogenicplanet.github.io/depp/static/htmlReport.html)

## Why use this

1. It is using `esbuild` and `go` so it is quite a bit faster than most other tools
2. Most tools that I could find at least, didn't not support monorepos. This does and is built for monorepos

## Caveats 

This is not been extensively tested and might have some short comings, it may not identify every unused package but will definitely do a decent first pass


## Acknowledgement


> Credits to [@zack_overflow](https://github.com/zackradisic) for the amazing cover photo

This is built upon the excellent work down by [@evanw](https://github.com/evanw/) on `esbuild` and uses `esbuild` under the hood
