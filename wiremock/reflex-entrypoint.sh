#!/bin/bash
cd /home/wiremock

reflex -r '\.json' -s -- sh -c "bash /docker-entrypoint.sh --extensions com.opentable.extension.BodyTransformer --local-response-templating --no-request-journal"
