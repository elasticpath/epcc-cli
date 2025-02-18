# Tutorial

## Initialization

1. If you haven't already, export the following environment variables:

### Bash
```shell
export EPCC_CLIENT_ID=<CLIENT_ID>
export EPCC_CLIENT_SECRET=<CLIENT_SECRET>
export EPCC_API_BASE_URL=https://useast.api.elasticpath.com
```

2. Add epcc to your path:

```shell
cd <THIS_DIRECTORY>
export PATH=$PATH:($PWD)
```

3. If you haven't already, please ensure that auto completion is enabled in your shell for the EPCC CLI. See the main README.md for more information.

4. To test everything is working so far try running the following command:


```shell
$epcc get customers
INFO[0000] (0001) GET https://useast.api.elasticpath.com/v2/customers ==> HTTP/2.0 200 OK 
{
  "data": [],
  "links": {
    "current": "https://useast.api.elasticpath.com/v2/customers?page[offset]=0",
    "first": "https://useast.api.elasticpath.com/v2/customers?page[offset]=0",
    "last": "https://useast.api.elasticpath.com/v2/customers?page[offset]=-100",
    "next": "https://useast.api.elasticpath.com/v2/customers?page[offset]=-1",
    "prev": "https://useast.api.elasticpath.com/v2/customers?page[offset]=0"
  },
  "meta": {
    "page": {
      "current": 1,
      "limit": 100,
      "offset": 0,
      "total": 0
    },
    "results": {
      "total": 0
    }
  }
}
```


## Simple CRUD

1. Start by typing `epcc cre` and hit **TAB**. The command line should complete to `epcc create`
2. Hit **TAB** (Twice depending on the shell) after `epcc create ` and you should see a list of resources that can be created in epcc.

```text
$epcc create 
account-address                            (Calls POST /v2/accounts/:accountId/addresses)
account                                    (Calls POST /v2/accounts)
account-entity-relationship                (Calls POST /v2/accounts/:accountId/relationships/:fieldId)
account-management-authentication-token    (Calls POST /v2/account-members/tokens)
account-member-entity-relationship         (Calls POST /v2/account-members/:accountMemberId/relationships/:fieldId)
account-membership                         (Calls POST /v2/accounts/:accountId/account-memberships)
application-key                            (Calls POST /v2/application-keys)
authentication-realm                       (Calls POST /v2/authentication-realms/)
authentication-realm-entity-relationship   (Calls POST /v2/authentication-realms/:authenticationRealmId/relationships/:fieldId)
cart                                       (Calls POST /v2/carts)
cart-checkout                              (Calls POST /v2/carts/:cartId/checkout)
...
```

in some cases the name of the resource matches the type in the JSON API (e.g., `customer`) and in some cases they differ (e.g., `v2-product`, `pcm-product`).
Let's create a customer, type the following `` (using auto-complete) and hit **ENTER**

```shell
$epcc create customer
WARN[0000] (0001) POST https://useast.api.elasticpath.com/v2/customers 
{
  "data": {
    "type": "customer"
  }
}
WARN[0000] HTTP/2.0 422 Unprocessable Entity            
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
}
Error: 422 Unprocessable Entity
ERRO[0000] Error occurred while processing command: 422 Unprocessable Entity
```

3. We aren't sure what all the necessary options for a customer can be, but you can use `epcc create customer --help` which will provide more details, there is a fair bit of detail including showing examples of how JSON encoding works, a link to the documentation allowed arguments, and some sample URL calls.
```shell
$epcc create customer --help
Creates a customer in a store/organization by calling /v2/customers.

Key and value pairs passed in will be converted to JSON with a jq like syntax.

The EPCC CLI will automatically determine appropriate wrapping (i.e., wrap the values in a data key or attributes key)

# Simple type with key and value 
epcc create customer key value => {"data":{"key":"b","type":"customer"}}

# Numeric types will be encoded as json numbers
epcc create customer key 1 => {"data":{"key":1,"type":"customer"}}
...
Documentation:
 https://documentation.elasticpath.com/commerce-cloud/docs/api/customers-and-accounts/customers/create-a-customer.html

Usage:
  epcc create customer [name NAME] [email EMAIL] [password PASSWORD] [flags]

Examples:
 ...

 # Create a customer passing in an argument
   epcc create customer name "Hello World"
 > POST /v2/customers
 > {
 >  "data": {
 >    "name": "Hello World",
 >    "type": "customer"
 >  }

```

4. In some cases the error messages from the API might be enough to tell you exactly what fields you need and the right casing but some endpoints don't give the clearest indication of the JSON required, or may not tell you the right value, another **highly recommended** option is to allow auto-completion will come to the rescue. Try typing `epcc create customer` and hit **TAB** a few times.

```text
email     name      password
```

5. The parameters needed for the customers call are `email`, `name`, and `password`.
   The epcc cli is a **thin client**, and designed to make exploring the API more seamless and reduce the boilerplate, but still should feel like a JSON API.
   The syntax to supply the json is to use space separated key and values. So try typing `epcc create customer name "John Smith" password "hello123"` and hit enter.

```shell
WARN[0000] POST https://useast.api.elasticpath.com/v2/customers 
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

6. When the response code is not a `2xx`, `epcc` will output the sending JSON to help you debug what is being sent. In the above we are still missing an e-mail address so let's create it now `epcc create customers name "John Smith" password hello123 email test@test.com`

- Quotes are needed only to follow the standard rules of shell escaping (and a few other cases).

```shell
INFO[0001] POST https://useast.api.elasticpath.com/v2/customers ==> HTTP/1.1 201 Created 
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
7. To update a customer you use essentially the same syntax as create except with the `epcc update`, but now you need an id. With `epcc` ids are the first position arguments after the resource, so to update a customer you can use `epcc update customer <ID>`. However, this likely means copying and pasting data, and so we will leverage **aliases**. To see a list of aliases you can run `epcc aliases list` and you should see a set of aliases like so:
```shell
  Resource Type ||                              Alias Name || Values
   customer                            email=test@test.com => ID: ecaee2c4-06db-48a3-a709-6049ee4cb9e2
   customer        id=ecaee2c4-06db-48a3-a709-6049ee4cb9e2 => ID: ecaee2c4-06db-48a3-a709-6049ee4cb9e2
   customer                               last_read=entity => ID: ecaee2c4-06db-48a3-a709-6049ee4cb9e2
   customer                                name=John_Smith => ID: ecaee2c4-06db-48a3-a709-6049ee4cb9e2
```
8. All of the following commands can be used to update the previous customer, you can of course pass in an id, but aliases support auto-completion and are predictable so when writing scripts an alias can be much simpler and help avoid the need to parse JSON.  Aliases are created based on the attributes `sku`, `slug`,`name`,`code`,`email` and `id` on seen resources.
   * epcc update customer name=John_Smith name "Jake Smith"  
   * epcc update customer email=test@test.com name "James Smith"
   * epcc update customer last_read=entity name "Jacob Smith"

