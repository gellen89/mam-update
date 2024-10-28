# MaM Dynamic Update

Program that handles dynamically setting the latest IP Address from your seedbox for MyAnonamouse

## Requirements

- Golang

## Building

To see the available commands, simply run `make` and it will print out what is available.

To do a simple build for your respective system, simply run:

```console
make build
```

This will output to `bin`.

You can then run `./bin/mam-update` to invoke the program.

## Environment Variables

| Variable      | Description                                                                                                                                                        | Required | Default Value |
| ------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------ | -------- | ------------- |
| MAM_ID        | The MAM ID given to you after creating a new session. This only needs to be provided on the first run. Subsequent runs pull from the stored cookies file.          | false    |               |
| MAMUPDATE_DIR | If not provided, defaults to $HOME. The program also sets a subdir called `.mamdynupdate`. So if not provided, all data files will exist in `$HOME/.mamdynupdate`. | false    | $HOME         |

## Persistent Data

All of this data is stored in `$MAMUPDATE_DIR/.mamdynupdate`

- `MAM.cookie`
  - This file will exist after the first run and stores HTTP Cookies from the HTTP Client cookie jar. This will be used used to popular the cookie jar on subsequent runs
- `MAM.ip`
  - This file stores the IP address found during the most recent run. It is used on subsequent runs to determine if the UIP address has changed.
- `last_run_time`
  - This file stores a timestamp that conforms to `RFC3339` of the last time the script ran. Ensures the script does not run more than once an hour.
