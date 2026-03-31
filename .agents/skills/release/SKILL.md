---
name: release
description: 'Automated release workflow for AI-K8S-OPS. Use when user wants to release a new version (release/beta). Steps: (1) Ask for release type, (2) Query current tags, (3) Calculate next version, (4) Generate changelog from commits, (5) Update CHANGELOG.md, (5.5) Update pkg/version/version.go, (6) Commit changes, (7) Create tag and push. Supports both stable releases and beta releases.'
license: MIT
allowed-tools: Bash
---

# Release Automation for AI-K8S-OPS

Automated release workflow that handles version calculation, changelog generation, and tag creation.

## Supported Release Types

| Type | Tag Format | Example | Use Case |
|------|------------|---------|----------|
| `release` | `v{major}.{minor}.{patch}` | `v1.3.0` | Stable releases |
| `beta` | `v{major}.{minor}.{patch}-beta.{n}` | `v1.3.0-beta.12` | Pre-release versions |

## Workflow

### Step 1: Determine Release Type

Ask user if they want to release `release` (stable) or `beta`:

```
Which release type?
1. release - Stable version (v1.3.0)
2. beta - Pre-release version (v1.3.0-beta.12)
```

### Step 2: Query Current Tags

```bash
# Get all tags sorted by version
git tag -l "v*" | sort -V

# Get the latest stable (non-beta) release tag
git tag -l "v*" | grep -v "beta" | sort -V | tail -1
```

### Step 3: Calculate Next Version

Always base version calculation on the latest **stable release** tag, not the latest overall tag.

**For release (stable):**
- Get latest stable tag (e.g., `v1.3.0`), increment patch -> `v1.3.1`
- Or ask for bump type (patch/minor/major)

**For beta:**
- Get latest stable tag (e.g., `v1.3.0`), determine next patch -> `v1.3.1`
- Check if any `v1.3.1-beta.*` tags already exist
- If none exist -> `v1.3.1-beta.0`
- If `v1.3.1-beta.2` is the latest -> `v1.3.1-beta.3`

```bash
# Get latest stable (non-beta) release tag
latest_release=$(git tag -l "v*" | grep -v "beta" | sort -V | tail -1)
if [ -z "$latest_release" ]; then
    latest_release="v0.0.0"
fi

# Parse version components from latest stable release
release_version=${latest_release#v}
if [[ "$release_version" =~ ^([0-9]+)\.([0-9]+)\.([0-9]+)$ ]]; then
    major="${BASH_REMATCH[1]}"
    minor="${BASH_REMATCH[2]}"
    patch="${BASH_REMATCH[3]}"
fi

# Calculate next version based on release type
if [ "$release_type" == "beta" ]; then
    # Beta always bases off latest stable release + patch bump
    next_patch=$((patch + 1))
    beta_base="${major}.${minor}.${next_patch}"

    # Check if any beta tags already exist for this version
    latest_beta=$(git tag -l "v${beta_base}-beta.*" | sort -V | tail -1)

    if [ -n "$latest_beta" ]; then
        # Extract and increment beta number
        beta_num=$(echo "$latest_beta" | sed "s/v${beta_base}-beta\.//")
        next_beta=$((beta_num + 1))
        next_version="${beta_base}-beta.${next_beta}"
    else
        # Start new beta series at beta.0
        next_version="${beta_base}-beta.0"
    fi
else
    # Stable release: increment patch by default
    next_patch=$((patch + 1))
    next_version="${major}.${minor}.${next_patch}"
fi
```

### Step 4: Generate Changelog from Commits

Get commits since last tag:

```bash
# Get the latest tag (including betas) for changelog generation
latest_tag=$(git tag -l "v*" | sort -V | tail -1)
if [ -z "$latest_tag" ]; then
    latest_tag=""
fi

# Get commits since last tag (excluding merge commits)
if [ -n "$latest_tag" ]; then
    git log ${latest_tag}..HEAD --pretty=format:"%s" --no-merges
else
    git log --pretty=format:"%s" --no-merges
fi

# Categorize commits
# feat: -> Added
# fix: -> Fixed
# chore:, docs:, style:, refactor:, perf:, test: -> Changed
```

