# Community Setup Guide

This guide helps you set up GitHub features to improve community engagement and contributions.

## 🏷️ GitHub Labels Setup

Apply the predefined labels to organize issues and PRs:

```bash
# Install GitHub CLI if not already installed
# https://github.com/cli/cli#installation

# Apply labels from configuration
gh label sync --file .github/labels.yml

# Or apply individual labels manually via GitHub web interface
```

## 📝 Repository Settings

### Branch Protection Rules
Set up branch protection for `master`:

1. Go to Settings → Branches
2. Add rule for `master` branch
3. Enable:
   - ✅ Require pull request reviews before merging
   - ✅ Require status checks to pass before merging
   - ✅ Require branches to be up to date before merging
   - ✅ Include administrators

### Issue Templates
The issue templates are now configured in `.github/ISSUE_TEMPLATE/`:
- `bug_report.yml` - Structured bug reports
- `feature_request.yml` - Feature requests with categories
- `config.yml` - Disables blank issues, adds helpful links

## 🎯 Good First Issues

Mark beginner-friendly issues with the `good first issue` label:

```bash
# Example: Label an issue as good for beginners
gh issue edit 123 --add-label "good first issue"

# Add helpful description
gh issue edit 123 --body "$(cat <<EOF
This is a great issue for first-time contributors!

**What needs to be done:** [clear description]

**Files to look at:** 
- \`path/to/relevant/file.go\`

**Additional context:** [any helpful context]

Feel free to ask questions in the comments!
EOF
)"
```

## 🏆 Contributor Recognition

### All Contributors
Consider setting up [All Contributors](https://allcontributors.org/) to recognize all types of contributions:

```bash
# Install the CLI
npm install -g all-contributors-cli

# Initialize in your project
all-contributors init

# Add contributors
all-contributors add username code,doc,test
```

### GitHub Sponsors
You already have sponsor buttons set up. Consider:
- Adding specific funding goals
- Creating sponsor tiers with benefits
- Regular sponsor updates

## 📊 Project Insights

Enable useful GitHub features:
1. **Insights tab** - View contributor statistics, traffic, etc.
2. **Projects** - Create project boards for roadmap planning
3. **Wiki** - Enable for community-contributed documentation
4. **Releases** - Use for changelog and download tracking

## 🤖 Automation Ideas

Consider GitHub Actions for:

### Issue Management
```yaml
# .github/workflows/issue-management.yml
name: Issue Management
on:
  issues:
    types: [opened]
jobs:
  label-new-issues:
    runs-on: ubuntu-latest
    steps:
      - name: Add triage label
        run: gh issue edit ${{ github.event.issue.number }} --add-label "triage"
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
```

### Stale Issue Management
```yaml
# .github/workflows/stale.yml
name: Mark stale issues
on:
  schedule:
    - cron: "0 0 * * *"
jobs:
  stale:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/stale@v8
        with:
          days-before-stale: 30
          days-before-close: 7
          stale-issue-message: "This issue is stale and will be closed soon..."
```

## 📈 Community Growth Tips

1. **Be responsive** - Reply to issues and PRs promptly
2. **Welcome newcomers** - Be friendly and helpful to first-time contributors
3. **Create documentation** - Good docs attract more users and contributors
4. **Share progress** - Regular updates keep the community engaged
5. **Ask for help** - Use "help wanted" labels when you need community support

## ✅ Next Steps Checklist

After setting up these files, complete these tasks:

- [ ] Apply GitHub labels: `gh label sync --file .github/labels.yml`
- [ ] Set up branch protection rules for master
- [ ] Create first "good first issue" and label it appropriately
- [ ] Announce the new community features in a GitHub release
- [ ] Update CONTRIBUTING.md to reference the new issue templates
- [ ] Consider setting up GitHub Projects for roadmap planning

Your project now has professional community management tools! 🎉