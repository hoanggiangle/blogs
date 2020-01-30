#!/bin/sh

set -e

CMD=$1

case $CMD in
	bash | sh | go)
		$@
		;;
	build)
		go $@
		;;
	*)
		go build -o /tmp/app
		exec /tmp/app $@
		;;
esac
