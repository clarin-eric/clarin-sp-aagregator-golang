#!/bin/bash

set -e

LATEST_VERSION="1.0.16"

#Use version supplied in the 1st argument or a default value instead
VERSION=${1:-$LATEST_VERSION}

echo "Initializing directory with build-script v${VERSION}"

#TODO: update
#   take into account custom copy_data.sh. When this file is a symlink don't replace it

git init
git submodule add https://gitlab.com/CLARIN-ERIC/build-script.git build-script
cd build-script
git checkout ${VERSION}
cd ..
ln -s build-script/build.sh build.sh
ln -s build-script/copy_data_noop.sh copy_data.sh
ln -s build-script/update_version_noop.sh update_version.sh
cp build-script/_gitlab-ci_default.yml .gitlab-ci.yml

mkdir image run
touch image/.gitkeep
touch run/.gitkeep
git add .
git commit -m "Intialized empty repo with build script v${VERSION}"

echo "Done"
