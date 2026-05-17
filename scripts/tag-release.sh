#!/usr/bin/env bash
set -e

# Tags the current commit with root + all sub-module tags and pushes.
# Usage: ./scripts/tag-release.sh <version>
# Example: ./scripts/tag-release.sh 1.0.0
# Run this on main after the prepare-release PR is merged.

VERSION="${1:?Usage: $0 <version>}"
TAGS="v${VERSION}"

for dir in $(./scripts/moduledirs.sh); do
    TAGS="${TAGS} ${dir}/v${VERSION}"
done

echo "Creating tags:"
for tag in ${TAGS}; do
    echo "  ${tag}"
    git tag "${tag}"
done

echo ""
echo "Pushing tags..."
git push origin ${TAGS}

echo "Done. Released: ${TAGS}"
