yarn audit
if [ $? -gt 8 ]
then
  exit 1
fi
exit 0