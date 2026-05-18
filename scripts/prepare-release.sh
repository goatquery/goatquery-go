#!/usr/bin/env bash
set -e

# Creates a branch with go.mod version bumps.
# Usage: ./scripts/prepare-release.sh <version>
# Example: ./scripts/prepare-release.sh 1.0.0

VERSION="${1:?Usage: $0 <version>}"
MODULE="github.com/goatquery/goatquery-go"
BRANCH="chore/prepare-release-v${VERSION}"

git checkout -b "${BRANCH}"

for dir in $(./scripts/moduledirs.sh); do
    (cd "${dir}" && go mod edit -require "${MODULE}@v${VERSION}")
done

# Update example project
(cd ./example && go mod edit -require "${MODULE}@v${VERSION}")
(cd ./example && go mod edit -require "${MODULE}/module/gorm@v${VERSION}")

git add .
git commit -m "chore: Prepare release v${VERSION}"
git push origin "${BRANCH}"
