#!/bin/sh
apk update
apk add curl
echo "Both of these should return unknown"
curl http://app1/domains/example.com
echo
curl http://app1/domains/exmaple2.com
echo
for i in `seq 1 499`
do
   curl -X PUT http://app1/events/example.com/delivered &
   curl -X PUT http://app2/events/example.com/delivered &
   curl -X PUT http://app2/events/example2.com/delivered &
   curl -X PUT http://app1/events/example2.com/delivered &
done
wait $(jobs -p)
echo "Both of these should still return unknown because we wrote 998 deliveries to example.com and example2.com"
curl http://app1/domains/example.com
echo
curl http://app1/domains/exmaple2.com
echo
curl -X PUT http://app1/events/example.com/delivered &
curl -X PUT http://app2/events/example.com/delivered &
wait $(jobs -p)
echo "Now the first should be catch-all and the second still unknown"
curl http://app1/domains/example.com
echo
curl http://app1/domains/exmaple2.com
echo
curl -X PUT http://app1/events/example.com/bounced
echo "Now the first should be not catch-all and the second still unknown"
curl http://app1/domains/example.com
echo
curl http://app1/domains/exmaple2.com
echo