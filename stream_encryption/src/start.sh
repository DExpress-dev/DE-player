#! /bin/bash

MODULE_NAME="file_express"
MODULE_DIR="$( cd "$( dirname "$0"  )" && pwd  )"
MODULE_PATH=$MODULE_DIR"/"$MODULE_NAME

chmod +x stop.sh
chmod +x $MODULE_NAME

ps -ef | grep $MODULE_PATH | grep -v grep
if [ $? -ne 0 ]
then
echo "start $MODULE_PATH....."
setsid $MODULE_PATH &
else
echo "$MODULE_PATH runing....."
fi