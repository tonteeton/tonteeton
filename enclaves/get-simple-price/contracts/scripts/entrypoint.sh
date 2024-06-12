#!/bin/bash

set -e

echo $1
if [[ $1 == *.ts ]]; then
    script="sources/scripts/$1"
    shift
else
    echo "Error: First argument must be a TypeScript file ending with .ts"
    { cd sources/scripts ; ls *.ts; }
    exit 1
fi

yarn --silent ts-node "$script" "$@"
