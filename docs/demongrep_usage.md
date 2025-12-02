# Demongrep Usage Guidelines

`demongrep` is a high-performance codebase exploration tool designed to be more token-efficient than standard tools. This guide focuses on using `demongrep` via the Command Line Interface (CLI).

> *Documented for demongrep v0.1.0. Run `demongrep --help` to see current options.*

## Quick Start

Ensure you are in the root of your repository.

1.  **Index the codebase:**
    ```bash
    demongrep index
    ```
    *Run this whenever files change significantly.*

2.  **Search:**
    ```bash
    demongrep search "your query here"
    ```

## Core Commands

### 1. Indexing (`index`)
Builds the vector and keyword index for the current directory.
```bash
demongrep index
```
- **Arguments:**
    - `[PATH]`: Directory to index (default: current). Example: `demongrep index ./backend`
- **Options:**
    - `-f, --force`: Force re-indexing even if files haven't changed.

### 2. Searching (`search`)
Performs a hybrid search (Keyword + Vector) by default.
```bash
demongrep search "query string"
```
- **Options:**
    - `-m, --max-results <n>`: Limit results (default: 25).
    - `--vector-only`: Use vector-only search (disable hybrid).
    - `--rerank`: Enable neural reranking (Jina Reranker) for better accuracy.
    - `--filter-path <path>`: Filter results to files under a path.
    - `-c, --content`: Show full chunk content.
    - `-s, --sync`: Force re-index changed files before searching.
    - `--json`: Output JSON format (useful for automation/agents).
    - `--compact`: Show file paths only (like `grep -l`).

### 3. Statistics (`stats`)
View information about the current index.
```bash
demongrep stats
```
Shows number of files, chunks, and database size.

### 4. Maintenance (`clear`, `doctor`)
- `demongrep clear`: Deletes the current index.
- `demongrep doctor`: Checks for common issues.

## Best Practices for Claude

### When to Use Demongrep (Primary)
- **Semantic searches:** "How does authentication work?" or "Where are errors handled?"
- **Exploring unfamiliar code:** Understanding architecture, patterns, or data flow
- **Finding related code:** Locating all code related to a concept (not just exact matches)

### When to Use Native Tools (Fallback)
Use Glob/Grep when:
- Demongrep returns no results or errors
- You need **exact pattern matching** (regex like `function\s+\w+`)
- Searching for **specific file names** or extensions (`*.test.ts`)
- Looking for **exact strings** (variable names, error messages)

### Search Tips
- **Index When Needed:** Run `demongrep index` after significant code changes
- **Be Specific:** "How does the auth middleware work?" is better than "auth"
- **Use Quotes:** Always wrap your query in quotes
- **Use `--rerank`:** For complex queries where accuracy matters more than speed
- **Filter by path:** Use `--filter-path backend/` to narrow results

## Setup Notes

### Generated Files
Demongrep creates these directories that should be gitignored:
- `.demongrep.db/` - Vector database and index files
- `.fastembed_cache/` - Cached embedding model files

These are already added to `.gitignore` in this project.

## Troubleshooting

| Issue | Solution |
|-------|----------|
| No search results | Run `demongrep index` to rebuild the index |
| Stale results | Run `demongrep index --force` to force re-indexing |
| Slow first search | Normal - embedding model loads on first search |
| Index errors | Run `demongrep doctor` to diagnose issues |
| Corrupted index | Run `demongrep clear` then `demongrep index` |