Parse commit messages and categorize:

```python
import re
from collections import defaultdict

def parse_commits(commits_text):
    categories = defaultdict(list)

    for line in commits_text.strip().split('\n'):
        if not line:
            continue

        # Skip release commits
        if line.startswith('chore: release') or line.startswith('chore(release)'):
            continue

        # Parse conventional commit format
        match = re.match(r'^(\w+)(\([^)]+\))?!?:\s*(.+)', line)
        if match:
            type_ = match.group(1)
            message = match.group(3)
        else:
            type_ = 'chore'
            message = line

        # Map to changelog categories
        if type_ == 'feat':
            categories['Added'].append(message)
        elif type_ == 'fix':
            categories['Fixed'].append(message)
        else:
            categories['Changed'].append(message)

    return categories
```

### Step 5: Update CHANGELOG.md

Read existing CHANGELOG.md and insert new version at the top:

```python
from datetime import datetime
from collections import defaultdict

def update_changelog(next_version, categories):
    date = datetime.now().strftime('%Y-%m-%d')

    # Build new entry
    lines = [f'## [{next_version}] - {date}\n']

    for category in ['Added', 'Changed', 'Fixed']:
        if category in categories and categories[category]:
            lines.append(f'\n### {category}\n')
            for item in categories[category]:
                lines.append(f'- {item}\n')

    new_entry = ''.join(lines)

    # Read existing changelog
    with open('CHANGELOG.md', 'r') as f:
        content = f.read()

    # Insert after header (find first ## line)
    insert_pos = content.find('## [')
    if insert_pos == -1:
        insert_pos = len(content)

    updated = content[:insert_pos] + new_entry + '\n' + content[insert_pos:]

    with open('CHANGELOG.md', 'w') as f:
        f.write(updated)
```

### Step 5.5: Update pkg/version/version.go

Update the `Version` variable in `pkg/version/version.go` to match the new version:

```bash
# For both stable and beta releases, update the Version constant
sed -i '' "s/Version   = \"[^\"]*\"/Version   = \"${next_version}\"/" pkg/version/version.go
```

Note: `BuildTime` and `GitCommit` are injected at build time via `-ldflags` in the GitHub Actions workflow and Makefile, so they should remain as `"unknown"` in source code.

#### Verification

```bash
# Verify the update
grep "Version" pkg/version/version.go
```

### Step 6: Commit Changes

```bash
# Stage all changes (CHANGELOG.md and version.go)
git add -A

# Commit with version message
git commit -m "chore: release v${next_version}

- Update changelog
- Update version to ${next_version}"
```

### Step 7: Create Tag and Push

```bash
# Create annotated tag
git tag -a "v${next_version}" -m "Release v${next_version}"

# Push commit and tag
git push origin main
git push origin "v${next_version}"
```

## Complete Script

