#!/bin/bash
set -eo pipefail

# import dashboard template
influx apply --force yes -f /docker-entrypoint-initdb.d/templates/tickers.yaml
