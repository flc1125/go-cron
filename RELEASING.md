# Releasing go-cron

This guide describes how maintainers publish a stable release, from version selection to the GitHub release page. Run commands from the repository root.

## Release conventions

The [Makefile](Makefile) provides the release commands. [versions.yaml](versions.yaml) defines the module groups and their versions.

| Setting | Convention |
| --- | --- |
| Main branch | `4.x` |
| Remote | `origin` |
| Module set | `stable` |
| Release branch | `release/<version>` |
| Generated preparation branch | `prerelease_stable_<version>` |
| Core tag | `<version>` |
| Middleware tag | `<module-directory>/<version>` |
| PR title | `Release <version>` |
| GitHub release title | `<version>` |

The stable set currently contains the core module and six middleware modules. They share one release version. The `internal/tools` and `tests` modules are excluded from published tags, although preparation updates the test module's internal dependencies.

Use a Go toolchain supported by the repository, Git, Make, and an authenticated GitHub CLI (`gh`) with push and release access. Make builds the release tools in `.tools` as needed.

Set the main branch and module set in your shell:

```sh
BASE_BRANCH=4.x
MODSET=stable
```

The examples below use the completed `v4.13.0` release. Replace the version values with the target release and its previous published version. Keep the variables available if you resume in a new shell.

## 1. Choose a version

Start with a clean working tree, then update the main branch and fetch tags:

```sh
git status --short --branch
git switch "$BASE_BRANCH"
git pull --ff-only --tags origin "$BASE_BRANCH"
make gorelease
```

Read each module's inferred base version, suggested version, and diagnostics. Combine the suggestions with a review of the changes since the previous release, including public APIs, behavior, dependencies, and Go requirements. Choose a shared version that satisfies the stable modules' release requirements.

