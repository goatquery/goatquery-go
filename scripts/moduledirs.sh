#!/usr/bin/env bash

# Prints all sub-module directories (one per line).
# New modules added under module/ are auto-discovered.

find ./module -type f -name go.mod -exec dirname '{}' \;
