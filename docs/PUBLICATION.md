# Before publishing source

The project-specific email `jegati@proton.me`, GitHub project identity,
`jamgati.com`, and association with OVHcloud/Cloudflare are intentionally public.
Commit timestamps and useful benchmark hardware descriptions remain. Avoid personal
usernames, home-directory paths, machine names and purchasing/account narratives
in future documentation. Use portable paths and neutral deployment status.

After staging specific reviewed files, run:

```sh
make publication-check test-publication
git diff --cached --check
git diff --cached --stat
```

Review the staged diff locally as well. `publication-check` uses only Python's
standard library and Git; it makes no network requests and installs no tools.
It scans **the complete Git index**, including staged changes, not unstaged working
copies. It runs in CI and `make verify-local` too. CI checks the checked-out
snapshot; neither entry point scans intermediate commits in a proposed push.

The check flags personal home paths, private-key headers, common GitHub/AWS/Slack
credential signatures, credential-bearing URLs, and accidentally tracked runtime,
build or credential filenames. Findings contain filenames, line numbers and
categories, never matching content. Two exact synthetic URL rejection fixtures are
allowed in their specific test files; other content in those files is still checked.
Use placeholders such as `$HOME` and reserved example domains in documentation.
Resolve a flagged value before staging again; do not blindly add an exemption.

Binary contents are counted as skipped, with their filenames still checked. Review
PDF metadata and compressed/archive contents separately when adding or changing
them. This is a lightweight guard, not a complete secret scanner: arbitrary tokens,
encoded secrets, names, phone numbers, Git metadata and cross-site correlations may
escape detection. A passing result does not certify anonymity or authorize a push.
Deleting a value in a later commit does not remove it from earlier commits.

Local scratch notes can be excluded using Git's untracked local configuration.
This working checkout excludes `/commands.txt` in `.git/info/exclude`; the file is
preserved. Other contributors may add the same rule locally if needed. Local
exclusions are not shared in a clone and `git add -f` bypasses them; the publication
check also rejects `commands.txt` if it is staged. Share reviewed source exports,
not a zip of the whole working directory containing ignored runtime credentials.

The earlier private identity review is separate evidence. No history rewrite is
needed for the accepted public project associations, and none was performed.
