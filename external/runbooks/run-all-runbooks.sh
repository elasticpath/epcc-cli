#!/usr/bin/env bash

SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )

set -e
set -x

epcc reset-store .+

echo "Starting Misc Runbook"
epcc runbooks run misc get-store-info

echo "Starting Hello World"

epcc runbooks run hello-world create-customer
epcc runbooks run hello-world create-10-customers

epcc create customer --auto-fill
epcc runbooks run hello-world create-some-customer-addresses --customer_id last_read=entity

epcc runbooks run hello-world reset

echo "Starting Extend Customer Resources"
# We don't reset here, because we shouldn't need to

epcc runbooks run extend-customer-resources create-flow-and-field
epcc runbooks run extend-customer-resources create-example-customer
epcc runbooks run extend-customer-resources update-example-customer
epcc runbooks run extend-customer-resources reset


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


echo "Starting manual gateway how-to"

echo "SUCCESS"



