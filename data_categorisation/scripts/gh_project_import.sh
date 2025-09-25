#!/usr/bin/env sh
set -euo pipefail

ORG="Skyscanner"
PROJECT_NUMBER=13
TITLE='TurboLift Campaign: Data Governance'
FIELD_NAME="State"

echo "==> Project $ORG/$PROJECT_NUMBER, Title match: $TITLE"

# 1) Find PRs across the org, keep ONLY exact-title matches
gh search prs --owner "$ORG" --match title "$TITLE" \
  --json url,title --limit 1000 \
| jq -r --arg T "$TITLE" '.[] | select(.title == $T) | .url' \
| sort -u > wanted.txt

echo "==> Wanted PRs: $(wc -l < wanted.txt)"

# 2) Current items in the Project (avoid duplicates)
gh project item-list "$PROJECT_NUMBER" --owner "$ORG" --format json --limit 5000 \
| jq -r '(.items // .) | .[] | .content.url' \
| sort -u > current.txt

echo "==> Existing project items: $(wc -l < current.txt)"

# 3) Add only the missing ones
comm -13 current.txt wanted.txt | while read -r url; do
  [ -n "$url" ] || continue
  echo " + Adding: $url"
  gh project item-add "$PROJECT_NUMBER" --owner "$ORG" --url "$url" >/dev/null
done

# 4) Ensure single-select field "PR state" exists with options Open,Merged,Closed
have_field="$(
  gh project field-list "$PROJECT_NUMBER" --owner "$ORG" --format json \
  | jq -r ".fields[]? | select(.name==\"$FIELD_NAME\") | .id"
)"
if [ -z "${have_field:-}" ]; then
  gh project field-create "$PROJECT_NUMBER" --owner "$ORG" \
    --name "$FIELD_NAME" --data-type "SINGLE_SELECT" \
    --single-select-options "Open,Merged,Closed" >/dev/null
fi

# 5) Resolve project id, field id, and option ids (raw IDs)
PROJECT_ID="$(gh project view "$PROJECT_NUMBER" --owner "$ORG" --format json | jq -r '.id')"
FIELDS_JSON="$(gh project field-list "$PROJECT_NUMBER" --owner "$ORG" --format json)"
FIELD_ID="$(printf %s "$FIELDS_JSON" | jq -r ".fields[] | select(.name==\"$FIELD_NAME\") | .id")"
OPT_OPEN="$(printf %s "$FIELDS_JSON" | jq -r ".fields[]|select(.name==\"$FIELD_NAME\")|.options[]|select(.name==\"Open\").id")"
OPT_MERGED="$(printf %s "$FIELDS_JSON" | jq -r ".fields[]|select(.name==\"$FIELD_NAME\")|.options[]|select(.name==\"Merged\").id")"
OPT_CLOSED="$(printf %s "$FIELDS_JSON" | jq -r ".fields[]|select(.name==\"$FIELD_NAME\")|.options[]|select(.name==\"Closed\").id")"

echo "==> Field IDs ok. Project: $PROJECT_ID Field: $FIELD_ID"

# 6) For every PR item, set PR state = Open|Merged|Closed
gh project item-list "$PROJECT_NUMBER" --owner "$ORG" --format json --limit 5000 \
| jq -r '(.items // .)[] | select(.content.url != null) | [.id, .content.url] | @tsv' \
| while IFS=$'\t' read -r ITEM_ID PR_URL; do
    REPO="$(printf "%s" "$PR_URL" | sed -E 's#https://github.com/([^/]+/[^/]+)/pull/.*#\1#')"
    NUM="$(printf "%s" "$PR_URL" | sed -E 's#.*/pull/([0-9]+).*#\1#')"

    # Prefer detailed state; fallback to Search API for open/closed only
    LINE="$(gh pr view "$NUM" -R "$REPO" --json state,mergedAt --jq '[.state, .mergedAt] | @tsv' 2>/dev/null || true)"
    if [ -z "$LINE" ]; then
      STATE="$(gh api /search/issues -f q="repo:$REPO is:pr number:$NUM" --jq '.items[0].state' 2>/dev/null || echo "CLOSED")"
      LINE="$STATE\tnull"
    fi

    STATE="$(printf "%s" "$LINE" | cut -f1)"
    MERGED_AT="$(printf "%s" "$LINE" | cut -f2)"

    if [ "$STATE" = "OPEN" ]; then
      OPT="$OPT_OPEN"
    elif [ "$MERGED_AT" != "null" ] && [ -n "$MERGED_AT" ]; then
      OPT="$OPT_MERGED"
    else
      OPT="$OPT_CLOSED"
    fi

    gh project item-edit --id "$ITEM_ID" --project-id "$PROJECT_ID" \
      --field-id "$FIELD_ID" --single-select-option-id "$OPT" >/dev/null \
    || gh api graphql -f query='
        mutation($p:ID!,$i:ID!,$f:ID!,$o:ID!){
          updateProjectV2ItemFieldValue(input:{projectId:$p,itemId:$i,fieldId:$f,value:{singleSelectOptionId:$o}}){
            projectV2Item{ id }
          }
        }' -f p="$PROJECT_ID" -f i="$ITEM_ID" -f f="$FIELD_ID" -f o="$OPT" >/dev/null
done
