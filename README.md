This folder(common) is git subrepo.

When common lib has changes. You need:

# 1. Run tests, fix if not pass
go test ./...

# 2. Check last tags and determine version bump
git tag --sort=-version:refname | head -5

# Determine bump type based on changes:
# - MAJOR: breaking API changes
# - MINOR: new features, backward compatible
# - PATCH: bug fixes, small refactorings

# 3. If tests pass, commit changes (ensure you are located in common folder)
git add .
git commit -m "your message, keep it short"

# 4. Create and push new tag
# Example: git tag v1.2.3  (bump appropriately)
git tag vX.X.X
git push origin master --tags