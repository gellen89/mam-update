# MaM Dynamic Update

Program that handles dynamically setting the latest IP Address from your seedbox for MyAnonamouse.

## Configuration

The `MAM_ID` is only required for the very first run. Subsequent runs will pull from the stored cookie, saved in the `$MAMUPDATE_DIR`.

### Environment Variables

| Variable      | Description                                                                                                                                               | Required | Default Value    |
| ------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------- | -------- | ---------------- |
| MAM_ID        | The MAM ID given to you after creating a new session. This only needs to be provided on the first run. Subsequent runs pull from the stored cookies file. | false    |                  |
| MAMUPDATE_DIR | The base directory that config, data, and cache are stored.                                                                                               | false    | $HOME/.mamupdate |

### CLI Flags

| Flag     | Description                                                                                                                                               | Required | Default Value    |
| -------- | --------------------------------------------------------------------------------------------------------------------------------------------------------- | -------- | ---------------- |
| -mam-id  | The MAM ID given to you after creating a new session. This only needs to be provided on the first run. Subsequent runs pull from the stored cookies file. | false    |                  |
| -mam-dir | The base directory that config, data, and cache are stored.                                                                                               | false    | $HOME/.mamupdate |
| -force   | Can be used to override the `last_run_time`                                                                                                               | false    | false            |

### Persistent Data

- `MAM.cookie`
  - This file will exist after the first run and stores HTTP Cookies from the HTTP Client cookie jar. This will be used used to populate the cookie jar on subsequent runs
- `MAM.ip`
  - This file stores the IP address found during the most recent run. It is used on subsequent runs to determine if the IP address has changed.
- `last_run_time`
  - This file stores a timestamp that conforms to `RFC3339` of the last time the script ran. Ensures the script does not run more than once an hour.

## Building

### Requirements

- Golang
- Make

To see the available commands, simply run `make` and it will print out what is available.

To do a simple build for your respective system, simply run:

```console
make build
```

This will output to `bin`.

You can then run `./bin/mam-update` to invoke the program.

## Artifacts

### Binaries

Pre-built binaries can be found attached in the [Github Releases](https://github.com/gellen89/mam-update/releases) section.

Binaries are built for common platforms: Macos, Linux, Windows on multiple architectures.
This also includes SBOM as well as checksums with a GPG signature.

### Docker Images

Docker images are pre-built for amd64 and arm64 and can be found in [ghcr.io for this repository](https://github.com/gellen89/mam-update/pkgs/container/mam-update).
The registry also containts SBOM and a sig for each release.
