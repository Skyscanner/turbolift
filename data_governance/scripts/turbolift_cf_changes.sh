#!/usr/bin/env bash
set -euo pipefail

turbolift foreach --repos repo_lists/team_cf_repos.txt -- bash -lc '
  set -euo pipefail

  MATCH_YAML="$(mktemp)"
  MATCH_JSON="$(mktemp)"

  # Find YAML files
  rg -l -S "AWS::(S3::Bucket|DynamoDB::Table|DynamoDB::GlobalTable|RDS::DBInstance|RDS::DBCluster)" \
     -g "**/*.yml" -g "**/*.yaml" \
     -g "!**/.git/**" -g "!**/node_modules/**" \
     > "$MATCH_YAML" || true

  # Find JSON files
  rg -l -S "AWS::(S3::Bucket|DynamoDB::Table|DynamoDB::GlobalTable|RDS::DBInstance|RDS::DBCluster)" \
     -g "**/*.json" \
     -g "!**/.git/**" -g "!**/node_modules/**" \
     > "$MATCH_JSON" || true

  # Patch YAML
  while IFS= read -r f || [ -n "${f:-}" ]; do
    [ -n "${f:-}" ] || continue
    ../../../scripts/cf_script.py "$f"
  done < "$MATCH_YAML"

  # Patch JSON
  while IFS= read -r f || [ -n "${f:-}" ]; do
    [ -n "${f:-}" ] || continue
    ../../../scripts/cf_script.py "$f"
  done < "$MATCH_JSON"

  rm -f "$MATCH_YAML" "$MATCH_JSON"
'
