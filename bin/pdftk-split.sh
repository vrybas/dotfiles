#!/bin/bash

# The input PDF file
input=$1

# Get the total number of pages
pages=$(pdftk "$input" dump_data | grep NumberOfPages | awk '{print $2}')

# Define the page range step
step=10

# Split PDF into 10-page chunks
for ((i=1; i<=$pages; i+=step)); do
  j=$((i + step - 1))
  [[ $j -gt $pages ]] && j=$pages  # Adjust last chunk if page range exceeds total pages
  output=$(printf "output_%02d-%02d.pdf" $i $j)
  pdftk "$input" cat $i-$j output "$output"
done
