#!/usr/bin/env bash

set -e
set -x

SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )

echo "Starting Subscriptions Runbook - Clean up subscription resources"
epcc delete-all subscription-offerings
epcc delete-all subscription-subscribers
epcc delete-all subscription-schedules
epcc delete-all subscription-dunning-rules
epcc delete-all subscription-jobs
epcc delete-all subscription-features
epcc delete-all rule-promotions
epcc delete-all account-tags

echo "Starting Subscriptions Runbook - Create Offering"
epcc runbooks run subscriptions create-offering

echo "Starting Subscriptions Runbook - Verify Offering Resources"
# Verify nested resources were created under the offering
NUM_PLANS=$(epcc get subscription-offering-plans name=Streaming_Service --output-jq '.data | length')
if [ "$NUM_PLANS" -eq 1 ]; then
  echo "Correct number of offering plans: 1"
else
  echo "Expected 1 offering plan, but found $NUM_PLANS"
  exit 1
fi

NUM_PRICING_OPTIONS=$(epcc get subscription-offering-pricing-options name=Streaming_Service --output-jq '.data | length')
if [ "$NUM_PRICING_OPTIONS" -eq 1 ]; then
  echo "Correct number of offering pricing options: 1"
else
  echo "Expected 1 offering pricing option, but found $NUM_PRICING_OPTIONS"
  exit 1
fi

NUM_FEATURES=$(epcc get subscription-offering-features name=Streaming_Service --output-jq '.data | length')
if [ "$NUM_FEATURES" -eq 1 ]; then
  echo "Correct number of offering features: 1"
else
  echo "Expected 1 offering feature, but found $NUM_FEATURES"
  exit 1
fi

echo "Starting Subscriptions Runbook - Create Offering Promotion"
epcc runbooks run subscriptions create-offering-promotion

echo "Starting Subscriptions Runbook - Verify Promotion Feature on Offering"
NUM_FEATURES=$(epcc get subscription-offering-features name=Streaming_Service --output-jq '.data | length')
if [ "$NUM_FEATURES" -eq 2 ]; then
  echo "Correct number of offering features: 2 (1 access + 1 promotion)"
else
  echo "Expected 2 offering features, but found $NUM_FEATURES"
  exit 1
fi

echo "Starting Subscriptions Runbook - Create Account and Subscriber"
epcc runbooks run subscriptions create-account-and-subscriber

echo "Starting Subscriptions Runbook - Verify Subscriber"
NUM_SUBSCRIBERS=$(epcc get subscription-subscribers --output-jq '.data | length')
if [ "$NUM_SUBSCRIBERS" -ge 1 ]; then
  echo "Correct: at least 1 subscriber found"
else
  echo "Expected at least 1 subscriber, but found $NUM_SUBSCRIBERS"
  exit 1
fi

echo "Starting Subscriptions Runbook - Setup Billing"
epcc runbooks run subscriptions setup-billing

echo "Starting Subscriptions Runbook - Verify Billing Resources"
NUM_SCHEDULES=$(epcc get subscription-schedules --output-jq '.data | length')
if [ "$NUM_SCHEDULES" -ge 1 ]; then
  echo "Correct: at least 1 schedule found"
else
  echo "Expected at least 1 schedule, but found $NUM_SCHEDULES"
  exit 1
fi

NUM_DUNNING_RULES=$(epcc get subscription-dunning-rules --output-jq '.data | length')
if [ "$NUM_DUNNING_RULES" -ge 1 ]; then
  echo "Correct: at least 1 dunning rule found"
else
  echo "Expected at least 1 dunning rule, but found $NUM_DUNNING_RULES"
  exit 1
fi

echo "Starting Subscriptions Runbook - Run Billing Job"
epcc runbooks run subscriptions run-billing-job

echo "Starting Subscriptions Runbook - View Invoices (listing only, billing job is async)"
epcc get subscription-invoices

echo "Starting Subscriptions Runbook - Bulk Create Offerings"
epcc runbooks run subscriptions create-subscription-offerings --count 3

NUM_OFFERINGS=$(epcc get subscription-offerings --output-jq '.data | length')
if [ "$NUM_OFFERINGS" -ge 4 ]; then
  echo "Correct: at least 4 offerings found (1 named + 3 bulk)"
else
  echo "Expected at least 4 offerings, but found $NUM_OFFERINGS"
  exit 1
fi

echo "Starting Subscriptions Runbook - Reset"
epcc runbooks run subscriptions reset

echo "Subscriptions Runbook - SUCCESS"
