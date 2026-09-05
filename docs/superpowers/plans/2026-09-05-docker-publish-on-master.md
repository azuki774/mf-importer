# Docker Publish on Master Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Publish all five existing Docker images to GHCR for non-docs changes pushed to `master`, using the short commit hash as the image tag while preserving v* release publishing.

**Architecture:** Extend the existing `.github/workflows/publish.yml` rather than creating a second workflow. Add a `master` push trigger with `docs/**` ignored, and add one SHA metadata rule to each existing image job; the current build, login, permissions, Dockerfiles, and release tags remain unchanged.

**Tech Stack:** GitHub Actions, `docker/metadata-action`, `docker/build-push-action`, GitHub Container Registry, existing Go/Node Docker build definitions.

---

### Task 1: Extend the publish trigger and image tags

**Files:**
- Modify: `.github/workflows/publish.yml:3-7` for the push filters
- Modify: `.github/workflows/publish.yml:28-32,71-75,115-119,159-163,203-207` for the five metadata tag lists

- [ ] **Step 1: Add the master push trigger while retaining v* tags**

Change the workflow trigger to include both the existing release tags and master branch pushes, with docs-only filtering:

```yaml
on:
  push:
    branches:
      - master
    tags:
      - v*
    paths-ignore:
      - docs/**
```

The `paths-ignore` filter must be `docs/**`, so a push containing both a docs change and a non-docs change still runs the workflow.

- [ ] **Step 2: Add a bare short-SHA metadata tag to every image job**

Prepend this rule to each of the five existing `tags:` blocks, preserving all existing semver and `latest` rules:

```yaml
            type=sha,format=short,prefix=
```

The resulting metadata configuration for each job must retain these rules after the new line:

```yaml
            type=semver,pattern={{version}}
            type=semver,pattern={{major}}.{{minor}}
            type=semver,pattern={{major}}
            type=semver,pattern=latest
```

`type=sha,format=short,prefix=` produces the short commit hash without the metadata action's default `sha-` prefix. The rule is available for master pushes; semver rules continue to apply to v* tag pushes.

- [ ] **Step 3: Inspect the workflow diff and verify the five jobs are consistent**

Run:

```bash
git diff -- .github/workflows/publish.yml
rg -n "type=sha|type=semver|paths-ignore|branches:|tags:" .github/workflows/publish.yml
```

Expected results:

- `master`, `v*`, and `docs/**` appear in the trigger.
- `type=sha,format=short,prefix=` appears exactly five times.
- All five image names, Dockerfiles, `push: true`, GHCR login, and existing release tag rules remain present.
- No unrelated workflow or repository files are modified.

### Task 2: Validate the workflow and repository

**Files:**
- Verify: `.github/workflows/publish.yml`
- Verify: all repository files through the existing test command

- [ ] **Step 1: Run YAML and workflow checks available in the development shell**

Run:

```bash
nix develop -c actionlint .github/workflows/publish.yml
```

Expected: `actionlint` exits with status 0 and reports no diagnostics. If the development shell does not provide `actionlint`, parse the file with the repository's available YAML checker and record that limitation; do not alter workflow semantics to accommodate a missing local tool.

- [ ] **Step 2: Run the repository test suite**

Run:

```bash
env -u GOROOT -u GOTOOLDIR make test
```

Expected: `gofmt -l` emits no files, static checks succeed, and `go test -v ./...` passes. The environment cleanup avoids the documented Nix/mise Go toolchain mismatch.

- [ ] **Step 3: Review the complete staged diff before committing**

Run:

```bash
git status --short
git diff --cached --stat
git diff --cached
```

Expected: only `.github/workflows/publish.yml` is staged for the implementation commit, and the diff contains no credentials, real financial data, generated files, or unrelated edits.

- [ ] **Step 4: Commit the implementation**

Run:

```bash
git add .github/workflows/publish.yml
git commit -m "ci: publish images on master changes"
```

Expected: a new Conventional Commit on the feature branch `ci/publish-on-master`; do not merge or commit directly to `master`.
