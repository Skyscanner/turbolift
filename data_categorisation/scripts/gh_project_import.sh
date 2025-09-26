#!/usr/bin/env sh
set -euo pipefail

log(){ printf '%s\n' "[$(date +%H:%M:%S)] $*" >&2; }
err(){ printf '%s\n' "[$(date +%H:%M:%S)] ERROR: $*" >&2; }

# --- config ---
ORG="Skyscanner"
PROJECT_NUMBER=13
TITLE='TurboLift Campaign: Data Governance'
STATE_FIELD_NAME="State"
OWNER_FIELD_NAME="Owner"
CATALOG_PATH=".catalog.yml"   # file is always at repo root

# --- tools ---
need(){ command -v "$1" >/dev/null 2>&1 || { err "missing tool: $1"; exit 1; }; }
need gh; need jq; need yq
b64d(){ if base64 --help 2>&1 | grep -q '\-d'; then base64 -d; else base64 -D; fi; }

# --- helpers ---
fetch_catalog(){ # decode .catalog.yml from repo default branch
  repo="$1"
  ref="$(gh api "repos/$repo" --jq '.default_branch')" || { err "cannot read default_branch for $repo"; return 1; }
  json="$(gh api "repos/$repo/contents/$CATALOG_PATH" --method GET -F ref="$ref")" || { err "GET contents failed for $repo@$ref"; return 1; }
  [ "$(printf %s "$json" | jq -r '.type')" = "file" ] || { err "$CATALOG_PATH is not a file in $repo@$ref"; return 1; }
  [ "$(printf %s "$json" | jq -r '.encoding')" = "base64" ] || { err "unexpected encoding for $repo"; return 1; }
  printf %s "$json" | jq -r '.content' | b64d
}

extract_owner(){ # string or array -> string
  yml="$1"
  printf "%s" "$yml" \
  | yq -o=json '.component.owner // null' \
  | jq -r '
      if . == null then ""
      elif type == "array" then join(", ")
      elif type == "string" then .
      else "" end
    ' \
  | sed 's/^ *//; s/ *$//'
}

normalize_owner(){ # capitalize first letter of each comma-separated entry
  awk -v FS="," -v OFS=", " '
    {
      for (i=1;i<=NF;i++){
        gsub(/^ +| +$/, "", $i);
        if (length($i)>0){
          $i = toupper(substr($i,1,1)) substr($i,2)
        }
      }
      print
    }' <<EOF
$1
EOF
}

read_owner_field(){ # safe verify: use inline fragment for union
  item_id="$1"
  out="$(gh api graphql -f query='
    query($i:ID!){
      node(id:$i){
        ... on ProjectV2Item{
          fieldValues(first:100){
            nodes{
              __typename
              ... on ProjectV2ItemFieldTextValue{
                text
                field{
                  ... on ProjectV2FieldCommon { id name }
                }
              }
            }
          }
        }
      }
    }' -f i="$item_id" 2>__gql.err || true)"
  if [ -z "${out:-}" ]; then
    err "GraphQL verify failed: $(head -n1 __gql.err 2>/dev/null)"
    printf ""
    return 0
  fi
  printf "%s" "$out" | jq -r --arg fid "$OWNER_FIELD_ID" '
    .data.node.fieldValues.nodes[]
    | select(.field.id == $fid)
    | (.text // "")
  '
}

# --- start ---
log "Project $ORG/$PROJECT_NUMBER, Title match: $TITLE"

# 1) collect PR URLs by exact title
gh search prs --owner "$ORG" --match title "$TITLE" \
  --json url,title --limit 1000 \
  --jq '.[] | select(.title == "'"$TITLE"'") | .url' \
| sort -u > wanted.txt
log "Wanted PRs: $(wc -l < wanted.txt)"

# 2) current items in project
gh project item-list "$PROJECT_NUMBER" --owner "$ORG" --format json --limit 5000 \
  --jq '(.items // .) | .[] | .content.url' \
| sort -u > current.txt
log "Existing project items: $(wc -l < current.txt)"

# 3) add missing
comm -13 current.txt wanted.txt | while read -r url; do
  [ -n "$url" ] || continue
  gh project item-add "$PROJECT_NUMBER" --owner "$ORG" --url "$url" >/dev/null
done

# 4) ensure fields
STATE_FIELD_ID="$(
  gh project field-list "$PROJECT_NUMBER" --owner "$ORG" --format json \
  --jq '.fields[]? | select(.name=="'"$STATE_FIELD_NAME"'") | .id'
)"
[ -n "${STATE_FIELD_ID:-}" ] || gh project field-create "$PROJECT_NUMBER" --owner "$ORG" \
  --name "$STATE_FIELD_NAME" --data-type "SINGLE_SELECT" \
  --single-select-options "Open,Merged,Closed" >/dev/null

