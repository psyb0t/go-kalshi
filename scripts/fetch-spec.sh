#!/bin/bash
# Downloads one Kalshi API spec into api/. The specs are not committed; each
# generated package fetches its spec as the first step of `go generate`.
set -euo pipefail
trap 'on_error "$?" "$LINENO"' ERR

LOG_FILE="${LOG_FILE:-/tmp/$(basename "$0" .sh).log}"
exec > >(tee -a "$LOG_FILE") 2>&1

readonly SPEC_BASE_URL="https://docs.kalshi.com"
readonly DOWNLOAD_TIMEOUT_SECONDS=60
readonly SPEC_OPENAPI="openapi.yaml"
readonly SPEC_ASYNCAPI="asyncapi.yaml"
readonly SPEC_PERPS_ASYNCAPI="perps_asyncapi.yaml"
SPEC_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)/api"
readonly SPEC_DIR

on_error() {
	local exit_code="$1"
	local line="$2"

	printf '{"level":"ERROR","file":"%s","line":%d,"msg":"command failed","exit_code":%d}\n' \
		"${BASH_SOURCE[1]}" "$line" "$exit_code" >&2
}

log() {
	local level="$1"
	shift

	jq -nc \
		--arg time "$(date -u '+%Y-%m-%dT%H:%M:%SZ')" \
		--arg level "$level" \
		--arg file "${BASH_SOURCE[1]##*/}" \
		--arg line "${BASH_LINENO[0]}" \
		--arg func "${FUNCNAME[1]:-main}" \
		--arg msg "$*" \
		'{time:$time,level:$level,file:$file,line:($line|tonumber),func:$func,msg:$msg}' >&2
}

usage() {
	printf 'Usage: %s <%s|%s|%s>\n' "$(basename "$0")" \
		"$SPEC_OPENAPI" "$SPEC_ASYNCAPI" "$SPEC_PERPS_ASYNCAPI" >&2
}

# The specs are YAML and the dev image has no YAML parser, so the version is
# read from the top-level info block as plain lines.
spec_version() {
	local path="$1"

	awk '/^info:/ { in_info = 1; next } in_info && /^[^ ]/ { exit } in_info && $1 == "version:" { print $2; exit }' "$path" |
		tr -d "\"'"
}

main() {
	local spec="${1:-}"
	if [[ "$spec" != "$SPEC_OPENAPI" && "$spec" != "$SPEC_ASYNCAPI" && "$spec" != "$SPEC_PERPS_ASYNCAPI" ]]; then
		log ERROR "unknown spec name=${spec}"
		usage
		exit 2
	fi

	mkdir -p "$SPEC_DIR"
	local tmp
	tmp="$(mktemp "${SPEC_DIR}/.${spec}.XXXXXX")"
	# shellcheck disable=SC2064 # expand tmp now, the local is gone when EXIT fires
	trap "rm -f -- '${tmp}'" EXIT

	log INFO "downloading spec=${spec}"
	wget -q \
		-T "$DOWNLOAD_TIMEOUT_SECONDS" \
		-O "$tmp" \
		"${SPEC_BASE_URL}/${spec}"

	if ! grep -Eq '^(openapi|asyncapi): ' "$tmp"; then
		log ERROR "downloaded file is not an OpenAPI or AsyncAPI document spec=${spec}"
		exit 1
	fi

	chmod 0644 "$tmp"
	mv -- "$tmp" "${SPEC_DIR}/${spec}"
	log INFO "saved spec=${spec} version=$(spec_version "${SPEC_DIR}/${spec}")"
}

main "$@"
