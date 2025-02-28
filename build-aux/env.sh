#!/bin/bash
if [ -z "$ENV_SETUPED" ];then
    export ENV_SETUPED=1
    export ENTRYPOINT=${ENTRYPOINT:-entrypoint.sh}
    export KILLER_EXEC=${KILLER_EXEC:-$(which ll-killer)}
    export PATH=$PATH:$(dirname $0)
    
    if [ -z "$KILLER_EXEC" ];then
        echo "错误：未找到ll-killer，请确保ll-killer在PATH环境变量中，或使用'll-killer script $*'重新执行命令。" >&2
        exit 1
    fi
fi