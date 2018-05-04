#!/bin/bash
BUILD=$1
VER=$2
S3PREFIX=$3

echo "# v$VER"
echo ""
echo "Downloads for kaptain v$VER"
echo ""

find $BUILD -type file -print | xargs shasum -a 256 | sed -e "s/$BUILD\///" |  awk -v ver="$VER" -v s3prefix="$S3PREFIX" 'BEGIN{print "file | sha256 checksum\n---- | ----"}; { print "[" $2  "](https://s3-eu-west-1.amazonaws.com/" s3prefix "/" ver "/" $2 ") | " $1 }'
