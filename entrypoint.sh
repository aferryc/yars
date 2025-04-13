#!/bin/sh
set -e

case "$SERVICE_TYPE" in
	"compiler")
		exec /app/bin/compiler
		;;
	"reconciliation")
		exec /app/bin/reconciliation
		;;
	"server" | *)
		exec /app/bin/server
		;;
esac
