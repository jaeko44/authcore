#!/bin/sh

set -e
set -o xtrace

BASEDIR=$( cd "$(dirname "$0")"/.. ; pwd -P )
if [ -z "$BASEDIR" ]
then
    echo "cannot determine base path"
    exit 255
fi

export COMMIT_REF_NAME=${CI_COMMIT_REF_NAME:-$(git symbolic-ref HEAD | sed -e 's,.*/\(.*\),\1,')}
export BUILDDIR="$BASEDIR"/build
export DISTDIR="$BASEDIR"/dist
export DISTFILE="$DISTDIR"/authcore-dist-${COMMIT_REF_NAME}.tar.gz

rm -fr "$DISTDIR"
mkdir -p "$BUILDDIR"/db "$BUILDDIR"/web "$DISTDIR"
cp -R "$BASEDIR"/db/migrations "$BUILDDIR"/db/migrations
cp -R "$BASEDIR"/api "$BUILDDIR"/api
cp -R "$BASEDIR"/templates "$BUILDDIR"/templates
cp -R "$BASEDIR"/policies "$BUILDDIR"/policies
cp -R "$BASEDIR"/scripts/entrypoint.sh "$BUILDDIR"/entrypoint.sh
cp -R "$BASEDIR"/web/dist "$BUILDDIR"/web/dist
tar zcvf "$DISTFILE" --exclude .DS_Store -C build .