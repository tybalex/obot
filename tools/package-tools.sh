#!/bin/bash
set -e -x

cd $(dirname $0)/..

git clone --depth=1 https://github.com/otto8-ai/tools otto8-tools

for gomod in $(find otto8-tools -name go.mod); do
    if [ $(basename $(dirname $gomod)) == common ]; then
        continue
    fi
    (
        cd $(dirname $gomod)
        echo Building $PWD
        go build -o bin/gptscript-go-tool .
    )
done

cd otto8-tools
git clone --depth=1 https://github.com/gptscript-ai/workspace-provider
cd workspace-provider
go build -o bin/gptscript-go-tool .
rm -rf .git
cd ..
rm -rf .git
