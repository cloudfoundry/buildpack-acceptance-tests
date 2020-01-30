#!/usr/bin/env bash

set -eu
set -o pipefail

# shellcheck source=./print.sh
source "$(dirname "${BASH_SOURCE[0]}")/print.sh"

function util::cf::check() {
    util::print::title "Checking for available CF environment"
    if ! cf apps > /dev/null 2>&1; then
        echo "Not logged in to a CF environment!"
        return 1
    fi
}
