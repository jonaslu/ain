#!/bin/bash
EXIT_CODE=0
OUTPUT=''

CLEAR='\e[0m'

RED() {
    RED='\033[31m'
    echo -e "${RED}$@${CLEAR}"
}

run_old_new() {
    # TODO temfile stderr, then diff 'em
    # 
    NEW=$(go run cmd/ain/main.go -p $@ 2>&1 | grep -v "exit status")
    OLD=$(lbin/ain -p $@ 2>&1)

    RES=$(diff -y <(echo "${OLD}") <(echo "${NEW}"))

    if [ $? -ne 0 ] ; then
        RED "${@}"

        echo "Output differs old <-> new:"
        echo "${RES}"
        
        # echo "LOCAL"
        # echo "${NEW}"

        # printf "\n----\n"
        # echo "AIN"
        # echo "${OLD}"
    fi
}

FILES=($(find tests/ -type f -name "*.ain"))

for FILE in "${FILES[@]}"; do
    local base_dir=$(dirname $FILE)
    local file_base_name=$(basename $FILE .ain)
    local file_env_path="${base_dir}/.${file_base_name}.env"
    local global_env_path="${base_dir}/.env"

    # TODO _1 _2 files that combine
    if [ -e $file_env_path ]; then
        run_old_new -e $file_env_path $FILE
    elif [ -e $global_env_path ]; then
        run_old_new -e $global_env_path $FILE
    else
        run_old_new $FILE
    fi
done
