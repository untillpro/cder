#!/bin/bash
set -e
echo "I'm custom deployer"
echo args="$@"
echo Deployer environment:
env | grep -i deployer || :

case $1 in
   "deploy")
                echo "deployer.deploy"
                cd $2
                if [[ -f ./package.json ]]; then
                  needCI=$(git diff --stat master@{1} master package-lock.json || echo "err")
                  if [[ $needCI == "" ]]; then
                    echo "only build"
                    npm run build
                  else
                    echo "ci and build"
                    npm ci
                    npm run build
                  fi
                  kill $(pidof nginx) || true
                  cp -r ./build /usr/share/nginx/html
                else
                  mkdir -p /usr/share/nginx/html/build
                  kill $(pidof nginx) || true
                  cp -r ./* /usr/share/nginx/html/build
                fi
                service nginx start
                ;;
   "deploy-all")
                echo "deployer.deploy-all"
                ;;
   "start")
                echo "deployer.start"
                cd $2
                ;;
   "stop")
                echo "deployer.stop"
                cd $2
                service nginx stop || true
                service nginx stop
                rm -rf ./build /usr/share/nginx/html/build
                ;;
   *) echo "Sorry, not sure what you mean"
                exit 1
                ;;
esac
