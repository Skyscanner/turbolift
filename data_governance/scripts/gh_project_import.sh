#!/usr/bin/env sh
set -euo pipefail

log(){ printf '%s\n' "[$(date +%H:%M:%S)] $*" >&2; }
err(){ printf '%s\n' "[$(date +%H:%M:%S)] ERROR: $*" >&2; }
need(){ command -v "$1" >/dev/null 2>&1 || { err "missing tool: $1"; exit 1; }; }

# Constants identical to your original
ORG="Skyscanner"
PROJECT_NUMBER=13
TITLE='TurboLift Campaign: Data Governance'

need gh; need jq

log "Project $ORG/$PROJECT_NUMBER, Title match: $TITLE"

# 1) collect PR URLs by exact title
#    Keep --limit 1000 to preserve behavior
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

log "Import complete"
