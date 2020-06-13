# authcore-cypress
## End-to-end testing using Cypress for kitty project

This project provides an entry point for running end-to-end test for kitty project using Cypress.

Note:

The image will have the following environment variable for running the test:
```
CYPRESS_AUTHCORE_WEB_HOST             // Define the authcore-web host
```

### Local environment:

Check the configuration file `docker-compose.yml` in [authcore](https://gitlab.com/blocksq/authcore) repository

As the end-to-end test can be run independently, the docker container for authcore-cypress does not run any services by default. It only provides an entry point to perform test if necessary.

Before running Cypress, installation of the library is required. In docker environment run the following to perform installation:
```
$ docker-compose exec authcore-cypress yarn cypress install
```

There are two ways to use E2E test. First is to run the test in local environment, second is to run
the test in Docker environment

In local environment, it is most useful to run `yarn cypress open` to open the Cypress native
application. In this way there is GUI to run the test case, the test is shown in a browser so the
result can be checked immediately and it is possible to see console log if necessary.

For Docker environment, it is similar to the case in CI, which the tests are run in headless
browser and there is video recording.

To run the test, run the following:
```
$ docker-compose exec authcore-cypress yarn cypress run
```

For any modification of the test access `./cypress/integration` to create or change any files.

### Gitlab CI environment:

Check the configuration file `.gitlab-ci.yml` in [kitty](https://gitlab.com/blocksq/kitty) repository

The end-to-end testing job will only be perform on master branch.
