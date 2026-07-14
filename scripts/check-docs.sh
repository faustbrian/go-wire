#!/usr/bin/env bash

set -euo pipefail

go test ./...
scripts/generate-llms.py --check

status=0

while IFS=: read -r source link; do
    case "$link" in
        ""|http://*|https://*|mailto:*|\#*)
            continue
            ;;
    esac

    link="${link%%#*}"
    link="${link#<}"
    link="${link%>}"
    target="$(dirname "$source")/$link"
    if [[ ! -e "$target" ]]; then
        echo "$source: broken local link: $link" >&2
        status=1
    fi
done < <(rg --glob '*.md' --no-heading --with-filename --only-matching \
    --replace '$1' '\[[^]]+\]\(([^)[:space:]]+)(?:[[:space:]]+"[^"]+")?\)')

if rg --glob '*.md' --line-number '[[:blank:]]+$'; then
    echo "Markdown files contain trailing whitespace" >&2
    status=1
fi

if [[ "$status" -eq 0 ]]; then
    echo "Markdown links and generated documentation are valid"
fi

exit "$status"
