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
| `documento.md` | documento final en Markdown, listo para exportar | borrador avanzado |
| `pares.md` | las dos tablas de los 20 pares, generadas desde `pares.json` | generado |
| `capturas/` | las seis figuras del documento, tomadas del build actual | vigente |

`pares.md` se reescribe con `make pares-tabla` y la prueba `TestPairsMarkdownAlDia`
falla si el archivo se desincroniza de `testdata/pares.json`. Las tablas están
copiadas dentro de `documento.md`: si se regeneran, hay que volver a pegarlas.

Lo que falta para cerrar la entrega (video de 1 minuto y presentación de 10 minutos)
está en la sección *Definition of done* y en la *Run sheet* de
`docs/plans/app-credenciales/plan.md`.
