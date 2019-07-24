#/bin/bash
set -e
mkdir .air-shell 2> /dev/null || :
cp deployer.sh .air-shell

go build
./cder cd \
  --repo https://github.com/untillpro/untill-air-shell \
  -w .air-shell \
  --deployer-env deployerenv1=1\
  --deployer-env deployerenv2=2\
  -t 10
