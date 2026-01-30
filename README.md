# Squash Tree

Build and visualize squash trees from Git notes metadata. Records squash relationships so you can inspect composition and (later) unsquash.

## Prerequisites

- Go 1.21 or later
- Git in PATH

## Building

```bash
go build -o git-squash-tree ./cmd/git-squash-tree
```

Put the binary on your PATH (e.g. `export PATH="$PATH:$(pwd)"` or copy to `~/bin`). Git invokes the binary as `git-squash-tree`, so the executable must be named `git-squash-tree`.

## Usage

### One-time setup (recommended)

Install hooks so squashes are recorded automatically:

```bash
git squash-tree init
```

Or for all repos:

```bash
git squash-tree init --global
```

Then use Git as usual. After a squash (e.g. `git rebase -i`, `git merge --squash`), view the tree:

```bash
git squash-tree <commit>
```

Example:

```bash
git squash-tree HEAD
git squash-tree abc1234
```

### Without init (manual metadata)

If you did not run `init`, you can add metadata manually after a squash using the helper script (from the squash-tree repo):

```bash
./scripts/add-squash-metadata.sh <squash-commit> <base-commit> <child1> [<child2> ...]
```

Or use the CLI:

```bash
git squash-tree add-metadata --root=<squash> --base=<base> --children=<c1>,<c2>,<c3>
```

## Testing (manual)

1. **Build and PATH**
   ```bash
   go build -o git-squash-tree ./cmd/git-squash-tree
   export PATH="$PATH:$(pwd)"
   ```

2. **Create test repo** (from repo root; creates `test-repo/`, which is in `.gitignore`)
   ```bash
   ./test-setup.sh
   ```
   The script prints the squash commit hash (e.g. `SQUASH_COMMIT`).

3. **Show the tree**
   ```bash
   cd test-repo
   git squash-tree <squash-commit-hash>
   ```

4. **Optional: test init and hooks**  
   In the same repo, run `git squash-tree init`, then do another squash (e.g. `git reset --soft <base>` and `git commit -m "Squash"`), and run `git squash-tree HEAD`.

## Project structure

```
squash-tree/
├── cmd/git-squash-tree/   # CLI (show tree, init, add-metadata)
├── internal/
│   ├── git/               # Notes read/write, repo detection
│   ├── metadata/         # Metadata parsing and validation
│   └── tree/             # Tree build and visualization
├── scripts/
│   └── add-squash-metadata.sh   # Manual metadata helper
├── test-setup.sh         # Creates test-repo with squash metadata
├── requirements.md       # Design doc
└── README.md
```

## How it works

- Metadata is stored in Git notes under `refs/notes/squash-tree` (JSON: root, base, children with order).
- `git squash-tree <commit>` reads notes and prints a text tree.
- `init` installs Git hooks (pre-rebase, post-rewrite, post-merge, prepare-commit-msg) that call `git squash-tree add-metadata` when a squash is detected.

Design details: [requirements.md](requirements.md).
