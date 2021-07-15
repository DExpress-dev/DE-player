MODULE_NAME="file_express"
MODULE_PID=$MODULE_NAME".pid"

if [[ ! -e $MODULE_PID ]]
then
	echo $MODULE_PID" not found"
else
	cat $MODULE_PID | xargs kill
	rm -f $MODULE_PID
fi