`make gorelease` also checks the test module, which is excluded from the stable release set. Its suggestion does not determine the stable version. The Makefile catches individual gorelease failures, so an exit code of zero does not mean every module passed. Investigate diagnostics before proceeding; see [Troubleshooting](#troubleshooting).

For `v4.13.0`, gorelease suggested `v4.13.0` for `redismutex` and `otel`, and `v4.12.1` for the other stable modules. The shared release version was `v4.13.0`.

After selecting the version, set:

```sh
VERSION=v4.13.0
PREVIOUS_VERSION=v4.12.0
```

## 2. Prepare the release

Create the release branch:

```sh
git switch -c "release/$VERSION"
```

Edit `module-sets.stable.version` in `versions.yaml` to match `$VERSION`, including the `v` prefix. Commit the configuration before running preparation:

```sh
git add versions.yaml
git commit -m "chore(release): prepare $VERSION"
git status --short --branch
make prerelease MODSET="$MODSET"
```

The working tree must be clean when preparation starts. The command:

1. Verifies the module-set configuration.
2. Updates `version.go` and internal module requirements in `go.mod` files, including those in the test module.
3. Runs module tidying.
4. Creates `prerelease_stable_<version>` and commits the generated changes there.
5. Returns to the original `release/<version>` branch.

Merge the generated branch into the release branch and review the complete preparation diff:

```sh
git merge --ff-only "prerelease_${MODSET}_${VERSION}"
git diff --check "origin/$BASE_BRANCH...HEAD"
git diff "origin/$BASE_BRANCH...HEAD"
git status --short --branch
```

Confirm that `versions.yaml`, the value returned by `Version()`, and the internal module requirements agree. The returned version string omits the `v` prefix. The worktree should be clean after the merge.

## 3. Open and merge the release PR

Push the release branch:

```sh
git push -u origin "release/$VERSION"
```

Write a PR description covering the version-alignment changes, the history being released, and validation actually performed. Use these comparison ranges:

- Release preparation: PR base commit to PR head commit.
- Changes being released: previous release tag to PR base commit.

Resolve the range endpoints to commit SHAs when recording the inventory. Verify PR membership against the commit range, including dependency updates and any direct commits. The release-preparation PR is not yet part of its own base history.

Create the ignored `.docs` directory, write and review `.docs/release-pr.md`, then open the PR:

```sh
mkdir -p .docs
# Write and review .docs/release-pr.md before running gh pr create.
gh pr create \
  --base "$BASE_BRANCH" \
  --head "release/$VERSION" \
  --title "Release $VERSION" \
  --body-file .docs/release-pr.md
```

Review the PR and its checks:

```sh
PR_NUMBER=$(gh pr view "release/$VERSION" --json number --jq '.number')
gh pr checks "$PR_NUMBER" --watch
```

The current workflows run lint, builds, race tests on Go 1.26 and 1.27, and coverage reporting. See [Go Lint](.github/workflows/lint.yml) and [Go Test](.github/workflows/test.yml) for the current commands and matrix. Record successful, pending, and failed checks accurately.

Merge the PR into `4.x` after review and successful checks.

## 4. Update the main branch and create tags

After the PR is merged, update the main branch again immediately before tagging:

```sh
git switch "$BASE_BRANCH"
git pull --ff-only --tags origin "$BASE_BRANCH"
git status --short --branch
git log -1 --oneline
```

Confirm the worktree is clean, the branch is synchronized with `origin/4.x`, and `HEAD` is the intended release commit. If additional commits have landed since the release PR, review whether they belong in the release before tagging. Check that the stable version in `versions.yaml` still matches `$VERSION`.

```sh
make add-tags MODSET="$MODSET"
git tag --points-at HEAD --list "*$VERSION"
```

`add-tags` verifies the module set and creates local annotated tags at `HEAD` by default. For the current stable set, expect seven tags pointing to the same commit:

```text
v4.13.0
middleware/delayoverlapping/v4.13.0
middleware/distributednooverlapping/v4.13.0
middleware/distributednooverlapping/redismutex/v4.13.0
middleware/nooverlapping/v4.13.0
middleware/otel/v4.13.0
middleware/recovery/v4.13.0
```

Use the membership in `versions.yaml` when checking the expected tag list for future releases.

## 5. Push tags

```sh
make push-tags TAG="$VERSION"
git ls-remote --tags origin "refs/tags/*$VERSION" "refs/tags/*$VERSION^{}"
```

`push-tags` pushes local tags whose names end with the supplied version. Confirm that every expected tag reached `origin`. In the remote listing, entries ending in `^{}` show the commit referenced by each annotated tag; they should all match the intended release commit.

## 6. Publish the GitHub release

Generate a starting changelog for the two published tags:

```sh
mkdir -p .docs
gh api repos/flc1125/go-cron/releases/generate-notes \
  -f tag_name="$VERSION" \
  -f previous_tag_name="$PREVIOUS_VERSION" \
  --jq .body > ".docs/release-notes-$VERSION.md"
```

Edit the draft using the [release notes conventions](#release-notes-conventions). Verify the generated PR list against `git log "$PREVIOUS_VERSION..$VERSION"`. This tag-to-tag range includes the merged release-preparation PR. Preserve every included PR and its author attribution when editing the list.

Publish the reviewed notes as a stable release and mark it Latest:

```sh
gh release create "$VERSION" \
  --verify-tag \
  --title "$VERSION" \
  --latest \
  --notes-file ".docs/release-notes-$VERSION.md"
```

Verify the published body, tag, and release status:

```sh
gh release view "$VERSION" --json url,tagName,name,isDraft,isPrerelease,body
gh api repos/flc1125/go-cron/releases/latest --jq '.tag_name'
```

The release should use `$VERSION`, have both `isDraft` and `isPrerelease` set to `false`, and appear as Latest. Open the returned URL to review the rendered headings, code blocks, and links.

To correct the published wording, edit the same notes file and update the release:

```sh
gh release edit "$VERSION" --notes-file ".docs/release-notes-$VERSION.md"
```

Read the published body again after an update.

## Troubleshooting

### gorelease reports missing checksums

Run `go mod tidy` in the affected module, inspect the resulting `go.mod` and `go.sum` changes, and rerun gorelease. Commit any reviewed changes before invoking prerelease, which requires a clean worktree. If diagnostics persist, investigate and record them rather than treating the Make exit status as a passing result.

### prerelease reports a dirty worktree while Git reports it clean

Compare Git's status with its ignore rules:

```sh
git status --short --untracked-files=all
git status --short --ignored
git check-ignore -v .agents skills-lock.json
```

During the `v4.13.0` release, multimod's Git library treated `.agents/` and `skills-lock.json` as untracked even though `.git/info/exclude` ignored them for Git. Adding the entries to the repository's `.gitignore` and committing the change resolved the failure. These rules are now part of the repository.

### The generated prerelease branch already exists

A repeated preparation run can fail when `prerelease_stable_<version>` already exists. The failed command may have updated version files before attempting to create the branch.

Inspect both the worktree and the existing branch:

```sh
git status --short --branch
git diff
git log --oneline "HEAD..prerelease_${MODSET}_${VERSION}"
git diff "HEAD...prerelease_${MODSET}_${VERSION}"
```

If the existing branch contains the correct preparation changes, use it after resolving the failed run's worktree changes. If regeneration is needed, rename the existing branch to an unused backup name, revert only the changes produced by the failed run, and rerun preparation from the clean release branch. Preserve unrelated work when resolving the state.

## Release notes conventions

Use concise English prose and this structure:

1. One-sentence summary of the release.
2. **Highlights**: the changes most relevant to users, with package names, versions, and PR links.
3. **Upgrade**: the required Go version, the core upgrade command, and commands for directly imported middleware modules. Include migration instructions when compatibility changes require action.
4. **What's Changed**: the complete PR inventory, grouped by the changes present in the release.
5. **Full changelog**: a link comparing the previous release tag with the new tag.

Use direct wording such as "Update OpenTelemetry from `v1.45.0` to `v1.46.0`." Format package names and versions as inline code, and link PRs using their numbers. Preserve author attribution. Keep headings and lists consistent and use bold sparingly.

Focus the summary on actual changes. State the Go requirement once in Upgrade. Keep CI execution records and inventory counts in the release PR. Remove repeated compatibility statements and routine publication details from the release page. Add sections such as breaking changes only when the release needs them.

For a completed example, see [Release PR #1132](https://github.com/flc1125/go-cron/pull/1132) and the [v4.13.0 release](https://github.com/flc1125/go-cron/releases/tag/v4.13.0).
