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
$epcc create pcm-hierarchy name "Major Appliances" description "Free standing appliances" slug "Major-Applications-MA0"
INFO[0000] POST https://api.moltin.com/pcm/hierarchies ==> HTTP/2.0 201 Created 
{
  "data": {
    "type": "hierarchy",
    "id": "9d752403-1542-486c-a045-f761c3d1a80b",
    "attributes": {
      "description": "Free standing appliances",
      "name": "Major Appliances",
      "slug": "Major-Applications-MA0"
    },
    "meta": {
      "created_at": "2022-04-08T05:18:31.049Z",
      "updated_at": "2022-04-08T05:18:31.049Z"
    },
    "relationships": {
      "children": {
        "data": [],
        "links": {
          "related": "/hierarchies/9d752403-1542-486c-a045-f761c3d1a80b/children"
        }
      }
    }
  }
}
```

## Create the nodes
Three aliases are created for the hierarchy:
1. 'id=X' 
2. 'name=Major_Appliances'
3. 'slug=Major-Applications-MA0'

We can use these in place of the id, and auto-complete them, to generate all three nodes simply run the following:

```bash
$epcc create pcm-node name=Major_Appliances name Ranges description "All stoves and ovens" slug Ranges-MA1
INFO[0000] POST https://api.moltin.com/pcm/hierarchies/9d752403-1542-486c-a045-f761c3d1a80b/nodes ==> HTTP/2.0 201 Created 
{
  "data": {
    "type": "node",
    "id": "c5671fd2-6044-46cb-8db1-78d92e8759ba",
    "attributes": {
      "description": "All stoves and ovens",
      "name": "Ranges",
      "slug": "Ranges-MA1"
    },
    "meta": {
      "created_at": "2022-04-08T05:18:45.39Z",
      "updated_at": "2022-04-08T05:18:45.39Z"
    },
    "relationships": {
      "children": {
        "data": [],
        "links": {
          "related": "/hierarchies/9d752403-1542-486c-a045-f761c3d1a80b/nodes/c5671fd2-6044-46cb-8db1-78d92e8759ba/children"
        }
      },
      "products": {
        "data": [],
        "links": {
          "related": "/hierarchies/9d752403-1542-486c-a045-f761c3d1a80b/nodes/c5671fd2-6044-46cb-8db1-78d92e8759ba/products"
        }
      }
    }
  }
}
```

To create the next node, which is a subnode of the first, we will again use aliases

```bash
epcc create pcm-node name=Major_Appliances name "Electric Ranges" description "Electric stoves and ovens" slug "Electric-Ranges-MA2" relationships.parent.data.type node relationships.parent.data.id slug=Ranges-MA1
INFO[0000] POST https://api.moltin.com/pcm/hierarchies/29e53780-bf98-49cb-bd4a-971f7ccdff29/nodes ==> HTTP/2.0 201 Created 
{
  "data": {
    "type": "node",
    "id": "e6613c2c-d7ea-4c42-ad80-cda810d0e904",
    "attributes": {
      "description": "Electric stoves and ovens",
      "name": "Electric Ranges",
      "slug": "Electric-Ranges-MA2"
    },
    "meta": {
      "created_at": "2022-04-08T05:47:45.597Z",
      "updated_at": "2022-04-08T05:47:45.597Z"
    },
    "relationships": {
      "children": {
        "data": [],
        "links": {
          "related": "/hierarchies/29e53780-bf98-49cb-bd4a-971f7ccdff29/nodes/e6613c2c-d7ea-4c42-ad80-cda810d0e904/children"
        }
      },
      "parent": {
        "data": {
          "type": "node",
          "id": "753b3e79-ca42-4fb3-9ce0-01a5b2b5c7bd"
        }
      },
      "products": {
        "data": [],
        "links": {
          "related": "/hierarchies/29e53780-bf98-49cb-bd4a-971f7ccdff29/nodes/e6613c2c-d7ea-4c42-ad80-cda810d0e904/products"
        }
      }
    }
  }
}
```

and the third

```bash
epcc create pcm-node name=Major_Appliances name "Gas Ranges" description "Gas stoves and ovens" slug "Gas-Ranges-MA2" relationships.parent.data.type node relationships.parent.data.id slug=Ranges-MA1
INFO[0000] POST https://api.moltin.com/pcm/hierarchies/29e53780-bf98-49cb-bd4a-971f7ccdff29/nodes ==> HTTP/2.0 201 Created 
{
  "data": {
    "type": "node",
    "id": "a661a313-ef67-485d-8133-f853fe278006",
    "attributes": {
      "description": "Gas stoves and ovens",
      "name": "Gas Ranges",
      "slug": "Gas-Ranges-MA2"
    },
    "meta": {
      "created_at": "2022-04-08T05:48:34.732Z",
      "updated_at": "2022-04-08T05:48:34.732Z"
    },
    "relationships": {
      "children": {
        "data": [],
        "links": {
          "related": "/hierarchies/29e53780-bf98-49cb-bd4a-971f7ccdff29/nodes/a661a313-ef67-485d-8133-f853fe278006/children"
        }
      },
      "parent": {
        "data": {
          "type": "node",
          "id": "753b3e79-ca42-4fb3-9ce0-01a5b2b5c7bd"
        }
      },
      "products": {
        "data": [],
        "links": {
          "related": "/hierarchies/29e53780-bf98-49cb-bd4a-971f7ccdff29/nodes/a661a313-ef67-485d-8133-f853fe278006/products"
        }
      }
    }
  }
}
```

## Create the products and associate them with the hierarchy nodes

```bash
epcc create pcm-product name "BestEver Electric Range" sku "BE-Electric-Range-1a1a" slug "bestever-range-1a1a" description "This electric model offers an induction heating element and convection oven." status live commodity_type physical upc_ean \"111122223333\" mpn BE-R-1111-aaaa-1a1a
INFO[0000] POST https://api.moltin.com/pcm/products ==> HTTP/2.0 201 Created 
{
  "data": {
    "type": "product",
    "id": "8dc8c7d2-4225-4791-b8dc-c007a45d7aff",
    "attributes": {
      "commodity_type": "physical",
      "description": "This electric model offers an induction heating element and convection oven.",
      "mpn": "BE-R-1111-aaaa-1a1a",
      "name": "BestEver Electric Range",
      "sku": "BE-Electric-Range-1a1a",
      "slug": "bestever-range-1a1a",
      "status": "live",
      "upc_ean": "111122223333"
    },
    "meta": {
      "created_at": "2022-04-08T05:51:31.03Z",
      "updated_at": "2022-04-08T05:51:31.03Z"
    },
    "relationships": {
      "children": {
        "data": [],
        "links": {
          "self": "/products/8dc8c7d2-4225-4791-b8dc-c007a45d7aff/children"
        }
      },
      "component_products": {
        "data": [],
        "links": {
          "self": "/products/8dc8c7d2-4225-4791-b8dc-c007a45d7aff/relationships/component_products"
        }
      },
      "files": {
        "data": [],
        "links": {
          "self": "/products/8dc8c7d2-4225-4791-b8dc-c007a45d7aff/relationships/files"
        }
      },
      "main_image": {
        "data": null
      },
      "templates": {
        "data": [],
        "links": {
          "self": "/products/8dc8c7d2-4225-4791-b8dc-c007a45d7aff/relationships/templates"
        }
      },
      "variations": {
        "data": [],
        "links": {
          "self": "/products/8dc8c7d2-4225-4791-b8dc-c007a45d7aff/relationships/variations"
        }
      }
    }
  }
}
```