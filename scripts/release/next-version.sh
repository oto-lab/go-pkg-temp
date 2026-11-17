#!/usr/bin/env bash
set -euo pipefail

fail() {
  printf 'Release: %s\n' "$*" >&2
  exit 1
}

semver() {
  npx --yes semver@7.8.5 "$@"
}

version=$VERSION
preid=$PREID
[[ -z $preid || $preid =~ ^[a-z]+$ ]] || fail 'preid must contain lowercase letters only'

core='(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)'
tag_pattern="^v$core(-[0-9A-Za-z-]+(\.[0-9A-Za-z-]+)*)?$"
all_tags=$(git tag --list 'v*')
versions=()
while IFS= read -r tag; do
  if [[ $tag =~ $tag_pattern ]]; then
    versions+=("$tag")
  fi
done <<< "$all_tags"

current=0.0.0
if ((${#versions[@]})); then
  sorted=$(semver -p "${versions[@]}") || fail 'could not sort the version tags'
  current=${sorted##*$'\n'}
fi

explicit=false
case "$version" in
  patch|minor|major)
    if [[ -n $preid ]]; then
      next=$(semver -i "pre$version" --preid "$preid" "$current")
    else
      next=$(semver -i "$version" "$current")
    fi
    ;;
  prerelease)
    if [[ -n $preid ]]; then
      next=$(semver -i prerelease --preid "$preid" "$current")
    else
      [[ $current == *-* ]] || fail 'a prerelease needs an identifier after a stable version'
      next=$(semver -i prerelease "$current")
    fi
    ;;
  *)
    [[ -z $preid ]] || fail 'leave preid empty for an explicit version'
    base=${version#v}
    if [[ $base =~ ^$core-[a-z]+$ ]]; then
      id=${base##*-}
      series=()
      for tag in "${versions[@]}"; do
        if [[ $tag == "v$base."* && ${tag#"v$base."} =~ ^(0|[1-9][0-9]*)$ ]]; then
          series+=("$tag")
        fi
      done
      if ((${#series[@]})); then
        sorted=$(semver -p "${series[@]}")
        next=$(semver -i prerelease --preid "$id" "${sorted##*$'\n'}")
      else
        next="$base.0"
      fi
    elif [[ $base =~ ^$core(-[a-z]+\.(0|[1-9][0-9]*))?$ ]]; then
      explicit=true
      next=$(semver "$base") || fail "invalid explicit version: $version"
    else
      fail "invalid version: $version"
    fi
    ;;
esac

[[ $next =~ ^$core(-[a-z]+\.(0|[1-9][0-9]*))?$ ]] || fail "expected a version with an optional lowercase prerelease identifier and number: $next"
tag=v$next
retry=false
if ((${#versions[@]})); then
  if [[ $explicit == true && $next == "$current" ]]; then
    retry=true
  elif ! semver -r ">$current" -p "$next" >/dev/null; then
    fail "$tag must be newer than v$current; enter v$current explicitly to retry it"
  fi
fi

# A retry releases the module recorded at the tag even if main has since changed its module path.
if [[ $retry == true ]]; then
  manifest=$(git show "$tag:go.mod")
else
  manifest=$(cat go.mod)
fi
module=$(awk '$1 == "module" {print $2; exit}' <<< "$manifest")
module=${module#\"}
module=${module%\"}
[[ -n $module ]] || fail 'go.mod does not declare a module path'

major=${next%%.*}
if [[ $major == 0 || $major == 1 ]]; then
  [[ ! $module =~ /v[0-9]+$ ]] || fail "$tag requires a module path without a /vN suffix: $module"
else
  [[ $module == */v"$major" ]] || fail "$tag requires a module path ending in /v$major: $module"
fi

printf 'tag=%s\nretry=%s\n' "$tag" "$retry"
