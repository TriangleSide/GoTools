## Pull Requests

- Before creating a new branch, follow these steps one at a time:
    1. `git stash --all`
    2. `git checkout main`
    3. `git pull origin main`
    4. `git stash pop`
    5. Resolve any merge conflicts.
- Use `git checkout -b <short-description>` to create a new branch.
- Use `git commit -m "<message>"` with a very terse commit message summarizing the change at a high level.
- Commit messages should end with a period.
- Do not co-author commits; keep the current git user as the sole author.
- Push the branch to remote with `git push -u origin <branch_name>`.
- Use `gh pr create --base main --fill` to open pull requests against the main branch.
- Use `git checkout main` to go to the main branch after the PR is created.