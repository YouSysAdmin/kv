# KV
KV is a secure and lightweight command-line key-value store built in Go.

[![Stand with Ukraine](https://raw.githubusercontent.com/vshymanskyy/StandWithUkraine/main/banner2-direct.svg)](https://github.com/vshymanskyy/StandWithUkraine/blob/main/docs/README.md)

## Features
- Key-value pair storage with optional buckets
- AES-256 encryption for all values

## Installation

```bash
go install github.com/yousysadmin/kv/cmd/kv@latest
```
```bash
# By default install to $HOME/.bin dir
curl -L https://raw.githubusercontent.com/yousysadmin/kv/master/scripts/install.sh | bash
```
Or download a release directly from GitHub: https://github.com/yousysadmin/kv/releases

## Usage

```bash
kv [command] [flags]
```

### Commands

- `kv add <key>|<key@bucket> <value>` – Add or update a value
- `kv get <key>|<key@bucket>` – Retrieve a value
- `kv list keys [<bucket>]` – List keys in the default or specified bucket
- `kv list buckets` – List all available buckets
- `kv delete key <key>|<key@bucket>` – Delete a key
- `kv delete bucket <bucket>` – Delete a bucket
- `kv version` – Show version information

The `kv list keys` command lists only keys in a bucket. You can use the `--values` flag to decrypt values and output them in `key:value` format.
You can also use the `--json` flag to format the output as JSON.
```
--json     output as json
--values   decrypt and output values
```

### Examples
#### Add:
```shell
kv add my-key my-value # add KV to the default bucket
kv add my-key@prod my-value # add KV to the `prod` bucket
```
#### Get:
```shell
kv get my-key # get key from the default bucket
kv get my-key@prod # get key from the `prod` bucket

# Use output for pass a auth token for curl
curl -H "Auth:$(kv get token@prod-api)"  https://example.com
```
#### List:
```shell
kv list buckets # list available buckets
# Output:
# default
# prod
# stage

kv list keys # list keys in a bucket
# Output:
# amin-password
# api-token

kv list keys --values
# Output:
# amin-password:SuperLongAdminPassword
# api-token:SuperLongToken

kv list keys --json
# Output:
# [{"key":"amin-password"},{"key":"api-token"}]

kv list keys --json --values
# Output:
# [{"key":"amin-password","value":"SuperLongAdminPassword"},{"key":"api-token","value":"SuperLongToken"}]
```
#### Delete:
Important: The result of the `delete` operation cannot be undone.
```shell
kv delete key my-key # delete key from the default bucket
kv delete key my-key@prod # delete key from the `prod` bucket

kv delete bucket prod # delete the `prod` bucket and all related records
```

## Configuration

> IMPORTANT:
> On first start, a new encryption key will be generated and saved to a file.
>
> You cannot change the encryption key for existing key-value records.

You can set encryption and DB path using CLI flags or environment variables:

| Option                 | CLI Flag                | Environment Variable        | Defaults  |
|------------------------|-------------------------|-----------------------------|-----------|
| Database path          | `--db`                  | `KV_DB_PATH`                | ~/.kv.db  |
| Encryption key         | `--encryption-key`      | `KV_ENCRYPTION_KEY`         | ""        |
| Encryption key file    | `--encryption-key-file` | `KV_ENCRYPTION_KEY_FILE`    | ~/.kv.key |

If no key is provided, a new one is automatically generated and stored in the file by path `~/.kv.key`.


