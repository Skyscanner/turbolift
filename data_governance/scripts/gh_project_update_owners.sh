#!/usr/bin/env sh
set -euo pipefail

log(){ printf '%s\n' "[$(date +%H:%M:%S)] $*" >&2; }
err(){ printf '%s\n' "[$(date +%H:%M:%S)] ERROR: $*" >&2; }
need(){ command -v "$1" >/dev/null 2>&1 || { err "missing tool: $1"; exit 1; }; }
b64d(){ if base64 --help 2>&1 | grep -q '\-d'; then base64 -d; else base64 -D; fi; }

# Config
ORG="Skyscanner"
PROJECT_NUMBER=13
TITLE='TurboLift Campaign: Data Governance'
CATALOG_PATH=".catalog.yml"

need gh; need jq; need yq; need awk; need sed; need sort; need tr

# Helpers
fetch_catalog(){ # decode .catalog.yml from repo default branch
  repo="$1"
  ref="$(gh api "repos/$repo" --jq '.default_branch')" || { err "cannot read default_branch for $repo"; return 1; }
  json="$(gh api "repos/$repo/contents/$CATALOG_PATH" --method GET -F ref="$ref")" || { err "GET contents failed for $repo@$ref"; return 1; }
  [ "$(printf %s "$json" | jq -r '.type')" = "file" ] || { err "$CATALOG_PATH is not a file in $repo@$ref"; return 1; }
  [ "$(printf %s "$json" | jq -r '.encoding')" = "base64" ] || { err "unexpected encoding for $repo"; return 1; }
  printf %s "$json" | jq -r '.content' | b64d
}

extract_owner(){ # prefer component.owner, then owner, then spec.owner
  yml="$1"
  printf "%s" "$yml" \
  | yq -o=json '(.component.owner // .owner // .spec.owner // null)' \
  | jq -r '
      if . == null then ""
      elif type == "array" then join(", ")
      elif type == "string" then .
      else "" end
    ' \
  | tr '\n' ',' \
  | sed 's/,,*/,/g; s/^,//; s/,$//'
}

normalize_owner_list(){ # normalize comma-separated list, title-case tokens
  awk -v FS="," -v OFS=", " '
    {
      out="";
      for (i=1;i<=NF;i++){
        s=$i;
        gsub(/^[[:space:]]+|[[:space:]]+$/, "", s);
        gsub(/[[:space:]]+/, " ", s);
        gsub(/"/, "", s);
        gsub(/\011/, "", s);
        gsub(/\015/, "", s);
        if (length(s)>0){
          s = toupper(substr(s,1,1)) substr(s,2)
          if (out=="") out=s; else out=out", "s
        }
      }
      print out
    }' <<EOF
$1
EOF
}

canonicalize_owner_label(){
  in="$1"
  s="$(printf "%s" "$in" \
      | sed -E 's/^[[:space:]]+|[[:space:]]+$//g; s/[[:space:]]+/ /g; s/[“”"]//g' \
      | tr -d '\302\240' \
      | tr -d '\r\t' \
      )"
  key="$(printf "%s" "$s" | tr '[:upper:]' '[:lower:]')"
  case "$key" in
    cassini|\'cassini|fcassini|cassini\ squad|squad\ cassini) echo "Cassini";;
    catalyst|\'catalyst) echo "Catalyst";;
    *) echo "$s";;
  esac
}

