#!/bin/bash

case "$1" in
  deploy)
    # Deploy public/ to firebase
    ./firebase deploy
    # And deploy gofuncs to gcloud functions
    echo "TODO: Deploy gofuncs"
    ;;
  *)
    echo "Usage: ./$0 deploy"
    ;;
esac
