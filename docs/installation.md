---
layout: default
title: Installation
---

Getting started with goenums is simple. Since it's a code generation tool that runs during development rather than a runtime dependency, you just need to install the CLI tool.

## Via Go Install

The recommended way to install goenums is via Go's built-in package manager:

```bash
go install github.com/zarldev/goenums@latest
```

This will download, compile, and install the latest version of goenums, making it available in your system's Go binary directory (`$GOPATH/bin` or `$GOBIN`).

## Verify Installation

To verify that `goenums` was installed correctly, run:

```bash
$ goenums -v
```

You should see the `goenums` logo and current version displayed:

```bash
   ____ _____  ___  ____  __  ______ ___  _____
  / __ '/ __ \/ _ \/ __ \/ / / / __ '__ \/ ___/
 / /_/ / /_/ /  __/ / / / /_/ / / / / / (__  ) 
 \__, /\____/\___/_/ /_/\__,_/_/ /_/ /_/____/
/____/
                version: v0.3.6
```

# Prerequisites

 - Go 1.21+ for full functionality including iterator support
 - Go 1.18-1.21 use the -l flag to generate code without iterator support

# Zero Dependencies

goenums is completely dependency-free, using only the Go standard library. This ensures minimal bloat, maximum stability, and eliminates dependency-related security concerns.

Next Steps: Learn how to [use goenums in your project]({{ '/usage' | relative_url }}).