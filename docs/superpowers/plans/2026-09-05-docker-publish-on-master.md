# Selective Docker Publish Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Publish only Docker images affected by a `master` push under a bare short-SHA tag, while preserving full semver publishing for `v*` tags.

**Architecture:** Add one change-detection job to the existing workflow and expose one output per image. Keep the five build jobs, gate each with its corresponding output or a tag-event override, and restrict the SHA metadata tag to branch events.

**Tech Stack:** GitHub Actions, `dorny/paths-filter` v4.0.3, Docker metadata/build actions, GHCR.

---

### Task 1: Add image-level change detection

**Files:**
- Modify: `.github/workflows/publish.yml`

- [ ] **Step 1: Add the change-detection job**

Insert this job before the existing build jobs:

```yaml
  changes:
    runs-on: ubuntu-latest
    permissions:
      contents: read
    outputs:
      importer: ${{ steps.filter.outputs.importer }}
      maw: ${{ steps.filter.outputs.maw }}
      api: ${{ steps.filter.outputs.api }}
      metrics: ${{ steps.filter.outputs.metrics }}
      frontend: ${{ steps.filter.outputs.frontend }}
    steps:
      - name: checkout
        if: github.ref_type == 'branch'
        uses: actions/checkout@3d3c42e5aac5ba805825da76410c181273ba90b1 # v7.0.1
        with:
          fetch-depth: 0

      - name: Detect image changes
        id: filter
        if: github.ref_type == 'branch'
        uses: dorny/paths-filter@ceb8a2b8f2d89434be7ff52d3de7ec3738c5cc9d # v4.0.3
        with:
          base: ${{ github.event.before }}
          ref: ${{ github.sha }}
          filters: |
            importer:
              - '.dockerignore'
              - '.github/workflows/publish.yml'
              - 'build/Dockerfile'
              - 'cmd/mf-importer/**'
              - 'go.mod'
              - 'go.sum'
              - 'internal/**'
            maw:
              - '.dockerignore'
              - '.github/workflows/publish.yml'
              - 'build/maw/Dockerfile'
              - 'cmd/mf-importer-maw/**'
              - 'go.mod'
              - 'go.sum'
              - 'internal/**'
            api:
              - '.dockerignore'
              - '.github/workflows/publish.yml'
              - 'build/api/Dockerfile'
              - 'cmd/mf-importer-api/**'
              - 'go.mod'
              - 'go.sum'
              - 'internal/**'
            metrics:
              - '.dockerignore'
              - '.github/workflows/publish.yml'
              - 'build/metrics/Dockerfile'
              - 'cmd/mf-importer-metrics/**'
              - 'go.mod'
              - 'go.sum'
              - 'internal/**'
            frontend:
              - '.dockerignore'
              - '.github/workflows/publish.yml'
              - 'build/fe/Dockerfile'
              - 'frontend/**'
```

- [ ] **Step 2: Gate every build job**

Add the matching dependency and condition directly below each build job name:

```yaml
  build_and_push:
    needs: changes
    if: ${{ github.ref_type == 'tag' || needs.changes.outputs.importer == 'true' }}

  build_and_push_maw:
    needs: changes
    if: ${{ github.ref_type == 'tag' || needs.changes.outputs.maw == 'true' }}

  build_and_push_api:
    needs: changes
    if: ${{ github.ref_type == 'tag' || needs.changes.outputs.api == 'true' }}

  build_and_push_metrics:
    needs: changes
    if: ${{ github.ref_type == 'tag' || needs.changes.outputs.metrics == 'true' }}

  build_and_push_fe:
    needs: changes
    if: ${{ github.ref_type == 'tag' || needs.changes.outputs.frontend == 'true' }}
```

Do not change the five image names, Dockerfiles, registry credentials, platforms, or push settings.

- [ ] **Step 3: Restrict short-SHA tags to master builds**

Replace the SHA rule in every metadata `tags:` block with:

```yaml
            type=sha,format=short,prefix=,enable=${{ github.ref_type == 'branch' }}
```

Keep all four existing semver rules unchanged. Branch events therefore produce the bare short SHA, while tag events produce only the existing semver tags.

- [ ] **Step 4: Inspect the implementation diff**

Run:

```bash
git diff --check
git diff -- .github/workflows/publish.yml
rg -c "dorny/paths-filter@ceb8a2b8f2d89434be7ff52d3de7ec3738c5cc9d" .github/workflows/publish.yml
rg -c "needs: changes" .github/workflows/publish.yml
rg -c "type=sha,format=short,prefix=,enable=" .github/workflows/publish.yml
```

Expected: no whitespace errors; the three counts are `1`, `5`, and `5`.

- [ ] **Step 5: Commit the workflow**

```bash
git add .github/workflows/publish.yml
git diff --cached --stat
git diff --cached
git commit -m "ci: publish only affected images"
```

Expected: only `.github/workflows/publish.yml` is included in this commit.

### Task 2: Validate the completed branch

**Files:**
- Verify: `.github/workflows/publish.yml`

- [ ] **Step 1: Validate GitHub Actions syntax and expressions**

Run:

```bash
nix develop -c actionlint .github/workflows/publish.yml
```

Expected: exit status 0 with no diagnostics. If `actionlint` is not provided by the development shell, report the missing tool and continue with the static checks below.

- [ ] **Step 2: Verify all change-to-image mappings**

Run:

```bash
rg -n "^  changes:|^  build_and_push|needs: changes|if:.*github.ref_type|type=sha|type=semver|build/.+Dockerfile|cmd/mf-importer|frontend/|internal/|go.mod|go.sum|\.dockerignore" .github/workflows/publish.yml
```

Expected:

- Each of the five build jobs has one `needs: changes` and the matching output condition.
- All five conditions allow tag events, so `v*` builds every image.
- All five SHA rules are enabled only for branch events.
- Shared Go paths appear in all four Go filters and never in the frontend-only filter.
- `.dockerignore` and the workflow path appear in all five filters.

- [ ] **Step 3: Run the repository suite**

Run:

```bash
env -u GOROOT -u GOTOOLDIR make test
```

Expected: `gofmt -l` emits no files, vet and staticcheck succeed, and all Go tests pass.

- [ ] **Step 4: Review branch state**

Run:

```bash
git status --short
git log --oneline --decorate -8
```

Expected: the worktree is clean, the implementation commit is on `ci/publish-on-master`, and no commit was made directly to `master`.
