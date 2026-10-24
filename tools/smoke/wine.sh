#!/bin/sh
# Start a Windows binary under Wine, wait for /api/salud, validate one
# credential and stop it. Needs wine and curl; never part of `make check`.
#
# Usage: wine.sh <exe> <port> [log]
set -eu

exe=${1:?falta el ejecutable}
puerto=${2:?falta el puerto}
log=${3:-/tmp/credenciales-wine.log}

command -v wine >/dev/null 2>&1 || { echo "falta wine"; exit 1; }
command -v curl >/dev/null 2>&1 || { echo "falta curl"; exit 1; }

WINEDEBUG=-all wine "$exe" -addr "127.0.0.1:$puerto" >"$log" 2>&1 &
pid=$!
trap 'kill $pid 2>/dev/null || true' EXIT

i=0
while [ "$i" -lt 60 ]; do
	if curl -sf "http://127.0.0.1:$puerto/api/salud" >/dev/null 2>&1; then
		break
	fi
	i=$((i + 1))
	sleep 0.5
done
# Give Wine a moment when the port opened late.
curl -sf "http://127.0.0.1:$puerto/api/salud" | grep -q '"estado":"ok"' ||
	{ echo "$exe no responde en /api/salud (ver $log)"; exit 1; }

# "abc" fails P1, P2, P4 and P5: enough to prove the engine runs inside Wine.
curl -sf -X POST -H 'Content-Type: application/json' \
	-d '{"tipo":"password","valor":"abc"}' \
	"http://127.0.0.1:$puerto/api/validar" | grep -q '"valido":false' ||
	{ echo "$exe no valida (ver $log)"; exit 1; }

echo "ok: $exe responde y valida bajo Wine (puerto $puerto)"