```bash
#!/bin/bash
set -e

echo "AI-K8S-OPS Release Automation"
echo "=============================="

# Step 1: Determine release type
echo ""
echo "Select release type:"
echo "1) release - Stable version"
echo "2) beta - Pre-release version"
read -p "Enter choice (1 or 2): " choice

if [ "$choice" == "1" ]; then
    release_type="release"
elif [ "$choice" == "2" ]; then
    release_type="beta"
else
    echo "Invalid choice"
    exit 1
fi

# Step 2: Query current tags
# Always get the latest stable (non-beta) release tag
latest_release=$(git tag -l "v*" | grep -v "beta" | sort -V | tail -1)

if [ -z "$latest_release" ]; then
    echo "No existing stable tags found, using v0.0.0"
    latest_release="v0.0.0"
fi

# Also get the overall latest tag for changelog generation
latest_tag=$(git tag -l "v*" | sort -V | tail -1)
if [ -z "$latest_tag" ]; then
    latest_tag=""
fi

echo ""
echo "Latest stable release: $latest_release"
echo "Latest tag: ${latest_tag:-none}"

# Step 3: Calculate next version
# Always base version calculation on the latest STABLE release tag
release_version=${latest_release#v}

# Parse stable version
if [[ "$release_version" =~ ^([0-9]+)\.([0-9]+)\.([0-9]+)$ ]]; then
    major="${BASH_REMATCH[1]}"
    minor="${BASH_REMATCH[2]}"
    patch="${BASH_REMATCH[3]}"
else
    echo "Could not parse version: $release_version"
    exit 1
fi

# Calculate next version
if [ "$release_type" == "beta" ]; then
    # Beta always bases off latest stable release + patch bump
    next_patch=$((patch + 1))
    beta_base="${major}.${minor}.${next_patch}"

    # Check if any beta tags already exist for this version
    latest_beta=$(git tag -l "v${beta_base}-beta.*" | sort -V | tail -1)

    if [ -n "$latest_beta" ]; then
        # Extract and increment beta number
        beta_num=$(echo "$latest_beta" | sed "s/v${beta_base}-beta\.//")
        next_beta=$((beta_num + 1))
        next_version="${beta_base}-beta.${next_beta}"
    else
        # Start new beta series at beta.0
        next_version="${beta_base}-beta.0"
    fi
else
    # Stable release: ask for bump type
    echo ""
    echo "Bump type for stable release (current: ${major}.${minor}.${patch}):"
    echo "1) patch -> ${major}.${minor}.$((patch + 1))"
    echo "2) minor -> ${major}.$((minor + 1)).0"
    echo "3) major -> $((major + 1)).0.0"
    read -p "Enter choice (1/2/3, default 1): " bump_choice

    case "$bump_choice" in
        2) next_version="${major}.$((minor + 1)).0" ;;
        3) next_version="$((major + 1)).0.0" ;;
        *) next_version="${major}.${minor}.$((patch + 1))" ;;
    esac
fi

echo ""
echo "Next version: v${next_version}"
read -p "Confirm? (y/n): " confirm
if [ "$confirm" != "y" ]; then
    echo "Aborted"
    exit 0
fi

# Step 4: Generate changelog
echo ""
echo "Generating changelog..."

# Get commits since last tag
if [ -n "$latest_tag" ]; then
    commits=$(git log ${latest_tag}..HEAD --pretty=format:"%s" --no-merges)
else
    commits=$(git log --pretty=format:"%s" --no-merges)
fi

if [ -z "$commits" ]; then
    echo "No new commits since ${latest_tag:-beginning}"
    read -p "Continue anyway? (y/n): " continue_anyway
    if [ "$continue_anyway" != "y" ]; then
        exit 0
    fi
fi

# Step 5: Update CHANGELOG.md
python3 << PYTHON
import re
from collections import defaultdict
from datetime import datetime

commits_text = """${commits}"""
next_version = "${next_version}"

def parse_commits(text):
    categories = defaultdict(list)

    for line in text.strip().split('\n'):
        if not line:
            continue

        # Skip release commits
        if line.startswith('chore: release') or line.startswith('chore(release)'):
            continue

        # Parse conventional commit
        match = re.match(r'^(\w+)(?:\([^)]+\))?!?:\s*(.+)', line)
        if match:
            type_ = match.group(1)
            message = match.group(2)
        else:
            type_ = 'chore'
            message = line

        # Categorize
        if type_ == 'feat':
            categories['Added'].append(message)
        elif type_ == 'fix':
            categories['Fixed'].append(message)
        elif type_ in ['chore', 'docs', 'style', 'refactor', 'perf', 'test', 'build', 'ci']:
            categories['Changed'].append(message)
        else:
            categories['Changed'].append(message)

    return categories

categories = parse_commits(commits_text)

# Build new entry
date = datetime.now().strftime('%Y-%m-%d')
lines = [f'## [{next_version}] - {date}\n']

for category in ['Added', 'Changed', 'Fixed']:
    if category in categories and categories[category]:
        lines.append(f'\n### {category}\n')
        for item in categories[category]:
            lines.append(f'- {item}\n')

new_entry = ''.join(lines)

# Read and update changelog
try:
    with open('CHANGELOG.md', 'r') as f:
        content = f.read()
except FileNotFoundError:
    content = '# Changelog\n\n'

# Insert after the Changelog header
insert_pos = content.find('## [')
if insert_pos == -1:
    insert_pos = len(content)

updated = content[:insert_pos] + new_entry + '\n' + content[insert_pos:]

with open('CHANGELOG.md', 'w') as f:
    f.write(updated)

print(f"Updated CHANGELOG.md with version {next_version}")
PYTHON

# Step 5.5: Update pkg/version/version.go
echo ""
echo "Updating pkg/version/version.go..."
sed -i '' "s/Version   = \"[^\"]*\"/Version   = \"${next_version}\"/" pkg/version/version.go
echo "Version set to: ${next_version}"
grep "Version" pkg/version/version.go

# Step 6: Commit
echo ""
echo "Committing changes..."
git add -A
git commit -m "chore: release v${next_version}

- Update changelog
- Update version to ${next_version}"

# Step 7: Create tag and push
echo ""
echo "Creating tag v${next_version}..."
git tag -a "v${next_version}" -m "Release v${next_version}"

echo ""
echo "Pushing to origin..."
git push origin main
git push origin "v${next_version}"

echo ""
echo "Successfully released v${next_version}"
echo ""
echo "GitHub Actions workflow will now:"
echo "  1. Build Go binaries for linux/darwin (amd64/arm64)"
echo "  2. Build frontend SPA"
echo "  3. Package release tarballs"
echo "  4. Create GitHub Release with assets"
if [ "$release_type" == "beta" ]; then
    echo "  5. Mark as pre-release"
fi
```

