# Authcore - Secure seamless access to everything.

[![pipeline status](https://gitlab.com/blocksq/authcore-mirror/badges/master/pipeline.svg)](https://gitlab.com/blocksq/authcore-mirror/-/commits/master)
[![coverage report](https://gitlab.com/blocksq/authcore-mirror/badges/master/coverage.svg)](https://gitlab.com/blocksq/authcore-mirror/-/commits/master)
[![License](https://img.shields.io/badge/license-Apache%202-blue)](https://github.com/authcore/authcore/blob/master/LICENSE)

![Logo](assets/logo.svg)

----

## Authentication is complex. Let us deal with it.

* **Developer-first**<br>
  Rapid integrate with any apps with standard-based protocols and elegant REST API.
* **Beautiful UI**<br>
  Widgets Add login and user management functions to your app with little coding.
* **Impeccable Security**<br>
  Protect all user accounts with advanced security features. Third-party audited source code.

----

## Start using Authcore

Authcore can be installed in most GNU/Linux distributions and in a number of cloud providers.

[Documentation](https://docs.authcore.io/guides/install)

## Start developing Authcore

The project uses Docker and docker-compose as a development environment. You need to install the
latest version of Docker and docker-compose on your platform.

Install `web` and `widgets` dependencies:

```
$ docker-compose run --rm web yarn install
$ docker-compose run --rm widgets yarn install
```

First-time setup:

```
docker-compose run --rm server go run authcore.io/authcore/cmd/authcorectl setup -e <YOUR EMAIL>
```

Then, you can start the full development stack with this command:

```
$ docker-compose up
```

Open Admin Portal in browser: https://authcore.dev:8001. You will need to
[bypass](https://stackoverflow.com/questions/58802767/no-proceed-anyway-option-on-neterr-cert-invalid-in-chrome-on-macos)
browser's untrusted certificate error.

Open current API docs: https://authcore.dev:8001/docs/