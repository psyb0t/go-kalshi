#!/bin/bash
# Writes api/<name>.client.yaml from a downloaded Kalshi AsyncAPI document, with
# two changes that make the generated code a working client:
#
#  1. Kalshi's documents describe the server: a command the client sends is a
#     `receive` operation and every push is a `send` operation. gen-asyncapi
#     builds the side the document describes, so every operation action is
#     swapped.
#  2. The shared market ticker schema carries `pattern: '^[A-Z0-9-]+$'`, which
#     rejects real tickers with a dot, including the schema's own examples
#     (`FED-23DEC-T3.00`). That pattern is dropped.
set -euo pipefail
trap 'on_error "$?" "$LINENO"' ERR

LOG_FILE="${LOG_FILE:-/tmp/$(basename "$0" .sh).log}"
exec > >(tee -a "$LOG_FILE") 2>&1

SPEC_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)/api"
readonly SPEC_DIR
readonly CLIENT_SUFFIX=".client.yaml"

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
	printf 'Usage: %s <spec.yaml>\n' "$(basename "$0")" >&2
}

main() {
	local spec="${1:-}"
	if [[ -z "$spec" || "$spec" != *.yaml ]]; then
		usage
		exit 2
	fi

	local source_spec="${SPEC_DIR}/${spec}"
	local output_spec="${SPEC_DIR}/${spec%.yaml}${CLIENT_SUFFIX}"
	if [[ ! -f "$source_spec" ]]; then
		log ERROR "source spec missing path=${source_spec}"
		exit 1
	fi

	local tmp
	tmp="$(mktemp "${SPEC_DIR}/.${spec}.client.XXXXXX")"
	# shellcheck disable=SC2064 # expand tmp now, the local is gone when EXIT fires
	trap "rm -f -- '${tmp}'" EXIT

	# The spec is YAML and the dev image has no YAML parser. Operation actions
	# sit at a fixed indent inside the top-level operations block, and the dropped
	# pattern is a whole line, so line edits are exact here.
	local counts
	if ! counts="$(awk -v out="$tmp" '
		/^[^ #]/ { in_operations = ($0 ~ /^operations:/) }
		in_operations && /^    action: send[[:space:]]*$/ { print "    action: receive" > out; flipped++; next }
		in_operations && /^    action: receive[[:space:]]*$/ { print "    action: send" > out; flipped++; next }
		/^[[:space:]]+pattern: .\^\[A-Z0-9-\]\+\$.[[:space:]]*$/ { patterns++; next }
		{ print > out }
		END {
			printf "%d %d", flipped, patterns
			if (flipped == 0) exit 3
		}
	' "$source_spec")"; then
		log ERROR "no operation actions found to swap spec=${spec}"
		exit 1
	fi

	chmod 0644 "$tmp"
	mv -- "$tmp" "$output_spec"
	log INFO "matched client spec=${spec} output=${output_spec##*/} actions_patterns=${counts}"
}

main "$@"