## Post-Release

After pushing the tag:

1. GitHub Actions will automatically trigger the release workflow (`.github/workflows/release.yml`)
2. The workflow will:
   - Build Go binaries for multiple platforms (linux/amd64, linux/arm64, darwin/amd64, darwin/arm64)
   - Build the frontend SPA
   - Inject version, build time, and git commit into binaries via `-ldflags`
   - Package each platform into a tarball containing the server binary + CLI + frontend assets + config
   - Create a GitHub Release (marked as prerelease for beta tags)
   - Upload all tarballs and SHA256 checksums as release assets

## Build Artifacts

The release produces the following artifacts:

| Platform | Artifact |
|----------|----------|
| Linux amd64 | `ai-k8s-ops-{version}-linux-amd64.tar.gz` |
| Linux arm64 | `ai-k8s-ops-{version}-linux-arm64.tar.gz` |
| macOS amd64 | `ai-k8s-ops-{version}-darwin-amd64.tar.gz` |
| macOS arm64 | `ai-k8s-ops-{version}-darwin-arm64.tar.gz` |

Each tarball contains:
```
ai-k8s-ops-{version}-{os}-{arch}/
  bin/server           # API server binary
  bin/ai-k8s-ops       # CLI tool binary
  frontend-dist/       # Built frontend assets
  configs/config.example.yaml
  scripts/             # Deployment scripts
  README.md
```

## Manual Steps (When Not Using Script)

If you need to do this manually without the script:

1. **Check current tags**: `git tag -l | sort -V`
2. **Calculate next version** based on rules above
3. **Get commits**: `git log v1.3.0..HEAD --pretty=format:"%s" --no-merges`
4. **Edit CHANGELOG.md**: Add new section at top
5. **Update version.go**: `sed -i '' 's/Version   = "[^"]*"/Version   = "1.3.1"/' pkg/version/version.go`
6. **Commit**: `git add -A && git commit -m "chore: release v1.3.1"`
7. **Tag**: `git tag -a v1.3.1 -m "Release v1.3.1"`
8. **Push**: `git push && git push --tags`
