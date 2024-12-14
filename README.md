# MaM Dynamic Update

Program that handles dynamically setting the latest IP Address from your seedbox for MyAnonamouse.

## Configuration

The `MAM_ID` is only required for the very first run. Subsequent runs will pull from the stored cookie, saved in the `$MAMUPDATE_DIR`.

### Environment Variables

| Variable      | Description                                                                                                                                               | Required | Default Value          |
| ------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------- | -------- | ---------------------- |
| MAM_ID        | The MAM ID given to you after creating a new session. This only needs to be provided on the first run. Subsequent runs pull from the stored cookies file. | false    |                        |
| MAMUPDATE_DIR | The base directory that config, data, and cache are stored.                                                                                               | false    | $HOME/.mamupdate       |
| LOG_LEVEL     | Can be used to set the log level (debug, info, warn, error)                                                                                               | false    | info                   |
| IP_URL        | Can be used to set the URL used to retrieve an IP address                                                                                                 | false    | https://ifconfig.io/ip |

### CLI Flags

| Flag     | Description                                                                                                                                                                 | Required | Default Value  |
| -------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | -------- | -------------- |
| -mam-id  | The MAM ID given to you after creating a new session. This only needs to be provided on the first run. Subsequent runs pull from the stored cookies file. Overrides $MAM_ID | false    | $MAM_ID        |
| -mam-dir | The base directory that config, data, and cache are stored. Overrides $MAMUPDATE_DIR                                                                                        | false    | $MAMUPDATE_DIR |
| -force   | Can be used to override the `last_run_time`                                                                                                                                 | false    | false          |
| -level   | Can be used to set the log level (debug, info, warn, error). Overrides $LOG_LEVEL                                                                                           | false    | $LOG_LEVEL     |

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

## Release Verification

### Binaries

All binaries are signed with the GPG key found in the [KEYS](./KEYS) file.

Fingerprint: `7BE9 B51A 8DBE 0FE8 FBDE  323E 6737 C4DA ADBD 361B`

You can verify the computed checksums by doing the following:

1. Download the public key and the relevant `.sig` and `SHA256SUMS` file from the releases page.
2. Import it: `gpg --import KEYS`
3. Verify the signature: `gpg --verify mam-update_<version>_SHA256SUMS.sig mam-update_<version>_SHA256SUMS`

### Docker Images

All docker images are built using [ko](https://ko.build) and are signed with [cosign](https://github.com/sigstore/cosign)

The signing happens using `cosign` keyless sign mode via Github Actions.

To verify the image, run the following command:

```
cosign verify --certificate-oidc-issuer=https://github.com/login/oauth --certificate-identity-regexp github ghcr.io/gellen89/mam-update:latest
```

### Using the Docker Image

This example assumes running withing a VPN network like `gluetun`. If you aren't doing this, leave it off.
Just like the above configuration, the `MAM_ID` is only needed for the first invocation until there is a cached cookie. However, there is no harm in providing it everytime.

```
docker run \
-e MAM_ID=<token> \
-e MAMUPDATE_DIR=/mam \
-v /volume1/docker/mam/data:/mam \
--user $(id -u):$(id -g) \
--network container:gluetun \
ghcr.io/gellen89/mam-update:latest
```
