#!/bin/bash

set -e

# setup first time user. if not provided it will set as defaults.
check() {
    /app/authcorectl setup -e "$ADMIN_EMAIL" -t "$ADMIN_PHONE" -p "$ADMIN_PASSWORD" --ignore
}

# start the authcored main process
start() {
    /app/authcored
}

# skip checking in kubernetes and other cases
if [[ "$1" == "--skipcheck" ]]; then
    start
# non kubernetes, require test
else
    check && start
fi