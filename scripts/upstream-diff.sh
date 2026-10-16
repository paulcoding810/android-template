#!/usr/bin/env bash
# Diff a template's build config against a fresh `android create` project.
set -euo pipefail

TEMPLATE=${1:-template-single}
ANDROID_TEMPLATE=${2:-empty-activity}
FILES=(
  gradle/libs.versions.toml
  gradle/wrapper/gradle-wrapper.properties
  gradle.properties
  settings.gradle.kts
  build.gradle.kts
  app/build.gradle.kts
  app/src/main/AndroidManifest.xml
)

command -v android >/dev/null || { echo "== Version differences (template -> android CLI reference)"
python3 - "$TEMPLATE/gradle/libs.versions.toml" "$REF/gradle/libs.versions.toml" <<'PY'
import sys, tomllib

def coords(path):
    with open(path, "rb") as f:
        cat = tomllib.load(f)
    versions = cat.get("versions", {})
    out = {}
    for section, key in (("libraries", "module"), ("plugins", "id")):
        for entry in cat.get(section, {}).values():
            if isinstance(entry, str):
                name, _, ver = entry.rpartition(":")
                out[name] = ver
                continue
            name = entry.get(key) or f"{entry.get('group')}:{entry.get('name')}"
            ver = entry.get("version")
            if isinstance(ver, dict):
                ver = versions.get(ver.get("ref"), "?")
            elif ver is None and "version.ref" in entry:
                ver = versions.get(entry["version.ref"], "?")
            if ver:
                out[name] = ver
    return out

ours, ref = coords(sys.argv[1]), coords(sys.argv[2])
rows = [(n, ours[n], ref[n]) for n in sorted(ours.keys() & ref.keys()) if ours[n] != ref[n]]
for name, a, b in rows:
    print(f"  {name:60} {a:>14} -> {b}")
if not rows:
    print("  none")
only = sorted(ref.keys() - ours.keys())
if only:
    print("== Only in reference")
    for name in only:
        print(f"  {name:60} {ref[name]:>14}")
PY

echo
echo "== File diffs"
echo "android CLI not found" >&2; exit 1; }
[ -d "$TEMPLATE" ] || { echo "no such template: $TEMPLATE" >&2; exit 1; }

REF=$(mktemp -d)/ref
trap 'rm -rf "$(dirname "$REF")"' EXIT

android create --name=AndroidTemplate -o="$REF" "$ANDROID_TEMPLATE" >/dev/null
grep -rl com.example.androidtemplate "$REF" | xargs sed -i.bak 's/com\.example\.androidtemplate/com.paulcoding.androidtemplate/g'
find "$REF" -name '*.bak' -delete

echo "== Version differences (template -> android CLI reference)"
python3 - "$TEMPLATE/gradle/libs.versions.toml" "$REF/gradle/libs.versions.toml" <<'PY'
import sys, tomllib

def coords(path):
    with open(path, "rb") as f:
        cat = tomllib.load(f)
    versions = cat.get("versions", {})
    out = {}
    for section, key in (("libraries", "module"), ("plugins", "id")):
        for entry in cat.get(section, {}).values():
            if isinstance(entry, str):
                name, _, ver = entry.rpartition(":")
                out[name] = ver
                continue
            name = entry.get(key) or f"{entry.get('group')}:{entry.get('name')}"
            ver = entry.get("version")
            if isinstance(ver, dict):
                ver = versions.get(ver.get("ref"), "?")
            elif ver is None and "version.ref" in entry:
                ver = versions.get(entry["version.ref"], "?")
            if ver:
                out[name] = ver
    return out

ours, ref = coords(sys.argv[1]), coords(sys.argv[2])
rows = [(n, ours[n], ref[n]) for n in sorted(ours.keys() & ref.keys()) if ours[n] != ref[n]]
for name, a, b in rows:
    print(f"  {name:60} {a:>14} -> {b}")
if not rows:
    print("  none")
only = sorted(ref.keys() - ours.keys())
if only:
    print("== Only in reference")
    for name in only:
        print(f"  {name:60} {ref[name]:>14}")
PY

echo
echo "== File diffs"
echo "android CLI $(android --version 2>/dev/null | head -1) / $ANDROID_TEMPLATE vs $TEMPLATE"
for f in "${FILES[@]}"; do
  [ -f "$TEMPLATE/$f" ] || [ -f "$REF/$f" ] || continue
  git --no-pager diff --no-index --color=auto \
    "$TEMPLATE/$f" "$REF/$f" 2>/dev/null || true
done
