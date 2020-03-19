![publish to dockerhub](https://github.com/untillpro/cder/workflows/Test%20build.sh/badge.svg)
# cder
Main idea is to update container instances directly from git repos or artifacts



# Build
`untillpro/cder:v{ver}` and `untillpro/cdernode:v{ver}` docker containers will be created and pushed to Dockerhub on pushing tag to github (`build.sh` will be executed by github action). The tag should be `v*`.
## Containers used
- golang:1.14.0
- node:10.16.1

# Usage
- `cd` command
  - Each `--timeout` seconds
    - `--repo` pulled to `<--working-dir>/repos/lastURI(<--repo>)` folder. The last commit differs from the stored one -> `deployAll` is executed
      - `--extraRepo` if processed
        - `--extraRepo <url>` form is used
          - <url> is checked out to `--workDir/` ???
          - `go.mod`: `replace <url> => ../<lastURI(url)>` appended
        - `--extraRepo <urlFrom>=<urlTo>` form is used
          - <urlTo> is checked out to `--workDir/` ???
          - `go.mod`: `replace <urlFrom> => ../<lastURI(urlTo)>` appended
      - `go build -o <--output>`
      - stop currently executing process (if is)
        - SIGINT is sent (does nothing on windows (not supported))
        - not finished in 30 seconds -> kill
      - built exectable is moved to `<--working-dir>` and launched
        - `--args` are provided in command line
      - After deploy all repos (even those which wasn't changed) are reseted using `git reset --hard`
        - `go.mod` is reverted to original state
  - each `--extraRepo` url is pulled to `<--working-dir>/repos/lastURI(<--extraRepo>)`. The last commit differs from the stored one -> `deploy` is executed
    - nothing is made for golang repos
  - `deploy.sh` used instead golang delpyer if exists at `--working-dir` 
- `cdurl` command
  - content from `--url` is downloaded each `--timeout` seconds
    - should consist of 2 lines separated by `\n`
  - 1st line changed i.e. new artifact version is released
    - `<--working-dir>/artifacts/<--url>/work-dir` dir is recreated
    - content from 1st line url is downloaded and saved as `<--working-dir>/artifacts/<--url>/<lastURI.ext>`
    - unzipped to `<--working-dir>/artifacts/<--url>/work-dir`
    - assume 2nd line is changed
  - 2nd line changed
    - content from 2nd line url is downloaded and saved as `<--working-dir>/artifacts/<--url>/work-dir/deploy.sh` and executed  
- `-v` means verbose mode
- `--option1 arg1 arg2` are passed to `out.exe`

# Custom deployer (deploy.sh)
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
    
# Seeding Single Repo

```sh
./cder cd --repo https://github.com/untillpro/directcd-test \
  -o directcd-test.exe \
  -t 10 \
  -w .tmp
```

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

# Seeding URL
```sh
./cder cdurl \
  --url https://github.com/untillpro/url 
  -v \
  -t 10 \
  -w .tmp \
```

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
