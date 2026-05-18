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
mapfile -t unformatted < <(gofmt -l "${go_files[@]}")
if ((${#unformatted[@]} > 0)); then
  printf 'gofmt required:\n' >&2
  printf '  %s\n' "${unformatted[@]}" >&2
  exit 1
fi

check_go_mod_tidy() {
  cd "$GO_DIR"
  local tmp tidy_status
  tmp="$(mktemp -d)"
  tidy_status=0
  cp go.mod "$tmp/go.mod"
  if [[ -f go.sum ]]; then
    cp go.sum "$tmp/go.sum"
  fi

  GOTOOLCHAIN="$GO_TOOLCHAIN" go mod tidy

  if ! cmp -s go.mod "$tmp/go.mod"; then
    echo "go mod tidy would change go.mod:" >&2
    diff -u "$tmp/go.mod" go.mod >&2 || true
    tidy_status=1
  fi

  if [[ -f "$tmp/go.sum" ]]; then
    if [[ ! -f go.sum ]] || ! cmp -s go.sum "$tmp/go.sum"; then
      echo "go mod tidy would change go.sum:" >&2
      diff -u "$tmp/go.sum" go.sum >&2 || true
      tidy_status=1
    fi
  elif [[ -f go.sum ]]; then
    echo "go mod tidy would create go.sum" >&2
    tidy_status=1
  fi

  cp "$tmp/go.mod" go.mod
  if [[ -f "$tmp/go.sum" ]]; then
    cp "$tmp/go.sum" go.sum
  else
    rm -f go.sum
  fi
  rm -rf "$tmp"
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
