---
layout: default
title: Installation
---

# Installation

Getting started with goenums is simple. Since it's a code generation tool that runs during development rather than a runtime dependency, you just need to install the CLI tool.

## Via Go Install

The recommended way to install goenums is via Go's built-in package manager:

```bash
go install github.com/zarldev/goenums@latest
```

This will download, compile, and install the latest version of goenums, making it available in your system's Go binary directory (`$GOPATH/bin` or `$GOBIN`).

## Verify Installation

To verify that goenums was installed correctly, run:

```bash
goenums -v
```

You should see the goenums logo and current version displayed:

```bash
   ____ _____  ___  ____  __  ______ ___  _____
  / __ '/ __ \/ _ \/ __ \/ / / / __ '__ \/ ___/
 / /_/ / /_/ /  __/ / / / /_/ / / / / / (__  ) 
 \__, /\____/\___/_/ /_/\__,_/_/ /_/ /_/____/  
/____/
        version: v0.3.6
```