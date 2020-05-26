#!/usr/bin/env bash

set -euo pipefail

CPU_LOAD_OUTPUT=$1
MONITOR_PERIOD=$2

while true; do
  sleep "$MONITOR_PERIOD"
  cpu_load=$(uptime | awk '{print $(NF-2)}' | sed 's/,//g')
  echo "$cpu_load" >> "$CPU_LOAD_OUTPUT"
done