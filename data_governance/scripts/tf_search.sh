#!/usr/bin/env bash
set -euo pipefail

OUTFILE="./repo_lists/tf_repos.txt"

# S3
gh-search 'org:Skyscanner "resource \"aws_s3_bucket\""' -l > "$OUTFILE"
# RDS
gh-search 'org:Skyscanner "resource \"aws_db_instance\""' -l >> "$OUTFILE"
gh-search 'org:Skyscanner "resource \"aws_rds_cluster\""' -l >> "$OUTFILE"
# DynamoDB
gh-search 'org:Skyscanner "resource \"aws_dynamodb_table\""' -l >> "$OUTFILE"

sort -u -o "$OUTFILE" "$OUTFILE"
echo "Saved repo list to $OUTFILE ($(grep -c '^Skyscanner/' "$OUTFILE") repos)"
