# epcc-cli

A simple project for interacting with EPCC APIs via the command line, the goal is simplicity and quickness for API and not correctness or completeness.

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

#### Initialization

1. If you haven't already, export the following environment variables:

```shell
export EPCC_CLIENT_ID=<CLIENT_ID>
export EPCC_CLIENT_SECRET=<CLIENT_SECRET>
export EPCC_API_BASE_URL=https://api.moltin.com
```

2. Add epcc to your path:

```shell
cd <THIS_DIRECTORY>
export PATH=$PATH:($PWD)
```

3. If you haven't already, load the [completion](#completion) script for your shell above.

4. To test everything is working so far try running the following command:

`epcc get customers`

5. To add more profiles to use try the following command:
`epcc configure`

#### Simple CRUD

1. Start by typing `epcc cre` and hit **TAB**. The command line should complete to `epcc create`
2. Hit **TAB** (Twice depending on the shell) after `epcc create ` and you should see a list of resources that can be created in epcc.

```text
account                                    field                                      password_profile                           promotion                                  v2-product
account-management-authentication-token    file                                       pcm-hierarchy                              promotion_code                             
account_membership                         flow                                       pcm-node                                   user-authentication-password-profile-info  
currency                                   integration                                pcm-product                                user_authentication_info                   
customer                                   oidc-profile                               pcm-product-main-image                     user_authentication_oidc_profile_info
```

   in some cases the name of the resource matches the type in the JSON API (e.g., customer) and in some cases they differ (e.g., v2-product,pcm-product).
   Let's create a customer, type the following `epcc create customer` (using auto-complete) and hit ENTER

```shell
WARN[0000] POST https://api.moltin.com/v2/customers ==> HTTP/1.1 422 Unprocessable Entity 
{
  "errors": [
    {
      "detail": "The data.name field is required.",
      "title": "Failed Validation"
    },
    {
      "detail": "The data.email field is required.",
      "title": "Failed Validation"
    }
  ]
}Error: 422 Unprocessable Entity
ERRO[0000] Error occured while processing command 422 Unprocessable Entity
```

3. In the above you will see some JSON returned from the API. In some cases the error messages from the API might be tell you exactly what fields you need and the right casing,
   but some services don't give enough information. In any event, auto completion will come to the rescue. Try typing `epcc create customer` and hit **TAB** a few times.

```text
email     name      password
```

4. The parameters needed for the customers call are `email`, `name`, and `password`. 
   The epcc cli is a **thin client**, and designed to make exploring the API more seamless and reduce the boilerplate, but still should feel like the API.
   The syntax to supply the json is to use space separated key and values. So try typing `epcc create customer name "John Smith" password "hello123"` and hit enter.

```shell
WARN[0000] POST https://api.moltin.com/v2/customers 
{
  "data": {
    "type": "customer",
    "name": "John Smith",
    "password": "hello123"
  }
}WARN[0000] HTTP/1.1 422 Unprocessable Entity            
{
  "errors": [
    {
      "detail": "The data.email field is required.",
      "title": "Failed Validation"
    }
  ]
}Error: 422 Unprocessable Entity
ERRO[0000] Error occured while processing command 422 Unprocessable Entity
```

5. When the response code is not a `2xx`, `epcc` will output the sending JSON to help you debug what is being sent. In the above we are still missing an e-mail address so let's create it now `epcc create customers name "John Smith" password hello123 email test@test.com`

- Quotes are needed only to follow the standard rules of shell escaping (and a few other cases).

```shell
INFO[0001] POST https://api.moltin.com/v2/customers ==> HTTP/1.1 201 Created 
{
  "data": {
    "type": "customer",
    "id": "8f720da2-37d1-41b7-94da-3fd35d6b3c9b",
    "authentication_mechanism": "password",
    "email": "test@test.com",
    "name": "John Smith",
    "password": true
  }
}
```

6. To update a customer you use essentially the same syntax as create except with the `epcc update`, but now you need an id.
   To update the name you can use `epcc update customer 8f720da2-37d1-41b7-94da-3fd35d6b3c9b name "Jane Smith"`, and you will see the following output (if you use your ID and not the example one):

```shell
INFO[0001] PUT https://api.moltin.com/v2/customers/8f720da2-37d1-41b7-94da-3fd35d6b3c9b ==> HTTP/1.1 200 OK 
{
  "data": {
    "type": "customer",
    "id": "8f720da2-37d1-41b7-94da-3fd35d6b3c9b",
    "authentication_mechanism": "password",
    "email": "test@test.com",
    "name": "Jane Smith",
    "password": true
  }
}
```

7. Copying and pasting is terrible and as a result epcc-cli has a few ways of ameliorating the experience of working with ids. 
   To update the customer without the id, you can use an alias `last_customer` (and this will auto complete). For example `epcc update customer last_customer name "Jonah Smith"`
STEVE WILL FIX
