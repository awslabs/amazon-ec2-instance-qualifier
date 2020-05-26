#!/usr/bin/env bash

set -euo pipefail

MEM_LOAD_OUTPUT=$1
MONITOR_PERIOD=$2

while true; do
  sleep "$MONITOR_PERIOD"
  mem_used=$(free -m | grep Mem | awk '{print $3}')
  echo "$mem_used" >> "$MEM_LOAD_OUTPUT"
done
