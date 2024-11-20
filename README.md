# Elastic Path Composable Commerce Command Line Interface

## Overview

This project is designed as a tool for power users to interact with the [Elastic Path Composable Commerce API](https://documentation.elasticpath.com/commerce-cloud/docs/api/) via the command line and the project is designed to fill three distinct
niches:

1. Provide a fast way for users familiar with the API to interact with it.
2. Provide a simpler way to do scripting with the API (i.e., instead of using curl and creating JSON in the shell)
3. Provide a reusable set of scripts for creating data sets with [Runbooks](docs/runbook-development.md).

This tool is not meant for new users unfamiliar with the API, new users are highly encouraged to use the [Elastic Path Composable Commerce Postman Collection](https://elasticpath.dev/docs/commerce-cloud/api-overview/test-with-postman-collection)
instead of this tool.

Additionally, this tool is not necessarily meant to be a new command line equivalent of Commerce Manager, it should just feel at all times like you are interacting with a JSON based REST API.

## Getting Started

### Installation

1. Download the appropriate release from the [GitHub Release Page](https://github.com/elasticpath/epcc-cli/releases).
2. Add the `epcc` binary to your path.
3. Load the autocompletion into your shell (See instructions [here](#completion)).

It is highly recommended that new users check out the [Tutorial](docs/tutorial.md).

### Command Overview

The following is a summary of the main commands, in general you can type `epcc help` to get an updated list and see all commands as well as flags.

#### CRUD Commands

| Command                                                     | Description                                                               |
|-------------------------------------------------------------|---------------------------------------------------------------------------|
| `epcc get <RESOURCE> [ID] ... [QUERY_PARAM_KEY] [VAL] ...`  | Retrieves either a list of objects, or an particular object from the API. |
| `epcc create <RESOURCE> [ID]... [KEY] [VAL] [KEY] [VAL]...` | Create an object.                                                         |
| `epcc update <RESOURCE> [ID]...[KEY] [VAL] [KEY] [VAL]...`  | Update an object.                                                         |
| `epcc delete <RESOURCE> [ID]...`                            | Delete an object.                                                         |

#### Authentication Commands

| Command                         | Description                                                       |
|---------------------------------|-------------------------------------------------------------------|
| `epcc login client_credentials` | Login to the API using a Client Credential Token                  |
| `epcc login customer`           | Login to the API using a Customer Token                           |
| `epcc login account-management` | Login to the API using an Account Management Authentication Token |
| `epcc login implicit`           | Login to the API using an Implicit Token                          |
| `epcc login status`             | Determine the current state of the login                          |

#### Debugging Commands

| Command                                            | Description                                                                  |
|----------------------------------------------------|------------------------------------------------------------------------------|
| `epcc docs <RESOURCE>`                             | Open the API docs for a resource in your browser                             |  
| `epcc docs <RESOURCE> [create/read/update/delete]` | Open the API docs for a resource with a specification action in your browser |
| `epcc aliases list`                                | List all known resource aliases                                              |
| `epcc resource-list`                               | List all supported resources                                                 |
| `epcc test-json [KEY] [VAL] [KEY] [VAL] ...`       | Render a JSON document based on the supplied key and value pairs             |

#### Power User Commands

| Command                                 | Description                                                                |
|-----------------------------------------|----------------------------------------------------------------------------|
| `epcc reset-store <STORE_ID>`           | Reset the store to an initial state (on a best effort basis)               |
| `epcc runbooks show <RUNBOOK> <ACTION>` | Show a specific runbook (script)                                           |
| `epcc runbooks validate`                | Validates all runbooks (built in and user supplied, outputting any errors) |
| `epcc runbooks run <RUNBOOK> <ACTION>`  | Run a specific runbook (script)                                            |

#### Tuning Runbooks

1. `--execution-timeout` will control how long the `epcc` process can run before timing out.
2. `--rate-limit` will control the number of requests per second to EPCC.
3. `--max-concurrency` will control the maximum number of concurrent commands that can run simultaneously.
    * This differs from the rate limit in that if a request takes 2 seconds, a rate limit of 3 will allow 6 requests in flight at a time, whereas `--max-concurrency` would limit you to 3. A higher value will slow down initial start time.

#### Headers

Headers can be set in one of three ways, depending on what is most convenient

1. Via the `-H` argument.
    * This header will be one time only.
2. Via the `EPCC_CLI_HTTP_HEADER_0` environment variable.
    * This header will be always be set.
3. Via the `epcc header set`
    * These headers will be set in the current profile and will stay until unset. You can see what headers are set with `epcc headers status`
    * Headers set this way support aliases.
    * You can also additionally group headers into groups with `--group` and then clear all headers with `epcc headers clear <GROUP>`

### Configuration

#### Via Prompts

Run the `epcc configure` and it will prompt you for the required settings, when you execute any EPCC CLI command you can pass in the `--profile` argument, or set the `EPCC_PROFILE` environment variable to use that profile.

#### Via Environment Variables

The following environment variables can be set up to control which environment and store to use with the EPCC CLI.

| Environment Variable                | Description                                                                                                                                                                                                                                                                                                                                                          |
|-------------------------------------|----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| EPCC_API_BASE_URL                   | This is the API base URL which can be retrieved via CM.                                                                                                                                                                                                                                                                                                              |
| EPCC_BETA_API_FEATURES              | This variable allows you to set [Beta Headers](https://documentation.elasticpath.com/commerce-cloud/docs/api/basics/api-contract.html#beta-apis) for all API calls.                                                                                                                                                                                                  |
| EPCC_CLI_HTTP_HEADER_**N**          | Setting any environment variable like this (where N is a number) will cause it's value to be parsed and added to all HTTP headers (e.g., `EPCC_CLI_HTTP_HEADER_0=Cache-Control: no-cache` will add `Cache-Control: no-cache` as a header). FYI, the surprising syntax is due to different encoding rules. You can also specify headers using `-H` or `epcc headers`. |
| EPCC_CLI_SUPPRESS_NO_AUTH_MESSAGES  | This will supress warning messages about not being authenticated or logged out                                                                                                                                                                                                                                                                                       |
| EPCC_CLI_URL_MATCH_REGEXP_**N**     | Setting this value causes the _path_ section of a URL to be matched and replaced with a corresponding value from the `EPCC_CLI_URL_MATCH_SUBSTITION_**N**` header, if not set the empty string is used.                                                                                                                                                              |
| EPCC_CLI_URL_MATCH_SUBSTITION_**N** | The replacement string to use when a match is found. Capture groups and back references are supported (see [ReplaceAllString](https://pkg.go.dev/regexp#Regexp.ReplaceAllString)).                                                                                                                                                                                   |
| EPCC_CLIENT_ID                      | This is the Client ID which can be retrieved via CM.                                                                                                                                                                                                                                                                                                                 |                                            
| EPCC_CLIENT_SECRET                  | This is the Client Secret which can be retrieved via CM.                                                                                                                                                                                                                                                                                                             |
| EPCC_PROFILE                        | A profile name that allows for an independent session and isolation (e.g., distinct histories).                                                                                                                                                                                                                                                                      |
| EPCC_RUNBOOK_DIRECTORY              | A directory that will be scanned for runbook, a runbook ends with `.epcc.yml`.                                                                                                                                                                                                                                                                                       |

It is recommended to set EPCC_API_BASE_URL, EPCC_CLIENT_ID, and EPCC_CLIENT_SECRET to be able to interact with most things in the CLI.

### Auto Completion

For convenience this cli has been set up with auto-completion. To make the most of the EPCC CLI start by running the following commands to set up completion for your shell:

#### Zsh

If shell completion is not already enabled in your environment, you will need to enable it.
Run the following command once:

`echo "autoload -U compinit; compinit" >> ~/.zshrc`

To load completions for each session, execute once:

`epcc completion zsh > “${fpath[1]}/_epcc`

You will need to start a new shell for this setup to take effect

#### Bash

You will need to have the [bash-completion](https://github.com/scop/bash-completion) (
e.g., [Ubuntu](https://packages.ubuntu.com/search?keywords=bash-completion), [Arch](https://archlinux.org/packages/extra/any/bash-completion/), [Gentoo](https://packages.gentoo.org/packages/app-shells/bash-completion)) package installed, and restart
your bash session.

To load completions for each session, execute once:

##### Linux

`epcc completion bash > /etc/bash_completion.d/epcc`

##### macOS

`epcc completion bash > /usr/local/etc/bash_completion.d/epcc`

#### PowerShell

For PowerShell run:

`epcc completion powershell | Out-String | Invoke-Expression`

To load completions for every new session, run:

`epcc completion powershell > epcc.ps1`

and source this file from your PowerShell profile.

#### fish

For fish run:

`epcc completion fish | source`

To load completions for each session, execute once:

`epcc completion fish > ~/.config/fish/completions/epcc.fish`

## Tips

### JQ Output

The `--output-jq` option can post process the output of `epcc create` `epcc get` and `epcc update`, for instance the following can be used to create richer
output.

```bash
$epcc create customer --auto-fill 
INFO[0000] (0001) POST https://api.moltin.com/v2/customers ==> HTTP/2.0 201 Created 
{
  "data": {
    "type": "customer",
    "id": "49d8e601-d110-42b7-99d2-60db73a6fb62",
    "authentication_mechanism": "password",
    "email": "thorabartell@gutmann.org",
    "name": "Michele Schuppe",
    "password": false
  }
}

$epcc create customer --auto-fill 
INFO[0001] (0001) POST https://api.moltin.com/v2/customers ==> HTTP/2.0 201 Created 
{
  "data": {
    "type": "customer",
    "id": "bf642721-44e5-4919-9fa5-b9c7da1ded1f",
    "authentication_mechanism": "password",
    "email": "kavondonnelly@yost.info",
    "name": "Matt Robel",
    "password": false
  }
}
$epcc get customers  --output-jq '.data[] | "\(.name) has id \(.id)"'
INFO[0000] (0001) GET https://api.moltin.com/v2/customers ==> HTTP/2.0 200 OK 
[
  "Michele Schuppe has id 49d8e601-d110-42b7-99d2-60db73a6fb62",
  "Matt Robel has id bf642721-44e5-4919-9fa5-b9c7da1ded1f"
]
```

The [JQ Manual](https://stedolan.github.io/jq/manual/) has some additional guidance on syntax, although
this is based on [GoJQ which has a number of differences](https://github.com/itchyny/gojq#difference-to-jq).

### Waiting for things

The `--retry-while-jq` argument can be used to wait for certain conditions to happen (e.g., a catalog publication, or an eventual consistency condition).

For example:

```bash
epcc get pcm-catalog-release --retry-while-jq '.data.meta.release_status != "PUBLISHED"' name=Ranges_Catalog last_release
```

The [JQ Manual](https://stedolan.github.io/jq/manual/) has some additional guidance on syntax, although
this is based on [GoJQ which has a number of differences](https://github.com/itchyny/gojq#difference-to-jq).

### How to determine the store you are using

```bash
epcc runbooks run misc get-store-info
```

## Development Tips

### Retries vs Ignoring Errors

Retries in epcc-cli will retry the _exact_ rendered request so if you are using templated parameters i.e., `auto-fill` and the failure is deterministic (say a unique constraint),
a retry will just get stuck in a loop. In this case if you want to create many different records, you want to `--ignore-errors`.

### Fast rebuilds

For development the following command using [Reflex](https://github.com/cespare/reflex) can speed up your development time, by recreating the command line tool.

```bash
git fetch --all --tags && reflex -v -s -r '(\.go$)|(resources.yaml|go.mod)|(runbooks/.+\.ya?ml)$' -- sh -c "go build -ldflags=\"-X github.com/elasticpath/epcc-cli/external/version.Version=$(git describe --tags --abbrev=0)+1 -X github.com/elasticpath/epcc-cli/external/version.Commit=$(git rev-parse --short HEAD)-dirty\" -o ./epcc" 
```

### Git Hooks

The following git pre-commit hook will run go fmt before committing anything

```bash
#!/bin/bash

echo "Running go fmt"
go fmt "./..."

echo "Adding changed files back to git"
git diff --cached --name-only --diff-filter=ACM | grep -E "\.(go)$" | xargs  git add
```
