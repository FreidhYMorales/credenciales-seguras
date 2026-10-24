# Credenciales seguras

Validador léxico y generador de credenciales seguras construido con expresiones regulares
y autómatas finitos. Proyecto final del curso de Autómatas y Lenguajes Formales.

Valida correos (RFC 5322, subconjunto) y contraseñas (NIST SP 800-63B + CIS), informa
**qué regla exacta** falla, calcula el nivel de seguridad y genera credenciales caminando
el autómata (no sorteando cadenas al azar).

## Probar sin compilar

1. Ir a la página de [releases](https://github.com/FreidhYMorales/credenciales-seguras/releases/latest)
   y descargar `credenciales.exe` (Windows) o `credenciales-linux` (Linux x86-64).
2. Ejecutarlo. Abre el servidor en <http://127.0.0.1:8080>; no hace falta Go, Python ni
   ninguna otra dependencia: el binario lleva el frontend embebido.
3. En **Validar**, probar `abc` (falla P1, P2, P4 y P5) y `>nslx613LK2%|#PxFbA5` (válida,
   nivel Fuerte). En **Generar**, semilla `2026` y longitud 20 reproducen la misma
   contraseña y el mismo camino sobre el autómata.

El paquete `credenciales-seguras.zip` de la misma release trae los dos binarios, el
documento final en PDF y en DOCX, y un `LEEME.txt`.

> En Windows, el `.exe` está probado bajo Wine; si SmartScreen lo bloquea, es por ser un
> ejecutable sin firma: «Más información» → «Ejecutar de todas formas».

## Requisitos

- Go 1.27 o superior (para compilar).
- Nada más en tiempo de ejecución: el binario es estático y sirve su propio frontend.

## Compilar y ejecutar

```sh
make build            # bin/credenciales
make run              # http://localhost:8080
make build-windows    # bin/credenciales.exe
```

## Verificación

```sh
make check            # gofmt + go vet + go test (la única puerta de calidad)
```

## Estructura

```
cmd/app/            servidor HTTP
internal/automata/  AFND-ε, subconjuntos, árbol, minimización, simulación, toRE2
internal/rules/     las 15 reglas y sus autómatas
internal/validate/  reporte por regla, normalización, nivel de seguridad
internal/generate/  camino aleatorio guiado sobre el AFD producto
internal/api/       handlers y contratos JSON
web/                frontend (HTML/CSS/JS, sin build step)
testdata/           pares.json con los 20 casos de prueba
docs/               arquitectura, desarrollo, ADRs, contrato de API, plan
docs/entregas/      documentos ya entregados (avance 1)
docs/referencias/   enunciado y estándares de consulta
```

## Documentación

| Archivo | Contenido |
|---|---|
| `docs/entregas/final/Credenciales_seguras.pdf` | documento final, 20 pp. |
| `docs/architecture.md` | stack, límites entre módulos, puerta de calidad |
| `docs/development.md` | comandos y convenciones |
| `docs/api.md` | contrato de los endpoints |
| `docs/adr/` | decisiones de arquitectura |
| `docs/plans/app-credenciales/plan.md` | alcance, cronograma, definición de terminado |
| `docs/entregas/` | documentos entregados, con la versión vigente marcada |
| `docs/referencias/` | enunciado del curso y estándares citados |

## Autómatas

15 autómatas en tiempo de ejecución: `P1..P8` (contraseñas), `C1..C6` (correos) y `C*`
(correo completo). `P0` y `C0` son verificaciones de conjunto sobre el alfabeto y no
necesitan autómata. El autómata de las 24 permutaciones (`COMP`) queda documentado pero
no se construye en tiempo de ejecución.