get_option_id(){ # resolve option id by normalized name
  fields_json="$1"
  field_name="$2"
  opt_name="$3"
  printf "%s" "$fields_json" \
  | jq -r --arg fn "$field_name" --arg on "$opt_name" '
      .fields[] | select(.name==$fn)
      | (.options // [])[]?
      | . as $o
      | ($o.name // "") as $n
      | ($n
         | tostring
         | gsub("\\u00a0"; " ")
         | gsub("[\\u200b\\u200c\\u200d\\ufeff]"; "")
         | gsub("^[[:space:]]+"; "")
         | gsub("[[:space:]]+$"; "")
         | gsub("[[:space:]]+"; " ")
         | ascii_downcase) as $nn
      | ($on
         | tostring
         | gsub("\\u00a0"; " ")
         | gsub("[\\u200b\\u200c\\u200d\\ufeff]"; "")
         | gsub("^[[:space:]]+"; "")
         | gsub("[[:space:]]+$"; "")
         | gsub("[[:space:]]+"; " ")
         | ascii_downcase) as $target
      | select($nn == $target)
      | .id // empty
    ' \
  | head -n1
}

log "Project $ORG/$PROJECT_NUMBER, Title match: $TITLE"

PROJECT_ID="$(gh project view "$PROJECT_NUMBER" --owner "$ORG" --format json --jq '.id')"
FIELDS_JSON="$(gh project field-list "$PROJECT_NUMBER" --owner "$ORG" --format json)"

EXISTING_OWNERS_ID="$(printf %s "$FIELDS_JSON" | jq -r '.fields[]? | select(.name=="Owners") | .id')"
EXISTING_OWNER_ID="$(printf %s "$FIELDS_JSON"  | jq -r '.fields[]? | select(.name=="Owner")  | .id')"

if [ -n "$EXISTING_OWNERS_ID" ]; then
  OWNER_FIELD_NAME="Owners"
elif [ -n "$EXISTING_OWNER_ID" ]; then
  OWNER_FIELD_NAME="Owner"
else
  OWNER_FIELD_NAME="Owners"
fi

# Build option list by scanning repo catalogs referenced by wanted.txt and current items
: > __owners_all.txt

if [ -f wanted.txt ]; then
  cat wanted.txt > __urls_for_owner_scan.txt
else
  err "wanted.txt not found. Proceeding without it."
  : > __urls_for_owner_scan.txt
fi

if [ -f current.txt ]; then
  cat current.txt >> __urls_for_owner_scan.txt
else
  err "current.txt not found. Fetching current items now."
  gh project item-list "$PROJECT_NUMBER" --owner "$ORG" --format json --limit 5000 \
    --jq '(.items // .) | .[] | .content.url' \
  | sort -u >> __urls_for_owner_scan.txt
fi

sort -u __urls_for_owner_scan.txt \
| sed -E 's#https://github.com/([^/]+/[^/]+)/pull/.*#\1#' \
| sort -u \
| while read -r REPO; do
  [ -n "$REPO" ] || continue
  YAML="$(fetch_catalog "$REPO" 2>/dev/null || true)" || true
  [ -n "$YAML" ] || continue
  OWNER_RAW="$(extract_owner "$YAML")"
  [ -n "$OWNER_RAW" ] || continue
  normalize_owner_list "$OWNER_RAW" \
  | tr ',' '\n' \
  | sed 's/^[[:space:]]\{1,\}//; s/[[:space:]]\{1,\}$//; s/[[:space:]]\{1,\}/ /g; s/["]//g' \
  | tr -d '\r\t' \
  >> __owners_all.txt
done

sort -u __owners_all.txt | sed '/^$/d' > __owners_unique.txt || true
OWNER_OPTIONS_CSV="$(paste -sd, __owners_unique.txt || true)"

# Ensure Owner or Owners field exists with discovered options
OWNER_FIELD_ID="$(
  printf %s "$FIELDS_JSON" \
  | jq -r '.fields[]? | select(.name=="'"$OWNER_FIELD_NAME"'") | .id'
)"

if [ -z "${OWNER_FIELD_ID:-}" ]; then
  if [ -n "${OWNER_OPTIONS_CSV:-}" ]; then
    log "Creating $OWNER_FIELD_NAME with $(wc -l < __owners_unique.txt) options"
    gh project field-create "$PROJECT_NUMBER" --owner "$ORG" \
      --name "$OWNER_FIELD_NAME" --data-type "SINGLE_SELECT" \
      --single-select-options "$OWNER_OPTIONS_CSV" >/dev/null
  else
    gh project field-create "$PROJECT_NUMBER" --owner "$ORG" \
      --name "$OWNER_FIELD_NAME" --data-type "SINGLE_SELECT" \
      --single-select-options "Unknown" >/dev/null
  fi
fi

# Refresh field metadata and compute alternate field
FIELDS_JSON="$(gh project field-list "$PROJECT_NUMBER" --owner "$ORG" --format json)"
OWNER_FIELD_ID="$(printf %s "$FIELDS_JSON" | jq -r '.fields[] | select(.name=="'"$OWNER_FIELD_NAME"'") | .id')"

if [ "$OWNER_FIELD_NAME" = "Owners" ]; then
  ALT_OWNER_FIELD_NAME="Owner"
else
  ALT_OWNER_FIELD_NAME="Owners"
fi
ALT_OWNER_FIELD_ID="$(printf %s "$FIELDS_JSON" | jq -r '.fields[]? | select(.name=="'"$ALT_OWNER_FIELD_NAME"'") | .id')"

# Iterate items and set Owner value
: > __missing_owner_options.txt

gh project item-list "$PROJECT_NUMBER" --owner "$ORG" --format json --limit 5000 \
| jq -r '(.items // .)[] | select(.content.url != null) | [.id, .content.url] | @tsv' \
| while IFS=$'\t' read -r ITEM_ID PR_URL; do
  REPO="$(printf "%s" "$PR_URL" | sed -E 's#https://github.com/([^/]+/[^/]+)/pull/.*#\1#')"
  NUM="$(printf "%s" "$PR_URL" | sed -E 's#.*/pull/([0-9]+).*#\1#')"

  YAML="$(fetch_catalog "$REPO" 2>/dev/null || true)" || true
  [ -n "$YAML" ] || { err "cannot fetch $CATALOG_PATH for $REPO"; continue; }
  OWNER_RAW="$(extract_owner "$YAML")"
  [ -n "$OWNER_RAW" ] || { err "component.owner missing in $CATALOG_PATH for $REPO"; continue; }

  OWNER_VAL="$(normalize_owner_list "$OWNER_RAW")"
  OWNER_ONE="$(printf "%s" "$OWNER_VAL" | awk -F'[,\n]' '{gsub(/^[[:space:]]+|[[:space:]]+$/, "", $1); print $1}')"
  OWNER_ONE_CANON="$(canonicalize_owner_label "$OWNER_ONE")"

  OPT_ID="$(get_option_id "$FIELDS_JSON" "$OWNER_FIELD_NAME" "$OWNER_ONE_CANON")"
  FIELD_ID_TO_USE="$OWNER_FIELD_ID"
  FIELD_NAME_TO_USE="$OWNER_FIELD_NAME"

  if [ -z "$OPT_ID" ] && [ -n "${ALT_OWNER_FIELD_ID:-}" ]; then
    ALT_OPT_ID="$(get_option_id "$FIELDS_JSON" "$ALT_OWNER_FIELD_NAME" "$OWNER_ONE_CANON")"
    if [ -n "$ALT_OPT_ID" ]; then
      OPT_ID="$ALT_OPT_ID"
      FIELD_ID_TO_USE="$ALT_OWNER_FIELD_ID"
      FIELD_NAME_TO_USE="$ALT_OWNER_FIELD_NAME"
    fi
  fi

  if [ -n "$OPT_ID" ]; then
    gh project item-edit --id "$ITEM_ID" --project-id "$PROJECT_ID" \
      --field-id "$FIELD_ID_TO_USE" --single-select-option-id "$OPT_ID" >/dev/null
    log "OK $REPO PR#$NUM owner=$OWNER_ONE_CANON field=$FIELD_NAME_TO_USE"
  else
    err "missing Owner option '$OWNER_ONE_CANON' for $REPO PR#$NUM"
    printf "%s\n" "$OWNER_ONE_CANON" >> __missing_owner_options.txt
  fi
done

if [ -s __missing_owner_options.txt ]; then
  printf "\nOwners missing as options. Add these in the UI, then rerun:\n" >&2
  sort -u __missing_owner_options.txt >&2
fi

log "Owner update complete"
