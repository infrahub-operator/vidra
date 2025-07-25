---
title: Installing the Vidra CLI
sidebar_position: 1
---
import Admonition from '@theme/Admonition';

# Installing the Vidra CLI

You can install the Vidra CLI tool directly using Go or by downloading a pre-built binary.

## Prerequisites

- [Go](https://golang.org/dl/) 1.20 or newer

## Installation

### Option 1: Install via Go

You can install Vidra directly using the `go install` command:

```sh
go install github.com/infrahub-operator/vidra/vidra-cli@latest
```

This will download and build the latest version, placing the `vidra` binary in your `$GOPATH/bin` or `$HOME/go/bin` directory. Make sure this directory is in your `PATH`.

### Option 2: Download Pre-built Binary

Check the [Vidra releases page](https://github.com/your-org/vidra/releases) for pre-built binaries for your platform. Download the appropriate binary and move it to a directory in your `PATH`:

```sh
# Example for Linux/macOS
curl -LO https://github.com/infrahub-operator/vidra/releases/download/vX.Y.Z/vidra-cli
chmod +x vidra-cli
sudo mv vidra-cli /usr/local/bin/
```

Replace `vX.Y.Z` with the latest release version.

## Verify Installation

Check that Vidra is installed correctly:

```sh
vidra-cli --help
```

You should see the help output for the Vidra CLI.

## Generate ZSH Completion Script for Vidra CLI

```bash
vidra-cli completion zsh > _vidra-cli

mkdir -p ~/.zsh/completions
mv _vidra-cli ~/.zsh/completions/

fpath+=~/.zsh/completions
autoload -Uz compinit
compinit

source ~/.zshrc
```

## Upgrading

To upgrade, run the `go install` command again with the latest version, or download the latest release.

---

For more usage instructions, see the [Vidra CLI documentation](./usage).
