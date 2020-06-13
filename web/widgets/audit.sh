yarn audit
if [ $? -gt 0 ]
then
  exit 1
fi
exit 0