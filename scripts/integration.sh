#!/usr/bin/env bash
set -eu
set -o pipefail

readonly PROGDIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
readonly ROOTDIR="$(cd "${PROGDIR}/.." && pwd)"

# shellcheck source=./.util/git.sh
source "${PROGDIR}/.util/git.sh"

# shellcheck source=./.util/images.sh
source "${PROGDIR}/.util/images.sh"

# shellcheck source=./.util/tools.sh
source "${PROGDIR}/.util/tools.sh"

export CF_STACK=${CF_STACK:-"cflinuxfs3"}

function usage() {
  cat <<-USAGE
${PROGDIR}/integration.sh [OPTIONS]
OPTIONS
  --help                 prints the command usage
  --language <language>  specifies the language family to test (nodejs)
  --buildpack <path>     specifies a path to the buildpack under test (eg. nodejs-cnb)
  --build-image <image>  specified an image to use for the build phase (default: cloudfoundry/build:full-cnb)
  --run-image <image>    specified an image to use for the run phase (default: cloudfoundry/run:full-cnb)
  --cached               runs test suite with the --cached option enabled
  --debug                enables debug logging

USAGE
}

function main() {
    local language buildpack debug build_image run_image pack_version cached
      cached="false"
      debug="false"
      build_image="cloudfoundry/build:full-cnb"
      run_image="cloudfoundry/run:full-cnb"
      pack_version="latest"

    while [[ "${#}" != 0 ]]; do
      case "${1}" in
        --help|-h)
          usage
          exit 0
          ;;

        --language)
          language="${2}"
          shift 2
          ;;

        --buildpack)
          buildpack="$(cd "${2}" && pwd)"
          shift 2
          ;;

        --build-image)
          build_image="${2}"
          shift 2
          ;;

        --run-image)
          run_image="${2}"
          shift 2
          ;;

        --pack-version)
          pack_version="${2}"
          shift 2
          ;;

        --cached)
          cached="true"
          shift 1
          ;;

        --debug)
          debug="true"
          shift 1
          ;;

        *)
          usage
          util::print::error "unknown argument \"${1}\""
      esac
    done

    if [[ -z "${language}" ]]; then
      error "--language is a required flag"
    fi

    if [[ -z "${language}" ]]; then
      error "--buildpack is a required flag"
    fi

    util::print::title "Running Integration Test Suite"
    util::print::info "  language:     ${language}"
    util::print::info "  buildpack:    ${buildpack}"
    util::print::info "  build-image:  ${build_image}"
    util::print::info "  run-image:    ${run_image}"
    util::print::info "  pack-version: ${pack_version}"
    util::print::info "  cached:       ${cached}"
    util::print::info "  debug:        ${debug}"

    if [[ "${debug}" == "true" ]]; then
      export CUTLASS_DEBUG="true"
    fi

    GIT_TOKEN="$(util::git::token::fetch)"
    export GIT_TOKEN

    util::images::pull "${build_image}" "${run_image}"

    util::tools::install \
      --directory "${ROOTDIR}/.bin" \
      --pack-version "${pack_version}"

    export PATH="${ROOTDIR}/.bin:${PATH}"

    integration::run "${buildpack}" "${language}"
}

function integration::run() {
  local buildpack language
  buildpack="${1}"
  language="${2}"

  util::print::title "Running Buildpack Runtime Integration Tests"

  set +e
    local exit_code
    GOMAXPROCS=5 BUILDPACK_DIR="${buildpack}" \
      go test \
          "${ROOTDIR}/${language}/..." \
          -v \
          --timeout 0 \
          --cutlass.cached="${cached}"
    exit_code="${?}"
  set -e

  if [[ "${exit_code}" != "0" ]]; then
    util::print::error "** GO Test Failed **"
  else
    util::print::success "** GO Test Succeeded **"
  fi
}

function error() {
  local message
  message="${1}"

  echo "${message}"
  usage
  exit 1
}

main "${@:-}"
