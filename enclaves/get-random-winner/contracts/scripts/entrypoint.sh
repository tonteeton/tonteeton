#!/bin/bash

set -e

if [[ "$1" != *.ts ]]; then
    script="sources/scripts/${1}.ts"
else
    script="sources/scripts/${1}"
fi

if [[ ! -f $script ]]; then
    echo "Error: Script $script does not exist."
    echo "Available scripts:"
    (cd sources/scripts && ls *.ts)
    exit 1
fi

shift
yarn --silent ts-node "$script" "$@"
