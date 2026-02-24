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
- Shared tap repo: `mrekoj/homebrew-tap` (`brew tap mrekoj/tap`)
- All formulas live in `Formula/` in that single repo
- Add formula with version + SHA256 per asset
- Install: `brew install mrekoj/tap/<formula>`
- Layout:
  ```
  homebrew-tap/
  └── Formula/
      ├── rtui.rb
      └── voicepill.rb
  ```

*Last updated: February 23, 2026*
