#!/usr/bin/env bash
set -euo pipefail

OUTFILE="./repo_lists/cf_repos.txt"

# S3
gh-search 'org:Skyscanner "AWS::S3::Bucket"' -l > "$OUTFILE"
# RDS
gh-search 'org:Skyscanner "AWS::RDS::DBInstance"' -l >> "$OUTFILE"
gh-search 'org:Skyscanner "AWS::RDS::DBCluster"' -l >> "$OUTFILE"
# DynamoDB
gh-search 'org:Skyscanner "AWS::DynamoDB::Table"' -l >> "$OUTFILE"

sort -u -o "$OUTFILE" "$OUTFILE"
echo "Saved repo list to $OUTFILE ($(grep -c '^Skyscanner/' "$OUTFILE") repos)"
