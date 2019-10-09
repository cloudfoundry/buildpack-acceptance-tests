#!/usr/bin/env bash

set -eu
set -o pipefail

# shellcheck source=./print.sh
source "$(dirname "${BASH_SOURCE[0]}")/print.sh"

function util::images::pull() {
    while (( "${#}" )); do
        util::print::title "Pulling Docker Image: ${1}"

        docker pull "${1}"
        shift 1
    done
}
