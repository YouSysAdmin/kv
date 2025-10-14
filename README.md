# KV
KV is a secure and lightweight command-line key-value store built in Go.

[![Stand with Ukraine](https://raw.githubusercontent.com/vshymanskyy/StandWithUkraine/main/banner2-direct.svg)](https://github.com/vshymanskyy/StandWithUkraine/blob/main/docs/README.md)

## Features
- Key-value pair storage with optional buckets
- AES-256 encryption for all values
- Shared encryption key or separate encryption key for each bucket
- Import key-value from the AWS SSM Parameters service
- Read key value from a file, STDIN or plain tex

## Installation

```bash
go install github.com/yousysadmin/kv/cmd/kv@latest
```
```bash
# By default install to $HOME/.local/bin dir
curl -L https://raw.githubusercontent.com/yousysadmin/kv/master/scripts/install.sh | bash
```
Or download a release directly from GitHub: https://github.com/yousysadmin/kv/releases

## Usage

```bash
kv [command] [flags]
```

### Commands

- `kv add key <key>|<key@bucket> <value>` – Add or update a value
- `kv add bucket <name> - add a bucket`
- `kv get <key>|<key@bucket>` – Retrieve a value
- `kv list keys [<bucket>]` – List keys in the default or specified bucket
- `kv list buckets` – List all available buckets
- `kv delete key <key>|<key@bucket>` – Delete a key
- `kv delete bucket <bucket>` – Delete a bucket
- `kv version` – Show version information
- `kv import ssm` – Import KV from the AWS SSM service

The `kv list keys` command lists only keys in a bucket. You can use the `--values` flag to decrypt values and output them in `key:value` format.
You can also use the `--json` flag to format the output as JSON.
```
--json     output as json
--values   decrypt and output values
```

#### Completion
For session:
```shell
source <(kv completion zsh)
```

Permanent:
```shell
# https://blog.chmouel.com/posts/cobra-completions/#installation
# add to ~/.zshrc
autoload -U compinit;
compinit
mkdir -p ~/.zsh_completions/
fpath+=(~/.zsh_completions/)
```
```shell
# dump the completion in that directory
kv completion zsh > "${HOME}/.zsh_completions/_kv"
```

### Examples
#### Add bucket:
```shell
kv add bucket prod-secrets # create a bucket with a default encryption key
kv add bucket prod-secret --generate-new-key # create a bucket with a separate encryption key
```
#### Add key:
IMPORTANT: If the bucket does not exist, a new bucket will be created using the default encryption key.
           If you need a separate encryption key for the bucket, add a new bucket before KV.
```shell
kv add key my-key my-value # add KV to the default bucket
kv add key my-key@prod my-value # add KV to the `prod` bucket
kv add key longtext @readme.txt # read value from file
echo 'env=prod' | kv add key config@env @- # read value from stdin
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
# admin-password
# api-token

kv list keys --values
# Output:
# admin-password:SuperLongAdminPassword
# api-token:SuperLongToken

kv list keys --json
# Output:
# [{"key":"admin-password"},{"key":"api-token"}]

kv list keys --format=json --values
# Output:
# [{"key":"admin-password","value":"SuperLongAdminPassword"},{"key":"api-token","value":"SuperLongToken"}]

kv list keys --format=dotenv --values
# Output:
# PRIVATE_KEY="-----BEGIN RSA PRIVATE KEY-----\nHkVN9...\n-----END DSA PRIVATE KEY-----\n"

kv list keys --format=rails-dotenv --values
# Output:
# PRIVATE_KEY="-----BEGIN RSA PRIVATE KEY-----
# ...
# HkVN9...
# ...
# -----END DSA PRIVATE KEY-----"
```
#### Delete:
Important: The result of the `delete` operation cannot be undone.
```shell
kv delete key my-key # delete key from the default bucket
kv delete key my-key@prod # delete key from the `prod` bucket

kv delete bucket prod # delete the `prod` bucket and all related records
```

#### Import from SSM
```shell
# Import all secrets under a path using the default profile
# Region must be set in ~/.aws/config, ~/.aws/credentials or AWS_REGION environment variable
kv import ssm --bucket=mybucket /prod/secrets

# Use a specific AWS profile and region
kv import ssm --bucket=mybucket --profile=dev --region=ca-central-1 /dev/app/config

# Use static access key credentials (profile is ignored)
# Can use a default AWS CLI environment variable
# https://docs.aws.amazon.com/cli/v1/userguide/cli-configure-envvars.html#envvars-set
kv import ssm --bucket=mybucket --key-id=AKIA... --secret-key=... --region=us-west-2 /prod/secure

# Recursively import full key names under a prefix
# e.g., /prod/env/database_password
kv import ssm --bucket=mybucket --recursive /prod/env/

# Recursively import and trim key names
# e.g., /prod/env/database_password => database_password
kv import ssm --bucket=mybucket --recursive --trim-key-name /prod/env/

# Perform a dry-run import without writing data
kv import ssm --bucket=mybucket --dry-run /dev

# Dry-run and show key values for verification
kv import ssm --bucket=mybucket --dry-run --show-values /prod
```

## Configuration

> IMPORTANT:
> On first start, a default encryption key will be generated and saved to a file.
>
> You cannot change the encryption key for existing key-value records.

You can set encryption and DB path using CLI flags or environment variables:

| Option                 | CLI Flag                | Environment Variable        | Defaults  |
|------------------------|-------------------------|-----------------------------|-----------|
| Database path          | `--db`                  | `KV_DB_PATH`                | ~/.kv.db  |
| Encryption key         | `--encryption-key`      | `KV_ENCRYPTION_KEY`         | ""        |
| Encryption key store   | `--encryption-key-store`| `KV_ENCRYPTION_KEY_STORE`   | ~/.kv.key |

If no key is provided, a new one is automatically generated and stored in the file by path `~/.kv.key`.


