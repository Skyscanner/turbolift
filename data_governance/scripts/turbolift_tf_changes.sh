#!/usr/bin/env bash
set -euo pipefail

# Adjust the repo list path if you need to
turbolift foreach --repos repo_lists/team_tf_repos.txt -- bash -lc '
  set -euo pipefail

  MATCH_TF="$(mktemp)"
  MATCH_TFJSON="$(mktemp)"

  # Find HCL (.tf) resources
  rg -l -S "resource\\s+\\\"(aws_s3_bucket|aws_dynamodb_table|aws_db_instance|aws_rds_cluster)\\\"" \
     -g "**/*.tf" \
     -g "!**/.git/**" -g "!**/node_modules/**" \
     > "$MATCH_TF" || true

  # Find JSON-style Terraform (.tf.json)
  rg -l -S "\"(aws_s3_bucket|aws_dynamodb_table|aws_db_instance|aws_rds_cluster)\"" \
     -g "**/*.tf.json" \
     -g "!**/.git/**" -g "!**/node_modules/**" \
     > "$MATCH_TFJSON" || true

  # Patch .tf
  while IFS= read -r f || [ -n "${f:-}" ]; do
    [ -n "${f:-}" ] || continue
    ../../../scripts/tf_tags.py "$f"
  done < "$MATCH_TF"

  # Patch .tf.json
  while IFS= read -r f || [ -n "${f:-}" ]; do
    [ -n "${f:-}" ] || continue
    ../../../scripts/tf_tags.py "$f"
  done < "$MATCH_TFJSON"

  rm -f "$MATCH_TF" "$MATCH_TFJSON"

  # Optional: print whether the repo changed
  if git diff --quiet; then
    echo "No changes in this repo"
  else
    echo "CHANGED in this repo"
    # uncomment if you want auto-commits:
    # git add -A && git commit -m "Add data classification tags to Terraform resources"
  fi
'