9. The alias `last_read=entity` is a special one, and it refers to the last entity with type `customer` that the EPCC CLI saw, similarly when you get a list of all entities (i.e., epcc get customers), you will see aliases like `last_read=array[0]` which is the first element in the last array it was seen. Other aliases may exist as well, so use the `epcc aliases list <resource_type>` to see them.
   - Some resources require other resource IDs at creation to tie related components together (for example, when adding components to bundles).  In these cases, the generated ID might not be immediately available for entry, though aliases for the required components exist.  To solve this, use the pattern `<target_id_key> alias/<resource>/name=<alias_name>/id`.  This syntax tells the system to extract the ID field from the resource identified by the alias.
   - For example, the `publish-catalog-with-bundles` action under the `pxm-how-to.yml` runbook includes the line `components.dogbed.options[0].id alias/product/name=Fluffy_Bed/id`.  This command says that the `dogbed` component ID (under `options`) should use the ID from the product aliased with the name `Fluffy_Bed`.
   - For create you can use the `--save-as-alias` to create a specific alias for a resource, this might make it easier to use in your script.

10. The output of the epcc cli contains a lot of output, however almost all the output is on [standard error](https://en.wikipedia.org/wiki/Standard_streams#Standard_error_(stderr)), and the only thing *for crud commands* that is output is the JSON response, this means you can pipe the output to a JSON processing library like so:
```shell
$epcc get customer name=John_Smith  | jq .data.email
INFO[0000] (0001) GET https://useast.api.elasticpath.com/v2/customers/ecaee2c4-06db-48a3-a709-6049ee4cb9e2 ==> HTTP/2.0 200 OK 
"test@test.com"


11. That said, you can also use the --output-jq argument to process JQ without installing jq directly (Although this is based on [GoJQ which has a number of differences](https://github.com/itchyny/gojq#difference-to-jq).)
```shell
epcc get customer name=John_Smith --output-jq  ".data.email"
INFO[0000] (0001) GET https://useast.api.elasticpath.com/v2/customers/c22ef03d-fe3d-4042-9a60-acce4811660c ==> HTTP/1.1 200 OK 
"test@test.com"

```

## Advanced Features

1. If you try and create a customer address, you will notice that there is a lot of things required:
```shell
$epcc create customer-address name=John_Smith 
WARN[0000] (0001) POST https://useast.api.elasticpath.com/v2/customers/ecaee2c4-06db-48a3-a709-6049ee4cb9e2/addresses 
{
  "data": {
    "type": "address"
  }
}
WARN[0000] HTTP/2.0 400 Bad Request                     
{
  "errors": [
    {
      "detail": "first_name is required",
      "source": "data",
      "status": 400,
      "title": "required"
    },
    {
      "detail": "last_name is required",
      "source": "data",
      "status": 400,
      "title": "required"
    },
    {
      "detail": "line_1 is required",
      "source": "data",
      "status": 400,
      "title": "required"
    },
    {
      "detail": "county is required",
      "source": "data",
      "status": 400,
      "title": "required"
    },
    {
      "detail": "postcode is required",
      "source": "data",
      "status": 400,
      "title": "required"
    },
    {
      "detail": "country is required",
      "source": "data",
      "status": 400,
      "title": "required"
    }
  ]
}
```
2. You can check the documentation of anything by running `epcc docs customer-addresses`, which will open up documentation for the resource. `epcc docs customer-addresses create` 

3. Many resources support simply having *auto-filled values* for when you need the resource created but don't care about values. To create an address for the above customer, with a name "My New Address", and a region "WA" use the following:
```shell
$epcc create customer-address name=John_Smith --auto-fill name "My New Address" region "WA"
INFO[0000] (0001) POST https://useast.api.elasticpath.com/v2/customers/ecaee2c4-06db-48a3-a709-6049ee4cb9e2/addresses ==> HTTP/2.0 201 Created 
{
  "data": {
    "type": "address",
    "id": "809b64a5-ec25-403e-a14d-e256a6a969d9",
    "city": "Pittsburgh",
    "company_name": "USSearch",
    "country": "MO",
    "county": "IL",
    "first_name": "Casimir",
    "instructions": "guess what",
    "last_name": "Bahringer",
    "line_1": "207 New Isle furt",
    "line_2": "",
    "links": {
      "self": "https://useast.api.elasticpath.com/v2/addresses/809b64a5-ec25-403e-a14d-e256a6a969d9"
    },
    "name": "My New Address",
    "phone_number": "329.987.9543",
    "postcode": "64153",
    "region": "WA",
    "meta": {
      "timestamps": {
        "created_at": "2022-12-03T15:10:00.487679744Z",
        "updated_at": "2022-12-03T15:10:00.487679744Z"
      }
    }
  }
}
```
4. If you want to turn down the verbosity of the output, you can use `-s` which will not print the output, but you can also print friendly output with some advanced JQ.
```shell
$epcc create customer-address name=John_Smith --auto-fill name "My New Address" region "WA" -s
INFO[0000] (0001) POST https://useast.api.elasticpath.com/v2/customers/c22ef03d-fe3d-4042-9a60-acce4811660c/addresses ==> HTTP/1.1 201 Created 
$epcc create customer-address name=John_Smith --auto-fill name "My New Address" region "WA" --output-jq '.data | "Address ID:\(.id) created in city: \(.city), country: \(.country)"'
INFO[0000] (0001) POST https://useast.api.elasticpath.com/v2/customers/c22ef03d-fe3d-4042-9a60-acce4811660c/addresses ==> HTTP/1.1 201 Created 
"Address ID:d4fadda0-df21-4304-894b-70a85fd1ccf0 created in city: Irvine, country: CR"
```

5. You can see the set of API calls that were made with the `epcc logs list` command, and see the full HTTP request and response with `epcc logs show <ID>`, a negative id like -2 would show the second last request.

```shell
$epcc logs list
...
392 7:08AM DELETE /v2/customers/ecaee2c4-06db-48a3-a709-6049ee4cb9e2/addresses/c1f997dd-4d2c-49a5-a775-763e680ddba5 ==> 204
393 7:08AM GET /v2/customers/ecaee2c4-06db-48a3-a709-6049ee4cb9e2/addresses ==> 200
394 7:10AM POST /v2/customers/ecaee2c4-06db-48a3-a709-6049ee4cb9e2/addresses ==> 201

$epcc logs show 394
POST /v2/customers/ecaee2c4-06db-48a3-a709-6049ee4cb9e2/addresses HTTP/1.1
Host: useast.api.elasticpath.com
Transfer-Encoding: chunked
Content-Type: application/json
Accept-Encoding: gzip

125
{"data":{"city":"Pittsburgh","company_name":"USSearch","country":"MO","county":"IL","first_name":"Casimir","instructions":"guess what","last_name":"Bahringer","line_1":"207 New Isle furt","name":"My New Address","phone_number":"329.987.9543","postcode":"64153","region":"WA","type":"address"}}
0


HTTP/2.0 201 Created
Content-Type: application/json; charset=utf-8
Date: Sat, 03 Dec 2022 15:10:00 GMT
Strict-Transport-Security: max-age=31557600
Via: 1.1 varnish

{"data":{"id":"809b64a5-ec25-403e-a14d-e256a6a969d9","type":"address","name":"My New Address","first_name":"Casimir","last_name":"Bahringer","company_name":"USSearch","phone_number":"329.987.9543","line_1":"207 New Isle furt","line_2":"","city":"Pittsburgh","postcode":"64153","county":"IL","region":"WA","country":"MO","instructions":"guess what","links":{"self":"https://useast.api.elasticpath.com/v2/addresses/809b64a5-ec25-403e-a14d-e256a6a969d9"},"meta":{"timestamps":{"created_at":"2022-12-03T15:10:00.487679744Z","updated_at":"2022-12-03T15:10:00.487679744Z"}}}}
```

6. You can open commerce manager with the command `epcc commerce-manager` which will open the correct commerce manager based on which API endpoint you are using.

7. If you need to wait for a condition to occur (i.e., a catalog publication) the `--retry-while-jq` option can help, it will retry the request until the jq supplied returns `true`. For example:
```shell
  epcc get pcm-catalog-release --retry-while-jq '.data.meta.release_status != "PUBLISHED"' name=Ranges_Catalog pxm-how-to-create-catalog-and-publish-release
```

8. If you want to conditionally do something (i.e., create a currency which should only be done once, the `--if-alias-exists` and `--if-alias-does-not-exist` command can help), this can improve idempotency of scripts. 
```shell
$epcc create currency --if-alias-does-not-exist code=GBP code GBP exchange_rate 1 format £{price} decimal_point "." thousand_separator , decimal_places 2 default false enabled true
INFO[0001] (0001) POST https://useast.api.elasticpath.com/v2/currencies ==> HTTP/1.1 201 Created 
{
  "data": {
    "type": "currency",
    "id": "374de533-0964-479b-bd34-4664180f606b",
    "code": "GBP",
    "decimal_places": 2,
    "decimal_point": ".",
    "default": false,
    "enabled": true,
    "exchange_rate": 1,
    "format": "£{price}",
    "thousand_separator": ",",
    "links": {
      "self": "https://useast.api.elasticpath.com/currencies/374de533-0964-479b-bd34-4664180f606b"
    },
    "meta": {
      "owner": "store",
      "timestamps": {
        "created_at": "2023-06-18T16:15:30.045693431Z",
        "updated_at": "2023-06-18T16:15:30.045694681Z"
      }
    }
  }
}
$epcc create currency --if-alias-does-not-exist code=GBP code GBP exchange_rate 1 format £{price} decimal_point "." thousand_separator , decimal_places 2 default false enabled true
INFO[0000] Alias [code=GBP] does exist (value: 374de533-0964-479b-bd34-4664180f606b), not continuing run 
$epcc delete currency --if-alias-exists code=GBP code=GBP
INFO[0001] (0001) DELETE https://useast.api.elasticpath.com/v2/currencies/374de533-0964-479b-bd34-4664180f606b ==> HTTP/1.1 204 No Content 
null
$epcc delete currency --if-alias-exists code=GBP code=GBP
INFO[0000] Alias [code=GBP] does not exist, not continuing run
```

### Advanced JSON Encoding

The Elastic Path Cloud Commerce API has a few different conventions for how JSON documents should be shaped:

1. Many resources nest most values under the `data.attributes` key.
2. Many resources nest most values under the `data` key.
3. A few resources do not nest their values under anything but a root object.

For each resource type that is defined in the EPCC CLI, additional metadata is stored for whether, or not data should be encoded in an attributes key. The `epcc test-json` command can be used to show various outputs, as well as how to encode various arrays.

```shell
$epcc test-json a 1 b[0] 2 b[1] 3 c.d "true" c.e \"true\" c.f null c.g '"null"' c.h.i 3 c.h.j '"3"' k [] 
{
  "data": {
    "a": 1,
    "b": [
      2,
      3
    ],
    "c": {
      "d": true,
      "e": "true",
      "f": null,
      "g": "null",
      "h": {
        "i": 3,
        "j": "3"
      }
    },
    "k": []
  }
}
$epcc test-json --compliant  a 1 b[0] 2 b[1] 3 c.d "true" c.e \"true\" c.f null c.g '"null"' c.h.i 3 c.h.j '"3"' k [] 
{
  "data": {
    "attributes": {
      "a": 1,
      "b": [
        2,
        3
      ],
      "c": {
        "d": true,
        "e": "true",
        "f": null,
        "g": "null",
        "h": {
          "i": 3,
          "j": "3"
        }
      },
      "k": []
    }
  }
}
```

In both of the above examples, we can see how the EPCC CLI converts keys and values into a JSON object:
* `a 1` is converted into a key `a` with numeric value `1`.
* `b[0] 2 b[1] 3` creates array `b` with value `[2 3]`
* `c.d "true"` creates a `d` field in the `c` object with a *boolean* value `true`.
* `c.e \"true\"` creates a `e` field in the `c` object with a *string* value `"true"`. The EPCC CLI will by auto detect the type as much as possible, but in some cases (e.g., numeric skus you will to ensure that the EPCC CLI sees a value in quotes).
* `c.f "null" c.g '"null"'` creates an `f`, and `g` in the `c` object, with `null` and `"null"` respectively.  
* `c.h.i 3 c.h.j '"3"'` creates a double nested object under `c` and `h` with values `3` and `"3"` respectively.
* `k []` creates an empty array.

Another example of output is as follows:

```shell
$epcc test-json  relationships.test.type customer relationships.test.id foo attributes.hello world
{
  "data": {
    "attributes": {
      "hello": "world"
    },
    "relationships": {
      "test": {
        "type": "customer",
        "id": "foo"
      }
    }
  }
}
$epcc test-json  --compliant  relationships.test.type customer relationships.test.id foo attributes.hello world
{
  "data": {
    "attributes": {
      "hello": "world"
    },
    "relationships": {
      "test": {
        "type": "customer",
        "id": "foo"
      }
    }
  }
}
```

The important thing to note in the above, is that the first `attributes` value is ignored if it's a prefix of a compliant resource, the second is that values that start with `relationships` are not nested under `attributes`.

#### Dashed Arguments

If an argument starts with a dash, it can be treated as an option, this can make certain values hard to put in, you can use the `--` to stop argument processing and force values to be interpreted as a JSON afterwards.

For example:

```bash
$epcc test-json hello "-world" value -3.14
Error: unknown shorthand flag: 'w' in -world
ERRO[0000] Error occurred while processing command: unknown shorthand flag: 'w' in -world 
$epcc test-json hello -- "-world" value -3.14
{
  "data": {
    "hello": "-world",
    "value": -3.14
  }
}
```

### Runbooks

The EPCC CLI supports Runbooks which are either built in or external scripts that can be used to configure a store a certain way. These scripts can typically execute very fast and in parallel.

1. If you run the `epcc runbooks show` command you can see a list of runbooks that are supported.
```shell
epcc runbooks show 
Display the runbook contents

Usage:
  epcc runbooks show [command]

Available Commands:
  account-management               Sample commands for using account management
  extend-customer-resources-how-to Create flows on the customer resources
  hello-world                      A hello world runbook
  manual-gateway-how-to            Manual gateway tutorial
  pxm-how-to                       Create and manipulate the Getting Started PXM Catalog
```

2. Each runbook consists of a set of actions, so we can look at the `pxm-how-to` runbook with `epcc runbooks show pxm-how-to`
```shell
$epcc runbooks show pxm-how-to 
Creates the Getting Started PXM Catalog (https://documentation.elasticpath.com/commerce-cloud/docs/developer/how-to/get-started-pcm.html).

Usage:
  epcc runbooks show pxm-how-to [command]

Available Commands:
  create-catalog-and-publish Create the sample catalog and publish
  create-catalog-rule        Create a couple customers and a catalog rule, both customers use a password of `password`
  products-with-custom-data  An example of how to create a product with custom data
  products-with-variations   An example of how to create a product with variations
  reset                      Delete all catalogs, hierarchy and nodes
```

3. You can see what a runbook will do, by selecting the available action, so for example:

```shell
$epcc runbooks show pxm-how-to create-catalog-and-publish 
epcc create pcm-hierarchy name "Major Appliances" description "Free standing appliances" slug "Major-Appliances-MA0"
epcc create pcm-node name=Major_Appliances name "Ranges" description "All stoves and ovens" slug "Ranges-MA1"
epcc create pcm-node name=Major_Appliances name "Electric Ranges" description "Electric stoves and ovens" slug "Electric-Ranges-MA2" relationships.parent.data.type node relationships.parent.data.id name=Ranges
epcc create pcm-node name=Major_Appliances name "Gas Ranges" description "Gas stoves and ovens" slug "Gas-Ranges-MA2" relationships.parent.data.type node relationships.parent.data.id name=Ranges
epcc create pcm-product name "BestEver Electric Range" sku "BE-Electric-Range-1a1a" slug "bestever-range-1a1a" description "This electric model offers an induction heating element and convection oven." status live commodity_type physical upc_ean \"111122223333\" mpn BE-R-1111-aaaa-1a1a
epcc create pcm-node-product name=Major_Appliances name=Electric_Ranges data[0].type product data[0].id name=BestEver_Electric_Range
epcc create pcm-product name "BestEver Gas Range" sku "BE-Gas-Range-2b2b" slug "bestever-range-2b2b" description "This gas model includes a convection oven." status live commodity_type physical upc_ean \"222233334444\" mpn BE-R-2222-bbbb-2b2b
epcc create pcm-node-product name=Major_Appliances name=Gas_Ranges data[0].type product data[0].id name=BestEver_Gas_Range
epcc create currency code GBP exchange_rate 1 format £{price} decimal_point "." thousand_separator , decimal_places 2 default false enabled true
epcc create pcm-pricebook name "Preferred Pricing" description "Catalog with pricing suitable for high-volume customers."
epcc create pcm-product-price name=Preferred_Pricing currencies.USD.amount 300000 currencies.USD.includes_tax false currencies.GBP.amount 250000 currencies.GBP.includes_tax false sku BE-Electric-Range-1a1a
epcc create pcm-product-price name=Preferred_Pricing currencies.USD.amount 350000 currencies.USD.includes_tax false currencies.GBP.amount 300000 currencies.GBP.includes_tax false sku BE-Gas-Range-2b2b
epcc create pcm-catalog name "Ranges Catalog" description "Ranges Catalog" pricebook_id name=Preferred_Pricing hierarchy_ids[0] name=Major_Appliances
epcc create pcm-catalog-release name=Ranges_Catalog
epcc get pcm-catalog-releases name=Ranges_Catalog
```

4. Runbooks can be run with `epcc runbooks run pxm-how-to create-catalog-and-publish`
```shell
epcc runbooks run pxm-how-to create-catalog-and-publish 
INFO[0000] (0001) POST https://useast.api.elasticpath.com/pcm/hierarchies ==> HTTP/2.0 201 Created 
INFO[0001] (0002) POST https://useast.api.elasticpath.com/pcm/hierarchies/1777c591-dee5-4a6a-a2d9-72b8af2e582c/nodes ==> HTTP/2.0 201 Created 
INFO[0001] (0003) POST https://useast.api.elasticpath.com/pcm/hierarchies/1777c591-dee5-4a6a-a2d9-72b8af2e582c/nodes ==> HTTP/2.0 201 Created 
INFO[0001] (0004) POST https://useast.api.elasticpath.com/pcm/hierarchies/1777c591-dee5-4a6a-a2d9-72b8af2e582c/nodes ==> HTTP/2.0 201 Created 
INFO[0002] (0005) POST https://useast.api.elasticpath.com/pcm/products ==> HTTP/2.0 201 Created 
INFO[0002] (0006) POST https://useast.api.elasticpath.com/pcm/hierarchies/1777c591-dee5-4a6a-a2d9-72b8af2e582c/nodes/e22cdcf1-3d99-4bf6-bb0e-04be9d04e6d8/relationships/products ==> HTTP/2.0 201 Created 
INFO[0002] (0007) POST https://useast.api.elasticpath.com/pcm/products ==> HTTP/2.0 201 Created 
INFO[0003] (0008) POST https://useast.api.elasticpath.com/pcm/hierarchies/1777c591-dee5-4a6a-a2d9-72b8af2e582c/nodes/84412f8d-053a-4a6e-9617-2f53e24d1a20/relationships/products ==> HTTP/2.0 201 Created 
INFO[0003] (0009) POST https://useast.api.elasticpath.com/v2/currencies ==> HTTP/2.0 201 Created 
INFO[0003] (0010) POST https://useast.api.elasticpath.com/pcm/pricebooks ==> HTTP/2.0 201 Created 
INFO[0003] (0011) POST https://useast.api.elasticpath.com/pcm/pricebooks/cc0eaf11-d4e7-44fd-92e4-d26733548a32/prices ==> HTTP/2.0 201 Created 
INFO[0003] (0012) POST https://useast.api.elasticpath.com/pcm/pricebooks/cc0eaf11-d4e7-44fd-92e4-d26733548a32/prices ==> HTTP/2.0 201 Created 
INFO[0004] (0013) POST https://useast.api.elasticpath.com/pcm/catalogs ==> HTTP/2.0 201 Created 
INFO[0004] (0014) POST https://useast.api.elasticpath.com/pcm/catalogs/2b3ae101-75cc-4cdd-9537-b2638120cdc5/releases ==> HTTP/2.0 201 Created 
INFO[0004] (0015) GET https://useast.api.elasticpath.com/pcm/catalogs/2b3ae101-75cc-4cdd-9537-b2638120cdc5/releases ==> HTTP/2.0 200 OK
```

5. Runbooks can optionally take command line arguments as well, but that is outside the scope of this tutorial. For creating your own runbooks see [Runbook Development](runbook-development.md)

6. Runbooks can have many arguments, and many have a `reset` command, so if you want to clean up the above, you could run `epcc run pxm-how-to reset` which should clean up everything.
 
### Customer Authentication

Assuming you have created the customer above, and run the create catalog and publish step then we should be able to check out a cart on behalf of a customer.

1. To login on behalf of a customer, use `epcc login customer email test@test.com password hello123`

```shell
epcc login customer email test@test.com password hello123
INFO[0001] (0001) POST https://useast.api.elasticpath.com/v2/customers/tokens ==> HTTP/2.0 200 OK 
WARN[0001] You are current logged in with client_credentials, please switch to implicit with `epcc login implicit` to use customer token correctly. Mixing client_credentials and customer token can lead to unintended results. 
INFO[0001] (0002) GET https://useast.api.elasticpath.com/v2/customers/ecaee2c4-06db-48a3-a709-6049ee4cb9e2 ==> HTTP/2.0 200 OK 
INFO[0001] Successfully authenticated as customer John Smith <test@test.com>, session expires Sun, 04 Dec 2022 07:48:47 -0800 
{
  "data": {
    "type": "token",
    "id": "fe11ccf4-a7ba-4d2b-a03a-8581707d2e94",
    "customer_id": "ecaee2c4-06db-48a3-a709-6049ee4cb9e2",
    "expires": 1510684200,
    "token": "eyJhbGciOiAiSFMyNTYiLCAidHlwIjogIkpXVCJ9.eyJzdWIiOiI3OWNjMDQ4Ni1iYmRmLTQ5MWItYTBhMi03MjIzODNiNjI4OGIiLCJuYW1lIjoiUm9uIFN3YW5zb24iLCJleHAiOjE1MTA2ODQyMDAsImlhdCI6MTUxMDU5NzgwMCwianRpIjoiMzZmMDU5NDAtMGQzOC00MTFhLTg5MDktM2FlYTU4YmMxZjA5In0=.ea948e346d0683803aa4a2c09441bcbf7c79bd9234bed2ce8456ab3af257ea9f"
  }
}
```

2. In the above output, there is a warning that we are logged in with client_credentials (which is the default), and for authenticating as a customer we should use the implicit token. We can switch as shown below:

```shell
epcc login implicit
INFO[0000] POST https://useast.api.elasticpath.com/oauth/access_token ==> HTTP/2.0 200 OK 
INFO[0000] Successfully authenticated with implicit token, session expires Sat, 03 Dec 2022 08:50:51 -0800 
```

3. We can create a cart as follows:
```shell
$epcc create cart name "My Cart" description "Cart"
INFO[0000] (0001) POST https://useast.api.elasticpath.com/v2/carts ==> HTTP/2.0 201 Created 
{
  "data": {
    "type": "cart",
    "id": "3f81126c-723c-4f49-9fee-5fb0466aeb5c",
    "description": "Cart",
    "links": {
      "self": "https://useast.api.elasticpath.com/v2/carts/3f81126c-723c-4f49-9fee-5fb0466aeb5c"
    },
    "name": "My Cart",
    "relationships": {
      "customers": {
        "data": [
          {
            "type": "customer",
            "id": "ecaee2c4-06db-48a3-a709-6049ee4cb9e2"
          }
        ]
      },
      "items": {
        "data": null
      }
    },
    "meta": {
      "display_price": {
        "discount": {
          "amount": 0,
          "currency": "",
          "formatted": "0"
        },
        "tax": {
          "amount": 0,
          "currency": "",
          "formatted": "0"
        },
        "with_tax": {
          "amount": 0,
          "currency": "",
          "formatted": "0"
        },
        "without_discount": {
          "amount": 0,
          "currency": "",
          "formatted": "0"
        },
        "without_tax": {
          "amount": 0,
          "currency": "",
          "formatted": "0"
        }
      },
      "timestamps": {
        "created_at": "2022-12-03T15:51:44Z",
        "expires_at": "2022-12-10T15:51:44Z",
        "updated_at": "2022-12-03T15:51:44Z"
      }
    }
  }
}
```

4. We can add an item to the cart as follows:
```shell
epcc create cart-product-item name=My_Cart sku sku=BE-Gas-Range-2b2b quantity 1
INFO[0000] (0001) POST https://useast.api.elasticpath.com/v2/carts/3f81126c-723c-4f49-9fee-5fb0466aeb5c/items ==> HTTP/2.0 201 Created 
{
  "data": [
    {
      "type": "cart_item",
      "id": "209749d8-917f-4367-9443-f4b75e77bbaf",
      "catalog_id": "2b3ae101-75cc-4cdd-9537-b2638120cdc5",
      "catalog_source": "pim",
      "description": "This gas model includes a convection oven.",
      "image": {
        "file_name": "",
        "href": "",
        "mime_type": ""
      },
      "links": {
        "product": "https://useast.api.elasticpath.com/v2/products/87132b87-ae32-4cbc-a3c8-7875923c7217"
      },
      "manage_stock": false,
      "name": "BestEver Gas Range",
      "product_id": "87132b87-ae32-4cbc-a3c8-7875923c7217",
      "quantity": 1,
      "sku": "BE-Gas-Range-2b2b",
      "slug": "bestever-range-2b2b",
      "unit_price": {
        "amount": 350000,
        "currency": "USD",
        "includes_tax": false
      },
      "value": {
        "amount": 350000,
        "currency": "USD",
        "includes_tax": false
      },
      "meta": {
        "display_price": {
          "discount": {
            "unit": {
              "amount": 0,
              "currency": "USD",
              "formatted": "$0.00"
            },
            "value": {
              "amount": 0,
              "currency": "USD",
              "formatted": "$0.00"
            }
          },
          "tax": {
            "unit": {
              "amount": 0,
              "currency": "USD",
              "formatted": "$0.00"
            },
            "value": {
              "amount": 0,
              "currency": "USD",
              "formatted": "$0.00"
            }
          },
          "with_tax": {
            "unit": {
              "amount": 350000,
              "currency": "USD",
              "formatted": "$3,500.00"
            },
            "value": {
              "amount": 350000,
              "currency": "USD",
              "formatted": "$3,500.00"
            }
          },
          "without_discount": {
            "unit": {
              "amount": 350000,
              "currency": "USD",
              "formatted": "$3,500.00"
            },
            "value": {
              "amount": 350000,
              "currency": "USD",
              "formatted": "$3,500.00"
            }
          },
          "without_tax": {
            "unit": {
              "amount": 350000,
              "currency": "USD",
              "formatted": "$3,500.00"
            },
            "value": {
              "amount": 350000,
              "currency": "USD",
              "formatted": "$3,500.00"
            }
          }
        },
        "timestamps": {
          "created_at": "2022-12-03T15:52:54Z",
          "updated_at": "2022-12-03T15:52:54Z"
        }
      }
    }
  ],
  "meta": {
    "display_price": {
      "discount": {
        "amount": 0,
        "currency": "USD",
        "formatted": "$0.00"
      },
      "tax": {
        "amount": 0,
        "currency": "USD",
        "formatted": "$0.00"
      },
      "with_tax": {
        "amount": 350000,
        "currency": "USD",
        "formatted": "$3,500.00"
      },
      "without_discount": {
        "amount": 350000,
        "currency": "USD",
        "formatted": "$3,500.00"
      },
      "without_tax": {
        "amount": 350000,
        "currency": "USD",
        "formatted": "$3,500.00"
      }
    },
    "timestamps": {
      "created_at": "2022-12-03T15:51:44Z",
      "expires_at": "2022-12-10T15:52:54Z",
      "updated_at": "2022-12-03T15:52:54Z"
    }
  }
}
```

5. We can checkout the cart as follows, using the --auto-fill argument to avoid us having to put in a bunch of data.

```shell
epcc create --auto-fill cart-checkout name=My_Cart 
WARN[0000] (0001) POST https://useast.api.elasticpath.com/v2/carts/3f81126c-723c-4f49-9fee-5fb0466aeb5c/checkout 
{
  "data": {
    "type": "checkout",
    "billing_address": {
      "city": "Santa Ana",
      "company_name": "PIXIA Corp",
      "country": "Georgia",
      "county": "NY",
      "first_name": "Damion",
      "last_name": "Boehm",
      "line_1": "538 Mills ville",
      "postcode": "63825",
      "region": "Hawaii"
    },
    "shipping_address": {
      "city": "Lincoln",
      "company_name": "GitHub",
      "country": "Peru",
      "county": "MO",
      "first_name": "Piper",
      "last_name": "Hessel",
      "line_1": "1527 Wells port",
      "postcode": "70687",
      "region": "IL"
    }
  }
}
WARN[0000] HTTP/2.0 400 Bad Request                     
{
  "errors": [
    {
      "detail": "data.customer or data.contact must be provided",
      "source": "data",
      "status": 400,
      "title": "customer or contact is required"
    }
  ]
}
Error: 400 Bad Request
ERRO[0000] Error occurred while processing command: 400 Bad Request 
```

6. The above shows us that not everything will always be filled in, in some cases you might still get an error, to fix this we must supply a bit more details.

```bash
epcc create --auto-fill cart-checkout name=My_Cart customer.name "John Smith" customer.email "test@test.com"
INFO[0000] (0001) POST https://useast.api.elasticpath.com/v2/carts/3f81126c-723c-4f49-9fee-5fb0466aeb5c/checkout ==> HTTP/2.0 201 Created 
{
  "data": {
    "type": "order",
    "id": "c6702a7f-2c98-45ac-abad-40bd3ab10cbf",
    "anonymized": false,
    "billing_address": {
      "city": "Memphis",
      "company_name": "Equilar",
      "country": "Uganda",
      "county": "CT",
      "first_name": "Lera",
      "last_name": "Herzog",
      "line_1": "2311 Circles port",
      "line_2": "",
      "postcode": "98080",
      "region": "South Dakota"
    },
    "customer": {
      "email": "test@test.com",
      "name": "John Smith"
    },
    "links": {},
    "payment": "unpaid",
    "relationships": {
      "items": {
        "data": [
          {
            "type": "item",
            "id": "700cc607-665e-4bad-871c-6252d384c769"
          }
        ]
      }
    },
    "shipping": "unfulfilled",
    "shipping_address": {
      "city": "Sacramento",
      "company_name": "Scale Unlimited",
      "country": "Niue",
      "county": "CO",
      "first_name": "Immanuel",
      "instructions": "",
      "last_name": "Cronin",
      "line_1": "666 Bridge ton",
      "line_2": "",
      "phone_number": "",
      "postcode": "20835",
      "region": "OH"
    },
    "status": "incomplete",
    "meta": {
      "display_price": {
        "authorized": {
          "amount": 0,
          "currency": "USD",
          "formatted": "$0.00"
        },
        "balance_owing": {
          "amount": 350000,
          "currency": "USD",
          "formatted": "$3,500.00"
        },
        "discount": {
          "amount": 0,
          "currency": "USD",
          "formatted": "$0.00"
        },
        "paid": {
          "amount": 0,
          "currency": "USD",
          "formatted": "$0.00"
        },
        "tax": {
          "amount": 0,
          "currency": "USD",
          "formatted": "$0.00"
        },
        "with_tax": {
          "amount": 350000,
          "currency": "USD",
          "formatted": "$3,500.00"
        },
        "without_discount": {
          "amount": 350000,
          "currency": "USD",
          "formatted": "$3,500.00"
        },
        "without_tax": {
          "amount": 350000,
          "currency": "USD",
          "formatted": "$3,500.00"
        }
      },
      "timestamps": {
        "created_at": "2022-12-03T15:55:50Z",
        "updated_at": "2022-12-03T15:55:50Z"
      }
    }
  },
  "included": {
    "items": [
      {
        "type": "order_item",
        "id": "700cc607-665e-4bad-871c-6252d384c769",
        "catalog_id": "2b3ae101-75cc-4cdd-9537-b2638120cdc5",
        "catalog_source": "pim",
        "links": {},
        "name": "BestEver Gas Range",
        "product_id": "87132b87-ae32-4cbc-a3c8-7875923c7217",
        "quantity": 1,
        "relationships": {
          "cart_item": {
            "data": {
              "type": "cart_item",
              "id": "209749d8-917f-4367-9443-f4b75e77bbaf"
            }
          }
        },
        "sku": "BE-Gas-Range-2b2b",
        "unit_price": {
          "amount": 350000,
          "currency": "USD",
          "includes_tax": false
        },
        "value": {
          "amount": 350000,
          "currency": "USD",
          "includes_tax": false
        },
        "meta": {
          "display_price": {
            "discount": {
              "unit": {
                "amount": 0,
                "currency": "USD",
                "formatted": "$0.00"
              },
              "value": {
                "amount": 0,
                "currency": "USD",
                "formatted": "$0.00"
              }
            },
            "tax": {
              "unit": {
                "amount": 0,
                "currency": "USD",
                "formatted": "$0.00"
              },
              "value": {
                "amount": 0,
                "currency": "USD",
                "formatted": "$0.00"
              }
            },
            "with_tax": {
              "unit": {
                "amount": 350000,
                "currency": "USD",
                "formatted": "$3,500.00"
              },
              "value": {
                "amount": 350000,
                "currency": "USD",
                "formatted": "$3,500.00"
              }
            },
            "without_discount": {
              "unit": {
                "amount": 350000,
                "currency": "USD",
                "formatted": "$3,500.00"
              },
              "value": {
                "amount": 350000,
                "currency": "USD",
                "formatted": "$3,500.00"
              }
            },
            "without_tax": {
              "unit": {
                "amount": 350000,
                "currency": "USD",
                "formatted": "$3,500.00"
              },
              "value": {
                "amount": 350000,
                "currency": "USD",
                "formatted": "$3,500.00"
              }
            }
          },
          "timestamps": {
            "created_at": "2022-12-03T15:55:50Z",
            "updated_at": "2022-12-03T15:55:50Z"
          }
        }
      }
    ]
  }
}
```

### Clean Up

1. To clean up data we need to first switch back to the `client_credentials` option.
```shell
epcc login client_credentials 
INFO[0000] Destroying Customer Token as it should only be used with implicit tokens. 
INFO[0000] POST https://useast.api.elasticpath.com/oauth/access_token ==> HTTP/2.0 200 OK 
INFO[0000] Successfully authenticated with client_credentials, session expires Sat, 03 Dec 2022 08:59:59 -0800 
```
2. To clean up the data in pxm, we can use the following command:
```shell
epcc runbooks run pxm-how-to reset 
INFO[0000] (0001) DELETE https://useast.api.elasticpath.com/pcm/products/sku=shirt ==> HTTP/2.0 204 No Content 
INFO[0000] (0002) DELETE https://useast.api.elasticpath.com/pcm/variations/name=Shirt_Size ==> HTTP/2.0 204 No Content 
INFO[0001] (0003) DELETE https://useast.api.elasticpath.com/pcm/products/87132b87-ae32-4cbc-a3c8-7875923c7217 ==> HTTP/2.0 204 No Content 
WARN[0001] (0004) DELETE https://useast.api.elasticpath.com/v2/customers/name=Leslie_Knope ==> HTTP/2.0 404 Not Found 
WARN[0001] (0005) DELETE https://useast.api.elasticpath.com/v2/customers/name=Ron_Swanson ==> HTTP/2.0 404 Not Found 
INFO[0001] (0006) DELETE https://useast.api.elasticpath.com/pcm/catalogs/name=Ranges_Catalog_for_Special_Customers/releases/name=Ranges_Catalog_for_Special_Customers ==> HTTP/2.0 204 No Content 
INFO[0001] (0007) DELETE https://useast.api.elasticpath.com/pcm/catalogs/2b3ae101-75cc-4cdd-9537-b2638120cdc5/releases/ea868362-8e63-45d4-aa0a-9b7fe42c6ecc ==> HTTP/2.0 204 No Content 
INFO[0001] (0008) DELETE https://useast.api.elasticpath.com/pcm/products/837ab51c-559b-4822-816a-5b173d856a2e ==> HTTP/2.0 204 No Content 
INFO[0001] (0009) DELETE https://useast.api.elasticpath.com/pcm/catalogs/rules/name=Catalog_Rule_for_Civil_Servants ==> HTTP/2.0 204 No Content 
INFO[0002] (0010) DELETE https://useast.api.elasticpath.com/pcm/hierarchies/1777c591-dee5-4a6a-a2d9-72b8af2e582c ==> HTTP/2.0 204 No Content 
INFO[0002] (0011) DELETE https://useast.api.elasticpath.com/pcm/pricebooks/name=Loyal_Civil_Servants_Pricing ==> HTTP/2.0 204 No Content 
INFO[0002] (0012) DELETE https://useast.api.elasticpath.com/v2/currencies/ce16a0ed-8b11-42a6-ab4e-c9f25501bb7b ==> HTTP/2.0 204 No Content 
INFO[0002] (0013) DELETE https://useast.api.elasticpath.com/pcm/pricebooks/cc0eaf11-d4e7-44fd-92e4-d26733548a32 ==> HTTP/2.0 204 No Content 
INFO[0002] (0014) DELETE https://useast.api.elasticpath.com/pcm/catalogs/2b3ae101-75cc-4cdd-9537-b2638120cdc5 ==> HTTP/2.0 204 No Content 
INFO[0002] (0015) DELETE https://useast.api.elasticpath.com/pcm/catalogs/name=Ranges_Catalog_for_Special_Customers ==> HTTP/2.0 204 No Content 
INFO[0002] (0016) DELETE https://useast.api.elasticpath.com/pcm/products/sku=to-kill-a-mockingbird ==> HTTP/2.0 204 No Content 
WARN[0002] (0017) DELETE https://useast.api.elasticpath.com/v2/flows/name=Book ==> HTTP/2.0 404 Not Found 
WARN[0002] (0018) DELETE https://useast.api.elasticpath.com/v2/flows/name=Condition ==> HTTP/2.0 404 Not Found 
INFO[0002] Sleeping for 2 seconds                       
INFO[0004] Total requests 18, and total rate limiting time 5893 ms, and total processing time 6888 ms. Response Code Count: 204:14, 404:4,
```

3. To clean up all the customers we created we can use the following command (**be mindful about running this only in development stores**)
```shell
$epcc delete-all  customers
INFO[0000] Resource customers is a top level resource need to scan only one path to delete all resources 
INFO[0000] (0001) GET https://useast.api.elasticpath.com/v2/customers?page[limit]=25 ==> HTTP/2.0 200 OK 
INFO[0000] Total number of customers in /v2/customers is 1 
INFO[0001] (0002) DELETE https://useast.api.elasticpath.com/v2/customers/ecaee2c4-06db-48a3-a709-6049ee4cb9e2 ==> HTTP/2.0 204 No Content 
INFO[0001] (0003) GET https://useast.api.elasticpath.com/v2/customers?page[limit]=25 ==> HTTP/2.0 200 OK 
INFO[0001] Total ids retrieved for customers in /v2/customers is 0, we are done
```
4. To reset the store in it's entirety (although some things like orders cannot be deleted), you can use the `epcc reset-store <STORE_ID>` like follows:
```shell
$epcc reset-store e15a21a0-cce7-48f3-bfab-5cd265055af1
INFO[0000] (0001) GET https://useast.api.elasticpath.com/v2/settings ==> HTTP/2.0 200 OK 
INFO[0000] (0002) GET https://useast.api.elasticpath.com/v2/settings/customer-authentication ==> HTTP/2.0 200 OK 
INFO[0000] (0003) GET https://useast.api.elasticpath.com/v2/settings/account-authentication ==> HTTP/2.0 200 OK 
INFO[0000] (0004) GET https://useast.api.elasticpath.com/v2/merchant-realm-mappings ==> HTTP/2.0 200 OK 
INFO[0001] (0005) GET https://useast.api.elasticpath.com/v2/authentication-realms ==> HTTP/2.0 200 OK 
INFO[0001] (0006) PUT https://useast.api.elasticpath.com/v2/gateways/adyen ==> HTTP/2.0 200 OK 
...
INFO[0021] Resource v2-products is a top level resource need to scan only one path to delete all resources 
INFO[0022] (0100) GET https://useast.api.elasticpath.com/v2/products?page[limit]=25 ==> HTTP/2.0 200 OK 
INFO[0022] Total ids retrieved for v2-products in /v2/products is 0, we are done 
INFO[0022] Processing resource v2-variations            
INFO[0022] Resource v2-variations is a top level resource need to scan only one path to delete all resources 
INFO[0023] (0101) GET https://useast.api.elasticpath.com/v2/variations?page[limit]=25 ==> HTTP/2.0 200 OK 
INFO[0023] Total ids retrieved for v2-variations in /v2/variations is 0, we are done 
INFO[0023] The following 7 resources were not deleted because we have no way to get a collection: cart-checkout, cart-product-items, entries-relationship, order-payments, order-transaction-capture, order-transaction-refund, pcm-nodes 
INFO[0023] The following 10 resources were not deleted because we have no way to delete an element: authentication-realms, inventory-transactions, log-ttl-settings, merchant-realm-mappings, order-transactions, orders, pcm-child-products, pcm-node-products, pcm-product-templates, personal-data-erasure-requests 
INFO[0023] The following 1 resources were not deleted because they were ignored: application-keys 
INFO[0023] Total requests 101, and total rate limiting time 100 ms, and total processing time 22827 ms. Response Code Count: 200:97, 204:2, 400:2,
```

