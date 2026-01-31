# Release Playbook (Shared)

Versioning:
- Use SemVer (vX.Y.Z)
- Tag matches release assets

Release steps (generic):
1. Run full test suite
2. Build binaries for target platforms
3. Create git tag and push
4. Publish GitHub release with assets
5. Update package manager formulas (Homebrew, etc.)

Homebrew (summary):
- Create tap repo: <user>/homebrew-<app>
- Add formula with version + SHA256 per asset
- Test: brew install <user>/<tap>/<formula>

*Last updated: January 30, 2026*
