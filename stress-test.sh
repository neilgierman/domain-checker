#!/bin/bash
set -x
apk update
apk add curl jq
wget https://raw.githubusercontent.com/vishnubob/wait-for-it/master/wait-for-it.sh
chmod +x wait-for-it.sh
echo "1. Waiting for services to be up"
./wait-for-it.sh rabbitmq:5672 --timeout=60
./wait-for-it.sh mongodb:27017 --timeout=60
./wait-for-it.sh app1:80 --timeout=60
./wait-for-it.sh app2:80 --timeout=60
echo "2. Both of these should return unknown"
if [ "`curl -s http://app1/domains/example.com | jq -r .Status`" != "unknown" ]
then
   exit 1
fi
if [ "`curl -s http://app1/domains/example2.com | jq -r .Status`" != "unknown" ]
then
   exit 1
fi
for i in `seq 1 499`
do
   curl -s -X PUT http://app1/events/example.com/delivered &
   curl -s -X PUT http://app2/events/example.com/delivered &
   curl -s -X PUT http://app2/events/example2.com/delivered &
   curl -s -X PUT http://app1/events/example2.com/delivered &
done
wait $(jobs -p)
sleep 10
echo "3. Both of these should still return unknown because we wrote 998 deliveries to example.com and example2.com"
if [ "`curl -s http://app1/domains/example.com | jq -r .Status`" != "unknown" ]
then
   exit 1
fi
if [ "`curl -s http://app1/domains/example2.com | jq -r .Status`" != "unknown" ]
then
   exit 1
fi
curl -s -X PUT http://app1/events/example.com/delivered &
curl -s -X PUT http://app2/events/example.com/delivered &
wait $(jobs -p)
sleep 4
echo "4. Now the first should be catch-all and the second still unknown"
if [ "`curl -s http://app1/domains/example.com | jq -r .Status`" != "catch-all" ]
then
   exit 1
fi
if [ "`curl -s http://app1/domains/example2.com | jq -r .Status`" != "unknown" ]
then
   exit 1
fi
curl -s -X PUT http://app1/events/example.com/bounced &
wait $(jobs -p)
sleep 4
echo "5. Now the first should be not catch-all and the second still unknown"
if [ "`curl -s http://app1/domains/example.com | jq -r .Status`" != "not catch-all" ]
then
   exit 1
fi
if [ "`curl -s http://app1/domains/example2.com | jq -r .Status`" != "unknown" ]
then
   exit 1
fi
