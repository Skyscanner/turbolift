# turbolift

A simple tool to help apply changes across many GitHub repositories simultaneously.

## Philosophy

Anyone who has had to manually make changes to many GitHub repositories knows that it's hard to beat the simplicity of just cloning the repositories and updating them locally. You can use any tools necessary to make the change, and there's a degree of immediacy in having local files to inspect, tweak or run validation.

It's dumb but it works. It doesn't scale well, though. Manually cloning and raising PRs against tens/hundreds of repositories is painful and boring.

Turbolift essentially automates the boring parts and stays out of the way when it comes to actually making the changes. It automates cloning, committing, and raising PRs en-masse, so that you can focus on the substance of the change.

> Historical note: Turbolift supersedes an internal system at Skyscanner named Codelift. Codelift was a centralised batch system, requiring changes to be scripted upfront and run overnight. While Codelift was useful, we have found that a decentralised, interactive tool is far easier and quicker for people to use in practice.
>
> [This blog post](https://medium.com/@SkyscannerEng/turbolift-a-tool-for-refactoring-at-scale-70603314f7cc) gives a longer background for the thinking behind Turbolift.

## Demo

This demo shows Turbolift in action, creating a simple PR in two repositories:

![Screencast demo of turbolift in use](docs/demo.gif "Screencast demo of turbolift in use")

## Installation

<details>
<summary>Using brew (recommended)</summary>
Install turbolift using brew from Skyscanner's tap, as follows:

```shell
brew install skyscanner/tools/turbolift
```

Note that the GitHub CLI, `gh` is a dependency of Turbolift and will be installed automatically.

</details>

<details>
<summary>Downloading binaries</summary>

Pre-built binary archives can be downloaded from the [Releases](https://github.com/Skyscanner/turbolift/releases) page.

* Download, extract the archive, and move it onto your `PATH`.
* Note that the binaries are not currently notarized for MacOS Gatekeeper. If errors are displayed, use `xattr -c PATH_TO_TURBOLIFT_BINARY` to un-quarantine the binary, or right-click on the binary in Finder and choose 'Open' once to allow future execution. Distribution will be improved under https://github.com/Skyscanner/turbolift/issues/43.

You must also have the GitHub CLI, `gh`, installed:

* Install using `brew install gh`
</details>

> Before using Turbolift, run `gh auth login` once and follow the prompts, to authenticate against github.com and/or your GitHub Enterprise server.

## Basic usage:

Making changes with turbolift is split into six main phases:

1. `init` - getting set up
2. Identifying the repos to operate upon
3. Running a mass `clone` of the repos
4. Making changes to every repo
5. Committing changes to every repo
6. Creating a PR for every repo

It is expected that you'll go through these phases in series.

The turbolift tool automates as much of the Git/GitHub heavy lifting as possible, but leaves you to use whichever tools are appropriate for making the actual changes.

## Caveats

With great power comes great responsibility. We encourage Turbolift users to consider the following guidelines:

* Don't use Turbolift to raise pointless PRs. If a reviewer might think the change is trivial or unimportant, think about whether it's actually needed.
* If you need to make a change to a large number of repositories, we've found that it's generally better to raise PRs to a small subset at first and collect feedback. Simply comment out repositories in `repos.txt` to make Turbolift temporarily ignore them.
* For complicated or potentially contentious changes, think about ways to validate them before raising PRs. This could range from working in a pair, through writing a peer-reviewed script, all the way to preparing a design document for the planned changes.
* If you can run automated tests locally, then do (e.g. `turbolift foreach ` to run linting and tests for each repository).
* Raising draft PRs can be a good way to collect feedback, especially CI test results, with less pressure on reviewers. Use `turbolift create-prs --draft`
* In an organisation with shared infrastructure (e.g. CI), raising many PRs in a short timeframe can cause a lot of load. Consider spacing out PR creation using the `--sleep` option or by commenting out chunks of repositories in `repos.txt`.

## Detailed usage

### `init` - getting set up

As per the installation instructions above, make sure `gh` is installed and authenticated before starting.

If working with repositories on a GitHub Enterprise server, ensure that you have the environment variable `GH_HOST` set to point to that server.

To begin working with Turbolift and create a 'campaign' to hold settings and working copies of repositories:

```turbolift init --name CAMPAIGN_NAME```

This creates a new turbolift 'campaign' directory ready for you to work in.
Note that `CAMPAIGN_NAME` will be used as the branch name for any changes that are created

Next, please run:

```cd CAMPAIGN_NAME```

## Identifying the repos to operate upon

Update repos.txt with the names of the repos that need changing (either manually or using a tool to identify the repos).

[gh-search](https://github.com/janeklb/gh-search) is an excellent tool for performing GitHub code searches, and can output a list of repositories in a format that `turbolift` understands:

e.g.
```
$ gh-search --repos-with-matches YOUR_GITHUB_CODE_SEARCH_QUERY > repos.txt
```

### Working on multiple repo files

Occasionally you may need to work on different repo files. For instance the repos can be divided in sub categories and the same change don't apply to them the same way. 
The default repo file is called `repos.txt` but you can override this on any command with the `--repos` flag.

```console
turbolift foreach --repos repoFile1.txt -- sed 's/pattern1/replacement1/g'
turbolift foreach --repos repoFile2.txt -- sed 's/pattern2/replacement2/g'
```

### Running a mass `clone`

`turbolift clone` clones all repositories listed in the `repos.txt` file into the `work` directory.
By default the cloning policy is to create a branch to the target repository. If you do not have permissions to push a branch on the target repository, `turbolift` will fork it.

If you do want to fork all the repositories instead of letting turbolift deciding for you, use the `--fork` flag.

Usage:
```console
turbolift clone
```

### Making changes

Now, make changes to the checked-out repos under the `work` directory.
You can do this manually using an editor, using `sed` and similar commands, or using [`codemod`](https://github.com/facebook/codemod)/[`comby`](https://comby.dev/), etc.

**You are free to use any tools that help get the job done.**

If you wish to, you can run the same command against every repo using `turbolift foreach -- ...` (where `...` is the command you want to run).

For example, you might choose to:

* `turbolift foreach -- rm somefile` - to delete a particular file
* `turbolift foreach -- sed -i '' 's/foo/bar/g' somefile` - to find/replace in a common file
* `turbolift foreach -- make test` - for example, to run tests (using any appropriate command to invoke the tests)
* `turbolift foreach -- git add somefile` - to stage a file that you have created
* `turbolift foreach -- sh -c 'grep needle haystack.txt > output.txt'` - use a shell to run a command using redirection

Remember that when the command runs the working directory will be the
repository root. If you want to refer to files from elsewhere you need
to provide an absolute path. You may find the `pwd` command helpful here.
For example, to run a shell script from the current directory against
each repository:

```
turbolift foreach -- sh "$(pwd)/script.sh"
```

At any time, if you need to update your working copy branches from the upstream, you can run `turbolift foreach -- git pull upstream master`.

It is highly recommended that you run tests against affected repos, if it will help validate the changes you have made.

#### Logging and re-running with foreach

Every time a command is run with `turbolift foreach`, logging output for each repository is collected in a temporary directory
with the following structure:

```
temp-dir
   \ successful
       \ repos.txt        # a list of repos where the command succeeded
       \ org
           \ repo
               \ logs.txt # logs from the specific foreach execution on this repo
           ....
   \ failed
       \ repos.txt        # a list of repos where the command succeeded
       \ org
           \ repo
               \ logs.txt # logs from the specific foreach execution on this repo
```

You can use `--successful` or `--failed` to run a foreach command only against the repositories that succeeded or failed in the preceding foreach execution.

```
turbolift foreach --failed -- make test
```

### Committing changes

When ready to commit changes across all repos, run:

```turbolift commit --message "Your commit message"```

This command is a no-op on any repos that do not have any changes.
Note that the commit will be run with the `--all` flag set, meaning that it is not necessary to stage changes using `git add/rm` for changed files.
Newly created files _will_ still need to be staged using `git add`.

Repeat if you want to make multiple commits.

### Creating PRs

Edit the PR title and description in `README.md`.

Next, to push and raise PRs against changed repos, run:

```turbolift create-prs```

Use `turbolift create-prs --sleep 30s` to, for example, force a 30s pause between creation of each PR. This can be helpful in reducing load on shared infrastructure.

> Important: if raising many PRs, you may generate load on shared infrastucture such as CI. It is *highly* recommended that you:
> * slow the rate of PR creation by making Turbolift sleep in between PRs
> * create PRs in batches, for example by commenting out repositories in `repos.txt`
> * Use the `--draft` flag to create the PRs as Draft

#### Working with multiple PR description files

Occasionally you may want to work with more than one PR title and description. When this is the case, use the flag `--description` to specify an alternative file when creating prs.
The first line of the file chosen will be used as the PR title and the rest as the description body.

```console
turbolift create-prs --repos repoFile1.txt --description prDescriptionFile1.md
turbolift create-prs --repos repoFile2.txt --description prDescriptionFile2.md
```

### After creating PRs

#### Viewing status

While it's simple to search for PRs in GitHub search, `turbolift pr-status` can be used to view PR status in the terminal. For example:

Viewing a summary of PRs:
```
$ turbolift pr-status
...
State        Count
Merged       139
Open         53
Closed       29
Skipped      0
No PR Found  1
```

Viewing a detailed list of status per repo:
```
$ turbolift pr-status --list
...
Repository                                                State   Reviews           Build status    URL
redacted/redacted                                         OPEN    REVIEW_REQUIRED   SUCCESS         https://github.redacted/redacted/redacted/pull/262
redacted/redacted                                         OPEN    REVIEW_REQUIRED   SUCCESS         https://github.redacted/redacted/redacted/pull/515
redacted/redacted                                         OPEN    REVIEW_REQUIRED   SUCCESS         https://github.redacted/redacted/redacted/pull/342
redacted/redacted                                         MERGED  APPROVED          SUCCESS         https://github.redacted/redacted/redacted/pull/407
redacted/redacted                                         MERGED  REVIEW_REQUIRED   SUCCESS         https://github.redacted/redacted/redacted/pull/220
redacted/redacted                                         OPEN    REVIEW_REQUIRED   FAILURE         https://github.redacted/redacted/redacted/pull/105
redacted/redacted                                         MERGED  APPROVED          SUCCESS         https://github.redacted/redacted/redacted/pull/532
redacted/redacted                                         MERGED  APPROVED          SUCCESS         https://github.redacted/redacted/redacted/pull/268
redacted/redacted                                         OPEN    REVIEW_REQUIRED   FAILURE         https://github.redacted/redacted/redacted/pull/438
...
```

#### Updating PRs

Use the `update-prs` command to update PRs after creating them. Current options for updating PRs are:

- `--push` to push new commits
- `--amend-description` to update PR titles and descriptions
- `--close` to close PRs

If the flag `--yes` is not passed with an `update-prs` command, a confirmation prompt will be presented.
As always, use the `--repos` flag to specify an alternative repo file to the default `repos.txt`.

##### Examples

```turbolift update-prs --close [--yes]```
```turbolift update-prs --push [--yes]```
```turbolift update-prs --amend-description [--description prDescriptionFile1.md] [--yes]```

Note that when updating PR descriptions, as when creating PRs, the `--description` flag can be used to specify an 
alternative description file to the default `README.md`.
The updated title is taken from the first line of the file, and the updated description is the remainder of the file contents.

## Status: Preview

This tool is fully functional, but we have improvements that we'd like to make, and would appreciate feedback.

## Contributing

See [CONTRIBUTING.md](./CONTRIBUTING.md)

## Local development

To build locally:
```
make build
```

To run tests locally:
```
make test
```
