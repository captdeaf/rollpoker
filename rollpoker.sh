#!/bin/bash

export PATH="$PATH:/usr/local/go/bin"
export GOROOT="/usr/local/go"

export funcs="MakeTable"

case "$1" in
  gohost)
    (cd gofuncs ; go build -o gfhost host/main.go) && ./gofuncs/gfhost
    ;;
  webhost)
    (./firebase serve --only hosting)
    ;;
  deploy)
    # Deploy public/ to firebase
    ./firebase deploy
    # And deploy gofuncs to gcloud functions
    (cd gofuncs ; gcloud functions deploy $funcs --runtime go111 --trigger-http --allow-unauthenticated)
    ;;
  *)
    echo "Usage: ./$0 deploy"
    ;;
esac
