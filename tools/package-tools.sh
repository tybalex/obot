#!/bin/bash
set -e -x -o pipefail

cd $(dirname $0)/..

if [ ! -e otto8-tools ]; then
    git clone --depth=1 https://github.com/otto8-ai/tools otto8-tools
fi

./otto8-tools/scripts/build.sh

for pj in $(find otto8-tools -name package.json | grep -v node_modules); do
    if [ $(basename $(dirname $pj)) == common ]; then
        continue
    fi
    (
        cd $(dirname $pj)
        echo Building $PWD
        pnpm i
    )
done

cd otto8-tools
if [ ! -e workspace-provider ]; then
    git clone --depth=1 https://github.com/gptscript-ai/workspace-provider
fi

cd workspace-provider
CGO_ENABLED=0 go build -ldflags="-s -w" -o bin/gptscript-go-tool .

cd ..
if [ ! -e knowledge-tool ]; then
    git clone --depth=1 https://github.com/gptscript-ai/knowledge knowledge-tool
fi
cd knowledge-tool
make
ln -sf knowledge bin/gptscript-go-tool

cd ..
sed -e 's!github.com/gptscript-ai/knowledge!./knowledge-tool!g' -e 's/@main//g' -i index.yaml
