# 1) Quick VS Code + WSL setup checklist

Run these once in your WSL terminal:

* Verify Go: `go version` (Go ≥1.21 recommended).
* Open the repo from **WSL**: in VS Code use **Remote - WSL** → “Open Folder…” → `/home/mihir/code/go-gitlet`.
* Install extensions in VS Code (WSL side):

  * **Go** (by Go Team at Google)
  * **Remote - WSL**
* In the Command Palette → “Go: Install/Update Tools” and accept defaults (this installs gopls, dlv, etc.).
* Optional QA:

  * Enable format on save (Go extension handles `gofmt`/`goimports`).
  * Turn on “Go: Add Missing Imports & Remove Unused Imports on Save”.

---

# 2) Module + top-level repo layout

From the repo root (you can pick any module path; it can be local):

* Initialize modules: `go mod init github.com/<you>/go-gitlet`

Recommended tree (no code, just folders/files to create):

```
go-gitlet/
  cmd/
    gitlet/              # CLI entrypoint (main package lives here)
  internal/              # implementation details; not exported to other modules
    repo/                # high-level repository orchestration (opens .gitlet, runs commands)
    storage/             # content-addressable store (read/write objects on disk, hashing, paths)
    objects/             # definitions & helpers for Commit, Blob, (maybe) Tree “views”
    index/               # staging area representation & persistence
    refs/                # branch refs, HEAD, ref reading/writing
    walk/                # commit graph walking, LCA/split-point search
    merge/               # merge policy & conflict formatting
    status/              # diffing working dir vs index vs HEAD for `status`
    cli/                 # argument parsing & command routing (kept separate from main)
  test/
    integration/         # Go integration tests that drive the CLI as a subprocess
    fixtures/            # sample files like wug.txt, notwug.txt, etc.
  testdata/              # golden outputs or sample repositories (read-only by tests)
  docs/
    design.md            # your design doc (mandated by spec)
    invariants.md        # repo invariants, data-format notes, error message contracts
  scripts/               # helper scripts (e.g., local test runner)
  .gitignore
  README.md
```

Why this shape?

* `cmd/gitlet` holds the *app shell*; everything else is testable libraries.
* `internal/*` fences your implementation so only your CLI uses it.
* Separate `storage`, `objects`, `refs`, `index` mirrors the spec’s concerns and keeps merging/walking logic independent of disk I/O.

---

# 3) On-disk `.gitlet/` layout (persisted format)

Keep it simple and deterministic:

```
.gitlet/
  HEAD                   # "ref: refs/heads/master"
  refs/
    heads/
      master             # contains commit id (full SHA-1 hex)
      <branch>           # more branches
  objects/
    blobs/
      ab/cdef...         # split by first 2 hex chars to avoid huge dirs
    commits/
      12/3456...
  index                  # staging area state (see below)
  config                 # optional, if you want (e.g., user, timestamps mode)
  logs/                  # optional (not required by spec)
```

Notes:

* **Blobs**: raw file bytes stored by content hash (type-tagged; see below).
* **Commits**: serialized commit metadata (message, timestamp, parent(s), map filename→blob id).
* **Refs**: files that just contain a commit id (or a symbolic ref in `HEAD`).
* **Index**: your staging area file (track staged-for-add, staged-for-remove).

---

# 4) Content addressing & determinism (no code, just rules)

* Use **SHA-1** with a type tag to avoid collisions between kinds:

  * Example concept: `sha1("blob\n" + <file bytes>)` for blobs and `sha1("commit\n" + <canonical bytes>)` for commits.
* **Canonical commit bytes** must be stable across runs:

  * Sort filenames lexicographically before hashing/serializing.
  * If you serialize to JSON, ensure the order of fields is fixed and the file map is turned into a sorted slice first (Go maps are randomized).
  * Normalize timestamps (UTC, RFC3339 or your chosen format) and ensure the **initial commit** is the Unix epoch exactly.

---

# 5) Runtime model (packages + responsibilities)

* `objects`

  * Conceptual structures of **Blob** and **Commit** (ids are strings, not pointers).
  * Commit contains: message, timestamp, parent (and second parent for merges), and a **map of filename → blobId** (internally maintain as a sorted slice when serializing or hashing).
* `storage`

  * Read/write objects under `.gitlet/objects`, compute hashes, path sharding (`ab/cdef...`), atomic writes, safe file ops.
* `refs`

  * HEAD (symbolic ref), refs/heads/\* files. Helpers: read current branch, resolve to commit id, update branch head.
* `index`

  * Tracks staged additions (filename→blobId) and staged removals (set of filenames). Persist to `.gitlet/index`.
* `repo`

  * Orchestrates commands by composing `storage`, `refs`, `index`.
  * Enforces **spec-mandated error messages** and checks (exact strings, punctuation).
* `walk`

  * Graph traversal: `log`, `global-log`, ancestry marking, **split point (latest common ancestor)** for merges.
* `merge`

  * Implements the spec’s rules (auto-stage, conflict markers `<<<<<<<`, `=======`, `>>>>>>>`, etc.).
* `status`

  * Computes: branches (current marked `*`), staged files, removed files, and (extra credit) “modifications not staged” + “untracked”.

---

# 6) Command surface & invariants (behavioral checklist)

(no implementation, just what each layer must guarantee)

**init**

* Create `.gitlet`, write epoch **initial commit**, create `refs/heads/master`, set `HEAD` to it.
* Refuse if `.gitlet` exists (print exact error).

**add \[file]**

