#!/usr/bin/env sh
set -euo pipefail

log(){ printf '%s\n' "[$(date +%H:%M:%S)] $*" >&2; }
err(){ printf '%s\n' "[$(date +%H:%M:%S)] ERROR: $*" >&2; }
need(){ command -v "$1" >/dev/null 2>&1 || { err "missing tool: $1"; exit 1; }; }

# Config
ORG="Skyscanner"
PROJECT_NUMBER=13
TITLE='TurboLift Campaign: Data Governance'
STATE_FIELD_NAME="State"

need gh; need jq

log "Project $ORG/$PROJECT_NUMBER, Title match: $TITLE"

# Ensure State field exists with required options
STATE_FIELD_ID="$(
  gh project field-list "$PROJECT_NUMBER" --owner "$ORG" --format json \
  --jq '.fields[]? | select(.name=="'"$STATE_FIELD_NAME"'") | .id'
)"

[ -n "${STATE_FIELD_ID:-}" ] || gh project field-create "$PROJECT_NUMBER" --owner "$ORG" \
  --name "$STATE_FIELD_NAME" --data-type "SINGLE_SELECT" \
  --single-select-options "Open,Merged,Closed" >/dev/null

PROJECT_ID="$(gh project view "$PROJECT_NUMBER" --owner "$ORG" --format json --jq '.id')"
FIELDS_JSON="$(gh project field-list "$PROJECT_NUMBER" --owner "$ORG" --format json)"

STATE_FIELD_ID="$(printf %s "$FIELDS_JSON" | jq -r '.fields[] | select(.name=="'"$STATE_FIELD_NAME"'") | .id')"
OPT_OPEN="$(printf %s "$FIELDS_JSON"   | jq -r '.fields[]|select(.name=="'"$STATE_FIELD_NAME"'")|.options[]|select(.name=="Open").id')"
OPT_MERGED="$(printf %s "$FIELDS_JSON" | jq -r '.fields[]|select(.name=="'"$STATE_FIELD_NAME"'")|.options[]|select(.name=="Merged").id')"
OPT_CLOSED="$(printf %s "$FIELDS_JSON" | jq -r '.fields[]|select(.name=="'"$STATE_FIELD_NAME"'")|.options[]|select(.name=="Closed").id')"

# Iterate items and update State
gh project item-list "$PROJECT_NUMBER" --owner "$ORG" --format json --limit 5000 \
| jq -r '(.items // .)[] | select(.content.url != null) | [.id, .content.url] | @tsv' \
| while IFS=$'\t' read -r ITEM_ID PR_URL; do
  REPO="$(printf "%s" "$PR_URL" | sed -E 's#https://github.com/([^/]+/[^/]+)/pull/.*#\1#')"
  NUM="$(printf "%s" "$PR_URL" | sed -E 's#.*/pull/([0-9]+).*#\1#')"

  line="$(gh pr view "$NUM" -R "$REPO" --json state,mergedAt --jq '[.state, .mergedAt] | @tsv' 2>/dev/null || true)"
  if [ -z "$line" ]; then
    state="$(gh api /search/issues -f q="repo:$REPO is:pr number:$NUM" --jq '.items[0].state' 2>/dev/null || echo "CLOSED")"
    line="$state\tnull"
  fi

  state="$(printf "%s" "$line" | cut -f1)"
  merged_at="$(printf "%s" "$line" | cut -f2)"

  if [ "$state" = "OPEN" ]; then
    opt="$OPT_OPEN"
  elif [ "$merged_at" != "null" ] && [ -n "$merged_at" ]; then
    opt="$OPT_MERGED"
  else
    opt="$OPT_CLOSED"
  fi

  gh project item-edit --id "$ITEM_ID" --project-id "$PROJECT_ID" \
    --field-id "$STATE_FIELD_ID" --single-select-option-id "$opt" >/dev/null

  log "State set for $REPO PR#$NUM -> $( [ "$opt" = "$OPT_OPEN" ] && echo Open || { [ "$opt" = "$OPT_MERGED" ] && echo Merged || echo Closed; } )"
done

log "State update complete"
