# Turbolift Campaign for Data Governance

This repository provides a step-by-step guide to completing the actions required for the **Turbolift Campaign for Data Governance**.


## Steps

### 1. Navigate to the working directory
```bash
cd data_categorisation
```


### 2. Verify repository list
The file `cf_repos.txt` already includes all Skyscanner repositories containing data storage definitions targeted in this campaign.  
If you need to refresh the list, run:
```bash
./scripts/cf_search.sh
```


### 3. Filter repositories by team
The campaign is run per team. To generate a list of repositories owned by your team that contain CloudFormation data storage definitions, run:
```bash
./scripts/team_search.sh <your-team-name>
```


### 4. Clone repositories
Use Turbolift to clone your team’s repositories locally:
```bash
turbolift clone --repos repo_lists/team_cf_repos.txt
```


### 5. Apply campaign changes
Execute the script to make the necessary changes:
```bash
./scripts/turbolift_cf_changes.sh
```


### 6. Commit changes
Commit your updates with a standardized commit message:
```bash
turbolift commit --repos repo_lists/team_cf_repos.txt --message "feat: Data Governance tags added"
```


### 7. Push changes
Push the committed changes and create pull requests:
```bash
turbolift create-prs --repos repo_lists/team_cf_repos.txt
```


### 8. Update Github Project for Tracking
Update Github Project to track the pull requests:
```bash
./scripts/gh_project_import.sh
```
```bash
./scripts/gh_project_update_state.sh
```
```bash
./scripts/gh_project_update_owners.sh
```


### 9. Update pull request descriptions (if needed)
If you need to update the PR description after the PRs have been created, run:
```bash
turbolift update-prs --amend-description --repos repo_lists/team_cf_repos.txt
```

