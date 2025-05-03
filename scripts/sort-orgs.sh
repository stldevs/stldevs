#!/bin/bash

# This script ensures the aggregator/orgs.txt list is sorted

# Usage: ./scripts/sort-orgs.sh

# Check if the script is being run from the root of the repository
if [ ! -f "scripts/sort-orgs.sh" ]; then
  echo "Please run this script from the root of the repository."
  exit 1
fi

if [ ! -f "aggregator/orgs.txt" ]; then
  echo "The file aggregator/orgs.txt does not exist."
  exit 1
fi

sort -o aggregator/orgs.txt aggregator/orgs.txt
