#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
GO_DIR="$ROOT/code-dispatcher"
GO_TOOLCHAIN="${GOTOOLCHAIN:-go1.21.13+auto}"

cd "$ROOT"

echo "[quality] checking tracked file size limits"
max_bytes="${CODE_DISPATCHER_MAX_FILE_BYTES:-204800}"
max_lines="${CODE_DISPATCHER_MAX_FILE_LINES:-5000}"
violations=0
while IFS= read -r -d '' file; do
  bytes="$(wc -c < "$file" | tr -d ' ')"
  if (( bytes > max_bytes )); then
    echo "file too large: $file (${bytes} bytes > ${max_bytes})" >&2
    violations=1
  fi

  case "$file" in
    *.go|*.py|*.sh|*.md|*.yml|*.yaml|*.toml)
      lines="$(wc -l < "$file" | tr -d ' ')"
      if (( lines > max_lines )); then
        echo "file too long: $file (${lines} lines > ${max_lines})" >&2
        violations=1
      fi
      ;;
  esac
done < <(git ls-files -z)
if (( violations != 0 )); then
  exit 1
fi

echo "[quality] checking technical debt markers"
if git grep -nE '(^|[^[:alnum:]_.])(T[O]DO|F[I]XME)([^[:alnum:]_]|$)' -- .; then
  echo "technical debt markers must be resolved or tracked outside the code before merging" >&2
  exit 1
fi

echo "[quality] checking gofmt"
mapfile -t go_files < <(git ls-files 'code-dispatcher/*.go')
if ((${#go_files[@]} > 0)); then
  mapfile -t unformatted < <(gofmt -l "${go_files[@]}")
  if ((${#unformatted[@]} > 0)); then
    printf 'gofmt required:\n' >&2
    printf '  %s\n' "${unformatted[@]}" >&2
    exit 1
  fi
fi

check_go_mod_tidy() {
  if ! cd "$GO_DIR"; then
    echo "failed to enter Go module directory: $GO_DIR" >&2
    return 1
  fi

  local tmp tidy_status modfile sumfile
  tidy_status=0
  if ! tmp="$(mktemp -d)"; then
    echo "failed to create temporary directory for go mod tidy check" >&2
    return 1
  fi
  modfile="$tmp/go.mod"
  sumfile="$tmp/go.sum"

  cleanup_tmp() {
    if [[ -z "$tmp" || "$tmp" == "/" ]]; then
      echo "refusing to remove unsafe temporary directory path: ${tmp:-<empty>}" >&2
      return 1
    fi
    if ! rm -rf "$tmp"; then
      echo "failed to remove temporary directory: $tmp" >&2
      return 1
    fi
  }

  if ! cp go.mod "$modfile"; then
    echo "failed to back up go.mod before go mod tidy" >&2
    cleanup_tmp
    return 1
  fi
  if [[ -f go.sum ]]; then
    if ! cp go.sum "$sumfile"; then
      echo "failed to back up go.sum before go mod tidy" >&2
      cleanup_tmp
      return 1
    fi
  fi

  show_tidy_diff() {
    local before="$1" after="$2" diff_status=0
    diff -u "$before" "$after" >&2 || diff_status=$?
    if (( diff_status > 1 )); then
      echo "failed to render diff for $after" >&2
      return 1
    fi
    return 0
  }

  if ! GOTOOLCHAIN="$GO_TOOLCHAIN" go mod tidy -modfile="$modfile"; then
    echo "go mod tidy failed" >&2
    tidy_status=1
  else
    if ! cmp -s go.mod "$modfile"; then
      echo "go mod tidy would change go.mod:" >&2
      show_tidy_diff go.mod "$modfile" || tidy_status=1
      tidy_status=1
    fi

    if [[ -f "$sumfile" ]]; then
      if [[ ! -f go.sum ]]; then
        echo "go mod tidy would create go.sum" >&2
        tidy_status=1
      elif ! cmp -s go.sum "$sumfile"; then
        echo "go mod tidy would change go.sum:" >&2
        show_tidy_diff go.sum "$sumfile" || tidy_status=1
        tidy_status=1
      fi
    elif [[ -f go.sum ]]; then
      echo "go mod tidy would remove go.sum" >&2
      tidy_status=1
    fi
  fi

  if ! cleanup_tmp; then
    return 1
  fi
  return "$tidy_status"
}

echo "[quality] checking go mod tidy"
check_go_mod_tidy

cd "$GO_DIR"
echo "[quality] running go vet"
GOTOOLCHAIN="$GO_TOOLCHAIN" go vet ./...

echo "[quality] running staticcheck"
GOTOOLCHAIN="$GO_TOOLCHAIN" go run honnef.co/go/tools/cmd/staticcheck@2023.1.7 ./...

echo "[quality] checking cyclomatic complexity"
GOTOOLCHAIN="$GO_TOOLCHAIN" go run github.com/fzipp/gocyclo/cmd/gocyclo@v0.6.0 -over "${CODE_DISPATCHER_COMPLEXITY_MAX:-100}" .

echo "[quality] ok"
