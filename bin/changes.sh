#!/bin/bash

file=$1
start_date="2020-01-01"

commits=$(git log --since="$start_date" --pretty=format:"%h" -- "$file")

total_changes=0

for commit in $commits
do
  changes=$(git diff --shortstat "$commit"^ "$commit" -- "$file" | awk '{print $1 + $4}')
  total_changes=$((total_changes + changes))
done

echo "Total changed lines in $file from $start_date to $end_date: $total_changes"

