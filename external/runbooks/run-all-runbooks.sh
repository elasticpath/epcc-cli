#!/usr/bin/env bash

SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )

# We just need to hack this for now, due to API limitations
epcc delete cart no_cust_cart

set -e
set -x



#Let's test that epcc command works after an embarrassing bug that caused it to panic :(
epcc

echo "Starting Rule Promotions Runbook"
epcc reset-store .+
epcc runbooks run rule-promotions-how-to create-prequisites
epcc runbooks run rule-promotions-how-to create-rule-promotions
epcc runbooks run rule-promotions-how-to create-cart-and-add-ranges
epcc runbooks run rule-promotions-how-to reset

echo "Starting Multi Location Inventory Runbook"
epcc reset-store .+
epcc headers set ep-inventories-multi-location true
epcc runbooks run multi-location-inventory create-prerequisites
epcc runbooks run multi-location-inventory setup-inventory
epcc runbooks run multi-location-inventory check-inventory
epcc runbooks run multi-location-inventory create-order
epcc runbooks run multi-location-inventory check-inventory
epcc runbooks run multi-location-inventory authorize-payment
epcc runbooks run multi-location-inventory capture-payment
epcc runbooks run multi-location-inventory reset-inventory
epcc runbooks run multi-location-inventory reset-cart
epcc runbooks run multi-location-inventory reset
epcc headers delete ep-inventories-multi-location

echo "Starting Commerce Extensions Runbook"
epcc reset-store .+

epcc runbooks run commerce-extensions create-product-review-custom-api
epcc runbooks run commerce-extensions create-review-for-random-product
epcc runbooks run commerce-extensions delete-product-review-custom-api

echo "Starting Subscriptions Tests"
epcc reset-store .+
epcc runbooks run subscriptions create-subscription-offerings

echo "Starting Account Cart Association Tests"
epcc reset-store .+

epcc runbooks run account-cart-associations try-and-delete-all-carts
epcc runbooks run account-cart-associations create-prerequisites
epcc runbooks run account-cart-associations create-accounts-and-carts-with-product-items
epcc runbooks run account-cart-associations delete-accounts-and-carts
epcc runbooks run account-cart-associations create-accounts-and-carts-with-custom-items
epcc runbooks run account-cart-associations delete-accounts-and-carts
epcc runbooks run account-cart-associations reset




echo "Starting Misc Runbook"
epcc reset-store .+
epcc runbooks run misc get-store-info

echo "Starting pxm how to"
epcc reset-store .+

epcc runbooks run pxm-how-to create-catalog-and-publish
epcc runbooks run pxm-how-to create-catalog-rule
epcc runbooks run pxm-how-to products-with-custom-data
epcc runbooks run pxm-how-to products-with-variations
epcc runbooks run pxm-how-to publish-catalog-with-bundles
epcc runbooks run pxm-how-to reset


echo "Starting Hello World"
epcc reset-store .+
epcc runbooks run hello-world create-customer
epcc runbooks run hello-world create-10-customers

epcc create customer --auto-fill
epcc runbooks run hello-world create-some-customer-addresses --customer-id last_read=entity

epcc runbooks run hello-world concurrent-sleeps  --count 2
epcc runbooks run hello-world sequential-sleeps  --count 2

epcc runbooks run hello-world reset

echo "Starting Extend Customer Resources"
# We don't reset here, because we shouldn't need to

epcc runbooks run extend-customer-resources-how-to create-flow-and-field
epcc runbooks run extend-customer-resources-how-to create-example-customer
epcc runbooks run extend-customer-resources-how-to update-example-customer
epcc runbooks run extend-customer-resources-how-to reset


echo "Starting Account Management Runbook"

epcc reset-store .+
epcc runbooks run account-management enable-password-authentication
epcc runbooks run account-management create-singleton-account-member
epcc runbooks run account-management catalog-rule-example
epcc runbooks run account-management catalog-rule-example-reset


echo "Starting manual gateway how-to"
epcc reset-store .+
epcc runbooks run manual-gateway-how-to create-prerequisites
epcc runbooks run manual-gateway-how-to create-order
epcc runbooks run manual-gateway-how-to authorize-payment
epcc runbooks run manual-gateway-how-to capture-payment
epcc runbooks run manual-gateway-how-to reset-cart
epcc runbooks run manual-gateway-how-to reset

echo "Starting Customer Cart Association Tests"
epcc reset-store .+

epcc runbooks run customer-cart-associations try-and-delete-all-carts
epcc runbooks run customer-cart-associations create-prerequisites
epcc runbooks run customer-cart-associations create-customers-and-carts-with-product-items
epcc runbooks run customer-cart-associations delete-customer-and-carts-with-product-items
epcc runbooks run customer-cart-associations create-customers-and-carts-with-custom-items
epcc runbooks run customer-cart-associations delete-customer-and-carts-with-custom-items
epcc runbooks run customer-cart-associations reset


echo "SUCCESS"



