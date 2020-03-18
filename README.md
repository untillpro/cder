![build.sh](https://github.com/untillpro/cder/workflows/Test%20build.sh/badge.svg)
# cder

Main idea is to update container instances directly from git repos or artifacts

# Containers versions used
- golang:1.14.0
- node:10.16.1

# Build

## Create Docker Containers

`untillpro/cder:v{ver}` and `untillpro/cdernode:v{ver}` docker containers will be created and pushed to Dockerhub on pushing tag to github (`build.sh` will be executed by github action)

# Seeding Single Repo

```sh
./cder cd --repo https://github.com/untillpro/directcd-test \
  -o directcd-test.exe \
  -t 10 \
  -w .tmp
```

- Repo directcd-test is pulled every 10 seconds
- If changes occur
    - Prevous instance (if any) of `directcd-test.exe` is terminated
    - `go build -o directcd-test.exe` is invoked
    - `directcd-test.exe` is launched
- Working files are located in `.tmp` folder

# Seeding Few Repos

```sh
./cder cd \
  --repo https://github.com/untillpro/directcd-test \
  --extraRepo https://github.com/untillpro/directcd-test-print=https://github.com/maxim-ge/directcd-test-print \
  -v \
  -o out.exe \
  -t 10 \
  -w .tmp \
  -- --option1 arg1 arg2
```

- Repos specified by `--repo` (main repo) and `--extraRepo` flags are pulled every 10 seconds
- If changes occur 
  - Prevous instance (if any) of `out.exe` is terminated
  - `replace` directive is added to main repo `go.mod`
    - `replace github.com/untillpro/directcd-test-print => ../directcd-test-print`
  - main repo is built and launched
  - `go.mod` is reverted to original state
- `-v` means verbose mode
- `--option1 arg1 arg2` are passed to `out.exe`

# Custom Deployer

- If working directory contains `deploy.sh` it will be used to deploy
- deployer is executed using `env` command
- Working directory is one specified by `-w` flag
- First argument is one of the following:
  - `stop`
  - `deploy`
    - Executed for any changed repo
    - First argument is absolute path to changed repo
  - `deploy-all`
    - Executed once when any repo is changed
    - Absolute paths to ALL repositories folders are passed as arguments
- Environment variables for deployer can be supplied with `--deployer-env <name>=<value>` argument
- After deploy all repos (even those which wasn't changed) are reseted using `git reset --hard`

# Seeding URL
```sh
./cder cdurl \
  --url https://github.com/untillpro/url 
  -v \
  -t 10 \
  -w .tmp \
```
- content of specified url will be fetched each 10 seconds. It should consist of 2 lines separated by `\n`
  - 1st line - url to an artifact zip
  - 2nd line - url to a deployer shell script
- artifacti's URL changed (i.e. new version is released)
  - `<workingDir>/artifacts/<artifact-name>` folder is recreated
  - artifact zip is downloaded, saved to `<workingDir>/artifacts/<artifact-name>/<artifact-name>.zip` and unzipped to `<workingDir>/artifacts/<artifact-name>/work-dir` folder
- deployer URL is changed
  - deployer is downloaded, saved as `<workingDir>/artifacts/<artifact-name>/depoy.sh` and copied to `<workingDir>/artifacts/<artifact-name>/work-dir` folder
    - if 1st URL is not changed so far then existing zip is unzipped to `<workingDir>/artifacts/<artifact-name>/work-dir` folder (folder is recreated)
  - `deploy.sh` will be executed in `<workingDir>/artifacts/<artifact-name>/work-dir` folder
# Links

Hooks
- https://developer.github.com/v3/repos/hooks/#create-a-hook
- https://developer.github.com/v3/activity/events/types/#pushevent
- https://docs.gitea.io/en-us/webhooks

Stop process
- https://www.ctl.io/developers/blog/post/gracefully-stopping-docker-containers/
  - `docker stop ----time=30 foo`, SIGTERM
  - `docker kill ----signal=SIGQUIT nginx`, if you want to initiate a graceful shutdown of an nginx server, you should send a SIGQUIT
  - `docker kill ----signal=SIGWINCH apache`, Apache uses SIGWINCH to trigger a graceful shutdown
- https://husobee.github.io/golang/ecs/2016/05/19/ecs-graceful-go-shutdown.html

Misc 

- golang url https://play.golang.org/p/6kBtuHvUlQc
