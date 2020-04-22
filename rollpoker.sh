#!/bin/bash

export PATH="$PATH:/usr/local/go/bin"
export GOROOT="/usr/local/go"

export funcs="MakeTable Poker"

set -ex

case "$1" in
  gohost)
    (cd gofuncs ; go build -o gfhost host/main.go) && ./gofuncs/gfhost
    ;;
  gotest)
    (cd gofuncs ; go build -o gftest test/*.go) && ./gofuncs/gftest
    ;;
  webhost)
    (firebase serve --only hosting)
    ;;
  webdeploy)
    # Deploy public/ to firebase
    firebase deploy
    # And deploy gofuncs to gcloud functions
    ;;
  godeploy)
    for func in $funcs ; do
      (cd gofuncs ; gcloud functions deploy $func --runtime go113 --trigger-http --allow-unauthenticated)
    done
    ;;
  deploy)
    $0 webdeploy && $0 godeploy
    ;;
  *)
    echo "Usage: ./$0 deploy"
    ;;
esac