OWNER_FIELD_ID="$(
  gh project field-list "$PROJECT_NUMBER" --owner "$ORG" --format json \
  --jq '.fields[]? | select(.name=="'"$OWNER_FIELD_NAME"'") | .id'
)"
[ -n "${OWNER_FIELD_ID:-}" ] || gh project field-create "$PROJECT_NUMBER" --owner "$ORG" \
  --name "$OWNER_FIELD_NAME" --data-type "TEXT" >/dev/null

PROJECT_ID="$(gh project view "$PROJECT_NUMBER" --owner "$ORG" --format json --jq '.id')"
FIELDS_JSON="$(gh project field-list "$PROJECT_NUMBER" --owner "$ORG" --format json)"
STATE_FIELD_ID="$(printf %s "$FIELDS_JSON" | jq -r '.fields[] | select(.name=="'"$STATE_FIELD_NAME"'") | .id')"
OWNER_FIELD_ID="$(printf %s "$FIELDS_JSON" | jq -r '.fields[] | select(.name=="'"$OWNER_FIELD_NAME"'") | .id')"
OPT_OPEN="$(printf %s "$FIELDS_JSON"   | jq -r '.fields[]|select(.name=="'"$STATE_FIELD_NAME"'")|.options[]|select(.name=="Open").id')"
OPT_MERGED="$(printf %s "$FIELDS_JSON" | jq -r '.fields[]|select(.name=="'"$STATE_FIELD_NAME"'")|.options[]|select(.name=="Merged").id')"
OPT_CLOSED="$(printf %s "$FIELDS_JSON" | jq -r '.fields[]|select(.name=="'"$STATE_FIELD_NAME"'")|.options[]|select(.name=="Closed").id')"

# 5) iterate items
gh project item-list "$PROJECT_NUMBER" --owner "$ORG" --format json --limit 5000 \
| jq -r '(.items // .)[] | select(.content.url != null) | [.id, .content.url] | @tsv' \
| while IFS=$'\t' read -r ITEM_ID PR_URL; do
  REPO="$(printf "%s" "$PR_URL" | sed -E 's#https://github.com/([^/]+/[^/]+)/pull/.*#\1#')"
  NUM="$(printf "%s" "$PR_URL" | sed -E 's#.*/pull/([0-9]+).*#\1#')"

  # state
  line="$(gh pr view "$NUM" -R "$REPO" --json state,mergedAt --jq '[.state, .mergedAt] | @tsv' 2>/dev/null || true)"
  if [ -z "$line" ]; then
    state="$(gh api /search/issues -f q="repo:$REPO is:pr number:$NUM" --jq '.items[0].state' 2>/dev/null || echo "CLOSED")"
    line="$state\tnull"
  fi
  state="$(printf "%s" "$line" | cut -f1)"
  merged_at="$(printf "%s" "$line" | cut -f2)"
  if [ "$state" = "OPEN" ]; then opt="$OPT_OPEN"
  elif [ "$merged_at" != "null" ] && [ -n "$merged_at" ]; then opt="$OPT_MERGED"
  else opt="$OPT_CLOSED"; fi
  gh project item-edit --id "$ITEM_ID" --project-id "$PROJECT_ID" \
    --field-id "$STATE_FIELD_ID" --single-select-option-id "$opt" >/dev/null

  # owner
  YAML="$(fetch_catalog "$REPO")" || { err "cannot fetch $CATALOG_PATH for $REPO"; continue; }
  OWNER_RAW="$(extract_owner "$YAML")"
  [ -n "$OWNER_RAW" ] || { err "component.owner missing in $CATALOG_PATH for $REPO"; continue; }
  OWNER_VAL="$(normalize_owner "$OWNER_RAW")"
  log "Owner=$OWNER_VAL ($REPO)"

  # write and verify
  gh project item-edit --id "$ITEM_ID" --project-id "$PROJECT_ID" \
    --field-id "$OWNER_FIELD_ID" --text "$OWNER_VAL" >/dev/null
  saved="$(read_owner_field "$ITEM_ID" || true)"
  if [ "$saved" = "$OWNER_VAL" ]; then
    log "OK $REPO PR#$NUM owner set"
  else
    err "verify failed for $REPO PR#$NUM got='$saved' expected='$OWNER_VAL'"
  fi
done

log "Done"
