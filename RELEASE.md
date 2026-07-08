# Release process

Hitmaker is a Go binary distributed through npm. The npm package is a thin
installer: during `postinstall`, it downloads the native binary from the GitHub
release whose tag matches `npm/package.json`.

That means the GitHub release must exist before publishing to npm.

## Release checklist

1. Pick the next semver version.
2. Update `npm/package.json`.
3. Update `CHANGELOG.md`.
4. Run:

   ```bash
   make fmt
   make test
   make release-check
   ```

5. Confirm `dist/` contains:

   ```text
   hitmaker-darwin-amd64
   hitmaker-darwin-arm64
   hitmaker-linux-amd64
   hitmaker-linux-arm64
   hitmaker-windows-amd64.exe
   hitmaker-windows-arm64.exe
   checksums.txt
   ```

6. Commit the release prep:

   ```bash
   git add .
   git commit -m "Release vX.Y.Z"
   git tag vX.Y.Z
   git push origin HEAD --tags
   ```

7. Create the GitHub release and upload assets:

   ```bash
   gh release create vX.Y.Z dist/* \
     --title "vX.Y.Z" \
     --notes-file CHANGELOG.md
   ```

   The npm installer downloads assets named `hitmaker-<os>-<arch>` from this
   release. Do not publish npm before this step succeeds.

8. Publish npm:

   ```bash
   cd npm
   npm publish --access public
   ```

9. Smoke test from npm after publish:

   ```bash
   npm view hitmaker version
   npm i -g hitmaker
   hitmaker version
   hitmaker --help
   ```

## Local verification

`make build` embeds the npm package version into the Go binary. Check it with:

```bash
bin/hitmaker version
```

`make release-build` builds all binaries that the npm installer supports and
writes SHA-256 checksums to `dist/checksums.txt`.

`make release-check` runs tests, builds release assets, and validates the npm
package contents with `npm pack --dry-run`.
