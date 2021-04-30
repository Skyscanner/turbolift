# turbolift

A simple tool to help apply changes across many GitHub repositories simultaneously.

## Philosophy

Anyone who has had to manually make changes to many GitHub repositories knows that it's hard to beat the simplicity of just cloning the repositories and updating them locally. You can use any tools necessary to make the change, and there's a degree of immediacy in having local files to inspect, tweak or run validation.

It's dumb but it works. It doesn't scale well, though. Manually cloning and raising PRs against tens/hundreds of repositories is painful and boring.

Turbolift essentially automates the boring parts and stays out of the way when it comes to actually making the changes. It automates forking, cloning, committing, and raising PRs en-masse, so that you can focus on the substance of the change.

> Historical note: Turbolift supersedes an internal system at Skyscanner named Codelift. Codelift was a centralised batch system, requiring changes to be scripted upfront and run overnight. While Codelift was useful, we have found that a decentralised, interactive tool is far easier and quicker for people to use in practice. 

## Demo

This demo shows Turbolift in action, creating a simple PR in two repositories:

![Screencast demo of turbolift in use](docs/demo.gif "Screencast demo of turbolift in use")

## Installation

### Downloading binaries

Pre-built binary archives can be downloaded from the [Releases](https://github.com/Skyscanner/turbolift/releases) page.

* Download, extract the archive, and move it onto your `PATH`. 
* Note that the binaries are not currently notarized for MacOS Gatekeeper. If errors are displayed, use `xattr -c PATH_TO_TURBOLIFT_BINARY` to un-quarantine the binary, or right-click on the binary in Finder and choose 'Open' once to allow future execution. Distribution will be improved under https://github.com/Skyscanner/turbolift/issues/43.

You must also have the GitHub CLI, `gh`, installed:

* Install using `brew install gh`
* Before using Turbolift, run `gh auth login` once and follow the prompts, to authenticate against github.com and/or your GitHub Enterprise server.

## Basic usage:

Making changes with turbolift is split into six main phases:

1. `init` - getting set up
2. Identifying the repos to operate upon
3. Running a mass `clone` of the repos (which automatically creates a fork in your user space)
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
* Raising draft PRs can be a good way to collect feedback, especially CI test results, with less pressure on reviewers (TODO: relies on https://github.com/Skyscanner/turbolift/issues/11 being implemented)
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

### Running a mass `clone`

```turbolift clone```

This clones all repositories listed in the `repos.txt` file into the `work` directory.

### Making changes

Now, make changes to the checked-out repos under the `work` directory. 
You can do this manually using an editor, using `sed` and similar commands, or using [`codemod`](https://github.com/facebook/codemod)/[`comby`](https://comby.dev/), etc. 

**You are free to use any tools that help get the job done.**

If you wish to, you can run the same command against every repo using `turbolift foreach ...` (where `...` is the shell command you want to run).

For example, you might choose to:

* `turbolift foreach rm somefile` - to delete a particular file
* `turbolift foreach sed -i '' 's/foo/bar/g' somefile` - to find/replace in a common file
* `turbolift foreach make test` - for example, to run tests (using any appropriate command to invoke the tests)
* `turbolift foreach git add somefile` - to stage a file that you have created

At any time, if you need to update your working copy branches from the upstream, you can run `turbolift foreach git pull upstream master`.

It is highly recommended that you run tests against affected repos, if it will help validate the changes you have made.

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

If you need to mass-close PRs, it is easy to do using `turbolift foreach` and the `gh` GitHub CLI ([docs](https://cli.github.com/manual/gh_pr_close)):

For example:

```
turbolift foreach gh pr close --delete-branch YOUR_USERNAME:CAMPAIGN_NAME
```

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