* If file missing → “File does not exist.”
* Hash working file, store blob if new, stage for add.
* If identical to HEAD version → unstage it (and remove from index if there).

**commit "\[message]"**

* Require message non-blank; require something staged (add or rm).
* New commit = parent’s file map ± staged deltas. Clear index.
* Update current branch ref to new commit. Do **not** edit working files.

**rm \[file]**

* If neither staged nor tracked → “No reason to remove the file.”
* Unstage if staged-for-add. If tracked, stage-for-remove and delete from working dir.

**log**

* Walk first-parent back to the initial commit.
* Print format exactly; for merges include `Merge: <first7> <first7>` line.

**global-log**

* Iterate all commits (scan `objects/commits`); order irrelevant.

**find "\[message]"**

* Print all matching commit ids or “Found no commit with that message.”

**status**

* Sorted lexicographically; `*` on current branch.
* Sections exactly as spec; extra credit sections can be left blank.

**checkout**

1. `-- [file]`: write HEAD’s version to working dir (no staging changes).
2. `[commit] -- [file]`: same but from that commit (allow abbreviated ids).
3. `[branch]`: write that commit’s full snapshot, delete files absent there, clear index, switch `HEAD` (protect against untracked-file clobber).

**branch \[name] / rm-branch \[name]**

* Create/delete ref files; guard error cases.

**reset \[commit]**

* Like full-branch checkout to an arbitrary commit; move current branch ref; clear index; untracked-file protection.

**merge \[branch]**

* Guard: no staged changes; branch exists; not merging branch into itself; untracked-file protection.
* Compute **split point** (latest common ancestor). Handle the fast-forward/ancestor-noop cases.
* Apply file rules from spec (auto-stage changed files; conflict markers where needed).
* Auto-create a **merge commit** with two parents and message `Merged <given> into <current>.` and then, if conflicts occurred, print `Encountered a merge conflict.`

---

# 7) Testing strategy (Go style, still no code)

**Unit tests** (small, fast):

* `internal/storage`: round-trip object read/write, hash determinism, atomic writes.
* `internal/objects`: commit id stability given same content, sorted file lists.
* `internal/refs`: HEAD/branch switch semantics.
* `internal/index`: persistence & set operations.

**Integration tests** (spec-level):

* In `test/integration`, each test:

  * Creates a fresh temp working dir with a temp `.gitlet`.
  * Copies fixture files (`test/fixtures/wug.txt`, `notwug.txt`, etc.).
  * Invokes the CLI as a subprocess (e.g., `gitlet init`, `add`, `commit`, …).
  * Captures stdout/stderr and compares to **golden files** stored under `testdata/…`.
* Handling nondeterminism:

  * **Freeze time**: have `repo` accept a clock interface so tests can pin timestamps (e.g., initial commit = epoch, others fixed).
  * **Commit ids**: if ids depend on timestamps, fixed clock makes ids stable. If you still want flexibility, allow regex-like placeholders in golden files (e.g., `COMMIT <[0-9a-f]{40}>`) and post-process outputs before compare.
* Coverage checklist:

  * Every error message path (exact punctuation).
  * Path protections (untracked-file overwrite).
  * Abbreviated ids (`checkout/reset`).
  * Merge conflict scenarios (same change vs different change vs delete vs add).
  * `status` list ordering.
  * Branch delete on current branch failure.
  * Reset and branch checkout clear the index.
  * `global-log` over arbitrary commit order.

(If you want a tester experience like 61B’s DSL, you can write a tiny harness that reads `.in` files and runs commands, but standard Go integration tests with goldens are usually simpler.)

---

# 8) Dev workflow in VS Code (no code JSON)

* **Run**: use the built-in Run/Debug button on `cmd/gitlet` (Program Args set to e.g., `init`).
* **Debug**: breakpoints in `internal/*` packages work when launched from `cmd/gitlet`.
* **Tasks**: a simple task to run `go test ./...` and one to run `go vet ./...` is handy.
* **Lint (optional)**: add staticcheck or golangci-lint; keep it strict on unused errors and deterministic map use.

---

# 9) Milestones (so you see progress quickly)

1. **Scaffold** folders + `go mod init` (done today).
2. **init** end-to-end (writes `.gitlet`, epoch commit, HEAD/ref).
3. **add** + **commit** (staging + persistence + id determinism).
4. **checkout file** (both variants), then **log** + **global-log** + **find**.
5. **status** (base sections first).
6. **branch**, **rm-branch**, **checkout branch**, **reset**.
7. **rm**.
8. **merge** (last; build LCA first in `walk`).

Write integration tests as you go—don’t wait for merge.

---

# 10) Gotchas to avoid (common 61B→Go translation issues)

* **Map order**: never hash or print by ranging over a map—always sort keys first.
* **Relative paths**: always prefer `filepath.Join` and path normalization; assume no subdirectories in the working dir per spec.
* **Atomic writes**: write to a temp file in the same directory, then rename.
* **Staging logic**: “identical to HEAD” means **blob id equal**, not mtime or size.
* **Exact messages**: error strings & trailing periods must match spec exactly.
* **Initial commit time**: true Unix epoch (watch your timezones in tests vs printing).

---

If you’re down, next step is super light:

* create the folders in the tree above,
* run `go mod init ...`,
* add your `docs/design.md` and jot the invariants you’ll enforce (commit-id determinism, sorted filenames everywhere, HEAD semantics, etc.),
* then we’ll tackle **init** behavior end-to-end with a tiny integration test.