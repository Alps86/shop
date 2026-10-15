#!/bin/sh
set -eu
spacing_root=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)
spacing_config="$spacing_root/../../configs/.golangci.yml"
spacing_tmp=$(mktemp -d)
trap 'rm -rf "$spacing_tmp"' EXIT HUP INT TERM
printf 'module example.com/statementspacing\n\ngo 1.22.0\n' > "$spacing_tmp/go.mod"

run_case() {
    cp "$spacing_root/fixtures/$1.go.txt" "$spacing_tmp/spacing.go"
    spacing_status=0
    (cd "$spacing_tmp" && golangci-lint run --config "$spacing_config" --allow-parallel-runners ./...) > "$spacing_tmp/result" 2>&1 || spacing_status=$?
    if [ "$spacing_status" -ne "$2" ]; then
        cat "$spacing_tmp/result"
        printf 'Spacing-Regression fehlgeschlagen: %s (Status %s, erwartet %s)\n' "$1" "$spacing_status" "$2"
        exit 1
    fi
    if [ "$2" -eq 1 ]; then
        grep -q "$3" "$spacing_tmp/result"
    fi
    printf 'Spacing-Regression erfolgreich: %s\n' "$1"
}

run_case valid 0 ''
run_case callback-return 1 '(nlreturn)'
run_case multiline-return 1 '(nlreturn)'
run_case loop-followed-by-code 1 '(wsl_v5)'
