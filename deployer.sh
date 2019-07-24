#/bin/bash
set -e
echo "I'm custom deployer"
echo args="$@"
echo Deployer environment:
env | grep -i deployer || :

case $1 in
   "deploy") 
		echo "deployer.deploy"
		cd $2
		npm ci
		npm run build
		;;
   "deploy-all") 
		echo "deployer.deploy-all" 
		;;
   "start") 
		echo "deployer.start" 
		;;
   "stop") 
		echo "deployer.stop" 
		;;
   *) echo "Sorry, not sure what you mean"
		exit 1
		;;
esac