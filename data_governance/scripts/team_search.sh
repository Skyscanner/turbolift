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
  MATCH_COUNT=$(grep -Fxf "$TMP" "$BASEDIR/cf_repos.txt" | sort -u | tee -a "$BASEDIR/team_cf_repos.txt" | wc -l)
  echo "CloudFormation: $MATCH_COUNT repos"
else
  echo "No cf_repos.txt found."
fi

rm -f "$TMP"
