#!/bin/sh

farg="$1"
shift;

dockerCompose() {
  docker compose -p go-shortener "$@"
}

printHelp() {
  echo "Error: Command not recognized."
  echo "Please read the source code of the bash script for all available commands."
}

case $farg in
"build")
  go build -o bin/app ./cmd/app
  ;;

"run")
	. ./.env && go run ./cmd/app "$@"
  ;;

"test")
	go test -v ./...
  ;;

"fmt")
	go fmt ./...
  ;;

"dc")
  dockerCompose "$@"
  ;;

"dc:up")
  dockerCompose up -d
  ;;

"dc:down")
  dockerCompose down -v
  ;;

*)
  printHelp
  ;;
esac
