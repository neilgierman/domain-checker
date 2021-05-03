# Domain-Checker
[![Build Status](https://github.com/neilgierman/domain-checker/workflows/build/badge.svg)](https://github.com/neilgierman/domain-checker/actions)

Docker:
* Make sure you have docker and docker-compose installed
* Run `docker-compose up`
* This will bring up the system with a single database and queue and 2 app front ends (listening on 81 and 82)

Stress Tests:
* Make sure you have docker and docker-compose installed
* Run `docker-compose -f docker-compose-test.yml up --exit-code-from test`
* This will bring up the entire system and run scripts from the test container
* The system will exit when the test container is done with a return code from the test container