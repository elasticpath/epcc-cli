# epcc-cli

A simple project for interacting with EPCC APIs via the command line, the goal is simplicity and quickness for API and not correctness or compeletness.

## Getting Started

### Configuration

The following environment variables can be set up to control which environment and store to use with the cli.
To set the environment variables export the variable in your terminal.

e.g. `export EPCC_API_BASE_URL=https://api.moltin.com`

| Environment Variable   | Description                                                                                                                                                         |
|------------------------|---------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| EPCC_API_BASE_URL      | This is the API base URL which can be retrieved via CM.                                                                                                             |
| EPCC_CLIENT_ID         | This is the Client ID which can be retrieved via CM.                                                                                                                |                                            
| EPCC_CLIENT_SECRET     | This is the Client Secret which can be retrieved via CM.                                                                                                            |
| EPCC_BETA_API_FEATURES | This variable allows you to set [Beta Headers](https://documentation.elasticpath.com/commerce-cloud/docs/api/basics/api-contract.html#beta-apis) for all API calls. |

It is recommended to set EPCC_API_BASE_URL, EPCC_CLIENT_ID, and EPCC_CLIENT_SECRET to be able to interact with most things in the cli.

### Completion

For convenience this cli has been set up with auto-completion. To make the most of the cli start by running the following commands to set up completion for your shell:

#### Zsh

If shell completion is not already enabled in your environment, you will need to enable it.
Run the following command once:

`echo "autoload -U compinit; compinit" >> ~/.zshrc`

To load completions for each session, execute once:

`epcc completion zsh > “${fpath[1]}/_epcc`

You will need to start a new shell for this setup to take effect

#### Bash

You will need to have the Bash Completion package installed, and restart your bash session.

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

### Tutorial
