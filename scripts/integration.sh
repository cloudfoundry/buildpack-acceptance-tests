#!/usr/bin/env bash
set -e
set -u
set -o pipefail

readonly PROGDIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
readonly ROOTDIR="$(cd "${PROGDIR}/.." && pwd)"

# shellcheck source=./.util/git.sh
source "${PROGDIR}/.util/git.sh"

# shellcheck source=./.util/tools.sh
source "${PROGDIR}/.util/tools.sh"

# shellcheck source=./.util/cf.sh
source "${PROGDIR}/.util/cf.sh"

function usage() {
  cat <<-USAGE
integration.sh --language <language> --buildpack <path> [OPTIONS]

OPTIONS
  --help                         prints the command usage
  --language <language>          specifies the language family to test (nodejs)
  --buildpack <path>             specifies a path to a buildpack directory or zip under test (eg. nodejs-cnb.zip)
  --buildpack-version <version>  specifies a version number for the buildpack if --buildpack is a directory (eg. 1.7.1)
  --stack <stack>                name of the stack to use when running tests (default: cflinuxfs3)
  --cached                       runs test suite with the --cached option enabled (default: false)
  --debug                        enables debug logging (default: false)

USAGE
}

function main() {
    local language buildpack_input buildpack_version debug pack_version cached stack
      cached="false"
      debug="false"
      pack_version="latest"
      stack="cflinuxfs3"

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
          buildpack_input="${2}"
          shift 2
          ;;

        --buildpack-version)
          buildpack_version="${2}-$(date "+%Y%m%d%H%M%S")"
          shift 2
          ;;

        --stack)
          stack="${2}"
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

        "")
          shift 1
          ;;

        *)
          usage
          util::print::error "unknown argument \"${1}\""
      esac
    done

    export CF_STACK=${CF_STACK:-"${stack}"}

    if [[ -z "${language}" ]]; then
      error "--language is a required flag"
    fi

    if [[ -z "${buildpack_input}" ]]; then
      error "--buildpack is a required flag"
    fi

    if [[ "${buildpack_input}" == *.zip ]]; then
      local bp_version
      bp_version="$(unzip -p "${buildpack_input}" VERSION)"
      if [[ "${buildpack_version}" != "" ]]; then
        util::print::title "Overriding --buildpack-version"
        util::print::info "Buildpack version \"${buildpack_version}\", but buildpack zip specifies \"${bp_version}\"."
      fi

      buildpack_version="${bp_version}"
    fi

    if [[ -z "${buildpack_version}" ]]; then
      error "--buildpack-version is a required flag when --buildpack is a directory"
    fi

    if [[ "${debug}" == "true" ]]; then
      export CUTLASS_DEBUG="true"
    fi

    GIT_TOKEN="$(util::git::token::fetch)"
    export GIT_TOKEN

    if ! util::cf::check; then
      exit 1
    fi

    util::tools::install \
      --directory "${ROOTDIR}/.bin" \
      --pack-version "${pack_version}"

    export PATH="${ROOTDIR}/.bin:${PATH}"

    local buildpack
    buildpack="${buildpack_input}"

    if [[ "${buildpack_input}" != *.zip ]]; then
      local tmpdir
      tmpdir="$(mktemp -d)"

      cp -r "${buildpack_input}" "${tmpdir}"
      buildpack="$(set -e; integration::package "$(cd "${tmpdir}" && pwd)" "${buildpack_version}" "${stack}")"
    fi

    util::print::title "Running Integration Test Suite"
    util::print::info "  language:           ${language}"
    util::print::info "  buildpack:          ${buildpack}"
    util::print::info "  buildpack-version:  ${buildpack_version}"
    util::print::info "  stack:              ${stack}"
    util::print::info "  pack-version:       ${pack_version}"
    util::print::info "  cached:             ${cached}"
    util::print::info "  debug:              ${debug}"

    integration::run "${buildpack}" "${language}" "${buildpack_version}"
}

function integration::run() {
  local buildpack language version
  buildpack="${1}"
  language="${2}"
  version="${3}"

  set +e
    local exit_code
    GOMAXPROCS=5 \
      go test \
          "${ROOTDIR}/${language}/..." \
          -v \
          --timeout 0 \
          --cutlass.cached="${cached}" \
          --buildpack="${buildpack}" \
          --buildpack-version="${version}"
    exit_code="${?}"
  set -e

  if [[ "${exit_code}" != "0" ]]; then
    util::print::error "** GO Test Failed **"
  else
    util::print::success "** GO Test Succeeded **"
  fi
}

function integration::package() {
  util::print::title "Packaging Buildpack"

  local dir version stack
  dir="${1}"
  version="${2}"
  stack="${3}"

  pushd "${dir}" > /dev/null || return
    "${ROOTDIR}/.bin/cnb2cf" package \
      --version "${version}" \
      --stack "${stack}" \
      1>&2
  popd > /dev/null || return

  local zip_dir
  zip_dir="$(mktemp -d)"

  mv ${dir}/*.zip "${zip_dir}/"

  printf "%s" "$(ls ${zip_dir}/*.zip)"
}

function error() {
  local message
  message="${1}"

  usage
  util::print::error "${message}"
}

main "${@:-}"
