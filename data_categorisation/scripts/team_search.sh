#!/usr/bin/env bash
set -euo pipefail

if [ $# -lt 1 ]; then
  echo "Usage: $0 <team-name>"
  exit 1
fi

TEAM="$1"
BASEDIR="./repo_lists"

TMP="$(mktemp)"
gh-search "org:Skyscanner filename:catalog.yml $TEAM" -l \
  | grep -E '^Skyscanner/[[:alnum:]_.-]+$' > "$TMP"

if [ -f "$BASEDIR/cf_repos.txt" ]; then
  grep -Fxf "$TMP" "$BASEDIR/cf_repos.txt" | sort -u > "$BASEDIR/team_cf_repos.txt" || true
  echo "CloudFormation: $(grep -c '^Skyscanner/' "$BASEDIR/team_cf_repos.txt" || echo 0) repos"
fi

if [ -f "$BASEDIR/tf_repos.txt" ]; then
  grep -Fxf "$TMP" "$BASEDIR/tf_repos.txt" | sort -u > "$BASEDIR/team_tf_repos.txt" || true
  echo "Terraform: $(grep -c '^Skyscanner/' "$BASEDIR/team_tf_repos.txt" || echo 0) repos"
fi

rm -f "$TMP"
