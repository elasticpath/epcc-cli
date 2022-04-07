# Get Started with PCM

This document will walk through the [Get Started with Product Content Management](https://documentation.elasticpath.com/commerce-cloud/docs/developer/how-to/get-started-pcm.html) HOWTO, using the epcc cli.

In this walkthrough, you will.

1. Create a hierarchy with three nodes.
2. Create two products and associate them with nodes in the hierarchy.
3. Create a price book with prices for the products in two currencies
4. Create a catalog that references the hierarchy and price book.
5. Create a catalog rule that describes the context in which to display the catalog.

# Create a hierarchy and nodes

Let's create the following hierarchy and nodes:

* Major Appliances
  * Ranges
    * Electric Ranges
    * Gas Ranges

## Create the hierarchy

Create the hierarchy called Major Appliances. This is also the root node of the hierarchy.

```bash
epcc create pcm-hierarchy name "Major Appliances" description "Free standing appliances" slug "Major-Applications-MA0"
INFO[0000] POST https://useast.api.elasticpath.com/pcm/hierarchies ==> HTTP/2.0 201 Created 
{
  "data": {
    "type": "hierarchy",
    "id": "4012387b-ee24-467a-a7de-78af4d5333c2",
    "attributes": {
      "description": "Free standing appliances",
      "name": "Major Appliances",
      "slug": "Major-Applications-MA0"
    },
    "meta": {
      "created_at": "2022-04-07T21:34:03.083Z",
      "updated_at": "2022-04-07T21:34:03.083Z"
    },
    "relationships": {
      "children": {
        "data": [],
        "links": {
          "related": "/hierarchies/4012387b-ee24-467a-a7de-78af4d5333c2/children"
        }
      }
    }
  }
}
```