---
title: "Installation"
description: "Install cellphones from a release, with go install, or from source."
weight: 20
---

## Prebuilt binaries

Every [release](https://github.com/tamnd/cellphones-cli/releases) carries archives for Linux, macOS,
and Windows on amd64 and arm64, plus deb, rpm, and apk packages for Linux.
Download, unpack, put `cellphones` on your `PATH`, done. The `checksums.txt`
on each release is signed with keyless [cosign](https://docs.sigstore.dev/) if
you want to verify before running.

## With Go

```bash
go install github.com/tamnd/cellphones-cli/cmd/cellphones@latest
```

That puts `cellphones` in `$(go env GOPATH)/bin`, which is `~/go/bin` unless
you moved it. Make sure that directory is on your `PATH`.

## From source

```bash
git clone https://github.com/tamnd/cellphones-cli
cd cellphones-cli
make build        # produces ./bin/cellphones
./bin/cellphones version
```

## Container image

```bash
docker run --rm ghcr.io/tamnd/cellphones:latest --help
```

## Checking the install

```bash
cellphones version
```

prints the version and exits.
