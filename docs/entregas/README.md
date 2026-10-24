# Entregas

Documentos que ya se entregaron al curso. No son material de consulta: son la
evidencia de lo que se presentó y la base sobre la que se escribe el documento final.

## `avance-1/`

Primer avance, calificado con los 5 puntos de zona. Cubre los pasos 1 a 4 del método
sobre las reglas P0-P8 y C0-C6: alfabetos, expresiones regulares, gramáticas regulares
y derivaciones, construcción de Thompson con tablas de estados, método del árbol, AFD
mínimo, anexo de pruebas y el criterio de nivel de seguridad (Anexo C).

| Archivo | Qué es | Estado |
|---|---|---|
| `AYLF Final Avance corregido.pdf` | versión entregada, 35 pp. | **vigente** |
| `Avance Proyecto Final.pdf` | versión previa, 26 pp. | superado |
| `Avance_1_completo.docx` | fuente editable del PDF vigente | vigente |
| `Avance_1_completo.odt` | exportación a LibreOffice | derivado |
| `Avance_1.docx` | borrador inicial, sin el capítulo 3 | superado |

Los scripts de `codigo_avance/` que parchean el `.docx` apuntan a
`docs/entregas/avance-1/Avance_1_completo.docx` y se ejecutan desde la raíz del repo.

## `final/`

Entrega final: documento, capturas y los 20 pares en forma legible.

| Archivo | Qué es | Estado |
|---|---|---|
| `Credenciales_seguras.pdf` | documento final, 20 pp., el archivo para entregar | **vigente** |
| `Credenciales_seguras.docx` | fuente editable del PDF, generada desde el Markdown | vigente |
| `documento.md` | fuente en Markdown del documento final | vigente |
| `pares.md` | las dos tablas de los 20 pares, generadas desde `pares.json` | generado |
| `capturas/` | las seis figuras del documento, tomadas del build actual | vigente |

El paquete publicado está en
<https://github.com/FreidhYMorales/credenciales-seguras/releases/latest>: los dos binarios,
el zip, el PDF, el DOCX y `SHA256SUMS.txt`. `make entrega` arma `dist/credenciales-seguras.zip` con los dos binarios, el PDF, el
`.docx` y un `LEEME.txt`; es el paquete que se sube. El `.docx` y el `.pdf` se generan con
`make documento` (necesita `python-docx` y
LibreOffice; el resto del proyecto no). Ambos están versionados.

`pares.md` se reescribe con `make pares-tabla` y la prueba `TestPairsMarkdownAlDia`
falla si el archivo se desincroniza de `testdata/pares.json`. Las tablas están
copiadas dentro de `documento.md`: si se regeneran, hay que volver a pegarlas.

Lo que falta para cerrar la entrega (video de 1 minuto y presentación de 10 minutos)
está en la sección *Definition of done* y en la *Run sheet* de
`docs/plans/app-credenciales/plan.md`.
