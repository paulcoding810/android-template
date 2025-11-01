#!/usr/bin/env bash
set -euo pipefail

# Change Android package name throughout a project (refactored & optimized)

usage() {
  echo "Usage: $0 <new.package.name> <NewProjectName>"
  exit 1
}

if [ "$#" -ne 2 ]; then
  usage
fi

GIT_URL="https://github.com/paulcoding810/android-template"
NEW_PKG="$1"
NEW_PROJECT_NAME="$2"

OLD_PKG="com.paulcoding.androidtemplate"
OLD_PROJECT_NAME="AndroidTemplate"
PROJECT_NAME_LOWER=$(echo "$OLD_PROJECT_NAME" | tr '[:upper:]' '[:lower:]')
NEW_PROJECT_NAME_LOWER=$(echo "$NEW_PROJECT_NAME" | tr '[:upper:]' '[:lower:]')

PROJECT_DIR="$(pwd)/$NEW_PROJECT_NAME"

# Derived values: package in slash form
OLD_PKG_SLASH="${OLD_PKG//./\/}"
NEW_PKG_SLASH="${NEW_PKG//./\/}"

echo "Cloning template..."
git clone --depth 1 "$GIT_URL" "$NEW_PROJECT_NAME"

# Remove git metadata so the new project starts clean
rm -rf "$PROJECT_DIR/.git"

echo "🔍 Replacing package name:"
echo "   from: $OLD_PKG"
echo "   to:   $NEW_PKG"
echo "   in project: $PROJECT_DIR"
echo

# Move package directories (deepest-first)
echo "📂 Renaming package directories..."
# Use -depth so we handle deepest directories first
find "$PROJECT_DIR" -depth -type d -path "*/$OLD_PKG_SLASH" -print0 | while IFS= read -r -d '' old_dir; do
  new_dir="${old_dir//$OLD_PKG_SLASH/$NEW_PKG_SLASH}"
  if [ "$old_dir" != "$new_dir" ]; then
    echo "   $old_dir → $new_dir"
    mkdir -p "$(dirname "$new_dir")"
    if mv "$old_dir" "$new_dir"; then
      # Clean up any now-empty parent directories (stop at project root)
      parent="$(dirname "$old_dir")"
      while [ "$parent" != "$PROJECT_DIR" ] && [ -d "$parent" ] && [ -z "$(ls -A "$parent" 2>/dev/null || true)" ]; do
        rmdir "$parent" || break
        parent="$(dirname "$parent")"
      done
    fi
  fi
done

echo "🧠 Updating source files..."
grep -rl "$OLD_PKG" "$PROJECT_DIR" --include="*.java" --include="*.kt" --include="*.kts" \
  --include="*.xml" --include="*.gradle" --include="*.pro" --include="*.properties" \
  --include=AndroidManifest.xml --exclude=gen.sh | while read -r file; do
    sed -i "" "s/${OLD_PKG}/${NEW_PKG}/g" "$file"
done

sed -i "" "s/${PROJECT_NAME_LOWER}/${NEW_PROJECT_NAME_LOWER}/g" "$PROJECT_DIR/settings.gradle.kts"

grep -rl "$OLD_PROJECT_NAME" "$PROJECT_DIR/app/src/main"  --include="*.kt"  --include="*.xml"    | while read -r file; do
    sed -i "" "s/${OLD_PROJECT_NAME}/${NEW_PROJECT_NAME}/g" "$file"
done


echo "🧹 Removing template files..."
for f in "$PROJECT_DIR/README.md" "$PROJECT_DIR/gen.sh"; do
  if [ -e "$f" ]; then
    rm -f "$f"
  fi
done

echo
echo "✅ Done! Package name changed from $OLD_PKG to $NEW_PKG."
echo "Project created at: $PROJECT_DIR"
# ...existing code...