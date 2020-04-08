#!/usr/bin/env bash
set -eu
set -o pipefail

# shellcheck source=./print.sh
source "$(dirname "${BASH_SOURCE[0]}")/print.sh"

# shellcheck source=./git.sh
source "$(dirname "${BASH_SOURCE[0]}")/git.sh"

function util::tools::install() {
    util::print::title "Installing Testing Tools"

    local dir

    while [[ "${#}" != 0 ]]; do
      case "${1}" in
        --help|-h)
          util::tools::usage
          exit 0
          ;;

        --directory)
          dir="${2}"
          shift 2
          ;;

        *)
          util::print::error "unknown argument \"${1}\""
      esac
    done

    mkdir -p "${dir}"

    util::tools::packager::install "${dir}"
    util::tools::cnb2cf::install "${dir}"
}

function util::tools::packager::install () {
    local dir
    dir="${1}"

    if [[ ! -f "${dir}/packager" ]]; then
        util::print::title "Installing packager"
        go build -o "${dir}/packager" github.com/cloudfoundry/libcfbuildpack/packager
    fi
}

function util::tools::cnb2cf::install() {
    local dir
    dir="${1}"

    if [[ ! -f "${dir}/cnb2cf" ]]; then
        util::print::title "Installing cnb2cf"
        go build -o "${dir}/cnb2cf" github.com/cloudfoundry/cnb2cf
    fi
}
