"""Anexa la seccion 3 (punto 4 del enunciado) a una copia de Avance_1.docx."""
import copy
import os
import sys

import docx
from docx.enum.text import WD_BREAK, WD_ALIGN_PARAGRAPH
from docx.shared import Inches, Pt

from reglas import build, PALABRAS
from validador import validar_password, validar_correo, es_valido

BASE = os.path.dirname(os.path.abspath(__file__))
SRC = os.path.join(BASE, "..", "docs", "entregas", "avance-1", "Avance_1.docx")
OUT = os.path.join(BASE, "..", "docs/entregas/avance-1/Avance_1_completo.docx")
FIG = os.path.join(BASE, "figuras")

doc = docx.Document(SRC)
ref_body = doc.paragraphs[3]
ref_bullet = next(p for p in doc.paragraphs if p.style.name == "List Paragraph")
ref_tbl = doc.tables[0]
tbl_pr = copy.deepcopy(ref_tbl._tbl.tblPr)
hdr_pr = copy.deepcopy(ref_tbl.rows[0].cells[0]._tc.tcPr)


# ----------------------------------------------------------- helpers
def para(text="", bold=False, italic=False, style=None, align="just"):
    p = doc.add_paragraph(style=style)
    if text:
        r = p.add_run(text)
        r.bold, r.italic = bold, italic
        r.font.size = Pt(11)
    pf = p.paragraph_format
    pf.space_after = ref_body.paragraph_format.space_after
    pf.line_spacing = ref_body.paragraph_format.line_spacing
    if align == "just":
        p.alignment = ref_body.alignment
    elif align == "center":
        p.alignment = WD_ALIGN_PARAGRAPH.CENTER
    return p


def rich(parts, align="just"):
    """parts: lista de (texto, 'b'|'i'|'') para mezclar negritas."""
    p = para(align=align)
    for t, f in parts:
        r = p.add_run(t)
        r.bold, r.italic = "b" in f, "i" in f
        r.font.size = Pt(11)
    return p


def h2(t):
    p = doc.add_paragraph(style="Heading 2")
    r = p.add_run(t)
    r.bold, r.font.size = True, Pt(11)
    return p


def h3(t):
    p = doc.add_paragraph(style="Heading 3")
    r = p.add_run(t)
    r.bold, r.font.size = True, Pt(11)
    return p


def bullet(text, bold_lead=None):
    p = doc.add_paragraph(style="List Paragraph")
    old = p._p.pPr
    if old is not None:
        p._p.remove(old)
    p._p.insert(0, copy.deepcopy(ref_bullet._p.pPr))
    if bold_lead:
        r = p.add_run(bold_lead)
        r.bold, r.font.size = True, Pt(11)
    r = p.add_run(text)
    r.font.size = Pt(11)
    return p


def code(text):
    for line in text.split("\n"):
        p = doc.add_paragraph()
        p.paragraph_format.left_indent = Inches(0.25)
        p.paragraph_format.space_after = Pt(0)
        r = p.add_run(line if line else " ")
        r.font.name = "Courier New"
        r.font.size = Pt(9)
    doc.add_paragraph().paragraph_format.space_after = Pt(2)


def table(header, rows, widths, size=9.5, mono_cols=()):
    t = doc.add_table(rows=1 + len(rows), cols=len(header))
    t._tbl.remove(t._tbl.tblPr)
    t._tbl.insert(0, copy.deepcopy(tbl_pr))
    for j, h in enumerate(header):
        c = t.rows[0].cells[j]
        tcpr = c._tc.get_or_add_tcPr()
        for ch in hdr_pr:
            if ch.tag.endswith("}shd"):
                tcpr.append(copy.deepcopy(ch))
        c.paragraphs[0].text = ""
        r = c.paragraphs[0].add_run(h)
        r.bold, r.font.size = True, Pt(size)
    for i, row in enumerate(rows, 1):
        for j, v in enumerate(row):
            c = t.rows[i].cells[j]
            c.paragraphs[0].text = ""
            r = c.paragraphs[0].add_run(str(v))
            r.font.size = Pt(size)
            if j in mono_cols:
                r.font.name = "Courier New"
    for row in t.rows:
        for j, w in enumerate(widths):
            row.cells[j].width = Inches(w)
        for c in row.cells:
            for p in c.paragraphs:
                p.paragraph_format.space_after = Pt(1)
    doc.add_paragraph().paragraph_format.space_after = Pt(2)
    return t


def figure(name, width, caption):
    p = doc.add_paragraph()
    p.alignment = WD_ALIGN_PARAGRAPH.CENTER
    p.add_run().add_picture(os.path.join(FIG, name + ".png"), width=Inches(width))
    c = para(caption, italic=True, align="center")
    c.runs[0].font.size = Pt(10)


def fig_grid(names, width, caption):
    t = doc.add_table(rows=1, cols=len(names))
    for j, n in enumerate(names):
        c = t.rows[0].cells[j]
        c.paragraphs[0].alignment = WD_ALIGN_PARAGRAPH.CENTER
        c.paragraphs[0].add_run().add_picture(os.path.join(FIG, n + ".png"), width=Inches(width))
    c = para(caption, italic=True, align="center")
    c.runs[0].font.size = Pt(10)


# ------------------------------------------------- datos verificados
print("construyendo AFN/AFD de todas las reglas ...", flush=True)
R = build()
ST = {k: v.stats() for k, v in R.items()}
assert all(s["same"] for s in ST.values()), "subconjuntos y arbol deben coincidir"

PW = [
    "Pl4t0n0#Guatemala2026", "Mi perro come 7 Huesos!", "Xk9$mTq2#vLw8@Rz", "Vj7!nQe4%Ha2&Ub6(Kd",
    "H7$kP2#mQ9@xL4!zR8", "  Segura2026#xyz  ", "Abc123!", "contrasenasegura", "CONTRASENA2026!",
    "Password12345!!", "Aaaa1111!!!!bbbb", "Tr0ub4dor&3xYz!qW9", "abcd1234EFGH!@#$", "Zq8#Lp3$Mn5@Wx1^Yt",
    "Clave#Segura7Ñandú9",
]
EM = [
    "cristian.lopez@ejemplo.com", "usuario+etiqueta@correo.edu.gt", "a_b-c'd!e@sub.dominio.org",
    ".usuario@ejemplo.com", "usuario.@ejemplo.com", "usuario..nombre@ejemplo.com", "usuario@@ejemplo.com",
    "usuario@ejemplo", "usuario@-ejemplo.com", "usuario@ejemplo..com", "usuario ejemplo@correo.com",
    "usuario@ejemplo.c", "ana@ejemplo.c0m", "ana@ejem_plo.com", "a" * 65 + "@ejemplo.com",
    "ana@" + "b" * 64 + ".com",
]


def resumen(rep):
    fails = [f"{n.split()[0]}: {d}" if d else n.split()[0] for n, ok, d in rep if not ok]
    return "; ".join(fails)


def acorta(s, n=34):
    return s if len(s) <= n else s[: n - 3] + "..."


# ================================================================ TEXTO
doc.add_paragraph().add_run().add_break(WD_BREAK.PAGE)
p = para("3 Diseño de la expresión regular para el reconocimiento de correos y contraseñas seguras", bold=True)

para(
    "Este capítulo desarrolla el punto 4 de la tarea: el alfabeto, las reglas derivadas de los estándares "
    "investigados en los capítulos 1 y 2, las expresiones regulares, su conversión a AFND con transiciones vacías "
    "(algoritmo de Thompson) y su conversión a AFD (algoritmo de subconjuntos y método del árbol). Todas las "
    "construcciones se implementaron y verificaron con un programa propio (carpeta codigo_avance), de modo que "
    "los tamaños de autómata, las tablas y los diagramas de este documento provienen de una ejecución real y no de "
    "un cálculo manual."
)

h2("3.0 Decisiones de diseño")
bullet(
    " El enunciado indica que cada símbolo se lee de uno en uno y que los espacios no se toman en cuenta. Por eso el "
    "espacio se descarta antes de alimentar el autómata (equivale a una transición ε sobre el espacio) y no forma "
    "parte de ningún alfabeto. Esto ajusta las reglas R3 y R8 del capítulo 1, que admitían espacio; aquí se "
    "sustituyen por el alfabeto de 94 símbolos ASCII visibles.",
    "Lectura símbolo a símbolo.",
)
bullet(
    " Los lenguajes regulares son cerrados bajo intersección y complemento. En lugar de una única expresión "
    "gigante, cada regla tiene su propia ER y su propio AFD; una credencial es válida cuando todas las reglas de "
    "aceptación la reconocen y ninguna regla de rechazo la reconoce (autómata producto). Además de ser más simple, "
    "esto permite el reporte léxico exigido en el punto 7.a.i: se sabe exactamente qué regla falla.",
    "Una ER por regla.",
)
bullet(
    " Se trabaja con etiquetas por grupo de caracteres (L, l, N, S, A, H, E, P) para que las ER y los diagramas "
    "sean legibles. Cada etiqueta es un conjunto de símbolos del alfabeto; en el autómata equivale a una arista "
    "con todos los símbolos del grupo.",
    "Etiquetas.",
)

# ---------------------------------------------------------------- 3.1
h2("3.1 Definición del alfabeto (inciso a)")
h3("Alfabeto de contraseñas")
table(
    ["Etiqueta", "Grupo", "Símbolos", "Cant."],
    [
        ["L", "Mayúsculas", "A–Z", "26"],
        ["l", "Minúsculas", "a–z", "26"],
        ["N", "Números", "0–9", "10"],
        ["S", "Especiales", "! \" # $ % & ' ( ) * + , - . / : ; < = > ? @ [ \\ ] ^ _ ` { | } ~", "32"],
        ["Σp", "L ∪ l ∪ N ∪ S", "ASCII visible 0x21–0x7E (el espacio 0x20 se descarta)", "94"],
    ],
    [0.7, 1.3, 3.9, 0.6],
)
h3("Alfabeto de correos")
table(
    ["Etiqueta", "Grupo", "Símbolos", "Cant."],
    [
        ["A", "Letras", "A–Z, a–z", "52"],
        ["N", "Números", "0–9", "10"],
        ["H", "Guion", "-", "1"],
        ["E", "Otros atext", "! # $ % & ' * + / = ? ^ _ ` { | } ~", "18"],
        ["P", "Punto", ".", "1"],
        ["@", "Arroba", "@", "1"],
        ["at", "atext = A | N | H | E", "Caracteres de RFC 5322 permitidos en la parte local", "81"],
        ["Σe", "at ∪ P ∪ @", "Alfabeto de correos", "83"],
    ],
    [0.7, 1.6, 3.6, 0.6],
)
para(
    "Quedan fuera de Σe los 11 símbolos \" ( ) , : ; < > [ \\ ], porque solo aparecen en las formas quoted-string, "
    "domain-literal y comentarios de RFC 5322, que el capítulo 2 recomienda evitar y que este validador excluye "
    "deliberadamente. Un símbolo fuera del alfabeto se reporta como regla 0 (alfabeto) con su posición."
)

# ---------------------------------------------------------------- 3.2
h2("3.2 Reglas de contraseñas seguras (inciso b)")
table(
    ["ID", "Regla", "Tipo", "Fundamento"],
    [
        ["P0", "Todo símbolo pertenece a Σp", "Aceptación", "NIST: aceptar ASCII imprimible"],
        ["P1", "Longitud n con 15 ≤ n ≤ 64", "Aceptación (longitud)", "NIST Rev. 4: mínimo 15 sin MFA; permitir al menos 64"],
        ["P2", "Al menos una mayúscula (L)", "Aceptación (composición)", "CIS: al menos un carácter no alfabético; se refuerza con las 4 clases"],
        ["P3", "Al menos una minúscula (l)", "Aceptación (composición)", "Ídem"],
        ["P4", "Al menos un número (N)", "Aceptación (composición)", "CIS"],
        ["P5", "Al menos un especial (S)", "Aceptación (composición)", "CIS"],
        ["P6", "Tres o más símbolos idénticos consecutivos (aaa, 111, !!!)", "Rechazo", "Patrón trivial de diccionario"],
        ["P7", "Secuencia de 4 ascendente o descendente (abcd, dcba, 1234, 4321), sin distinguir mayúsculas", "Rechazo", "Contraseñas predecibles"],
        ["P8", "Contiene una palabra prohibida: " + ", ".join(PALABRAS) + " (sin distinguir mayúsculas)", "Rechazo", "NIST/CIS: blocklist y términos comunes"],
    ],
    [0.4, 2.9, 1.3, 1.9],
)
para(
    "La complejidad de la contraseña queda definida por P2–P5 (cuatro clases de caracteres) y la longitud por P1. "
    "NIST desaconseja imponer composición; se mantiene aquí porque un validador léxico académico no dispone de una "
    "base de contraseñas filtradas y la composición sí es expresable con expresiones regulares (ver capítulo 1). "
    "Las reglas de rechazo P6–P8 son los «casos no aceptados»: su ER describe la cadena prohibida y el AFD que la "
    "reconoce marca la contraseña como insegura. La blocklist real de millones de entradas no es una regla léxica; "
    "P8 solo implementa una lista finita de ejemplo, que sí es un lenguaje regular. El nivel de seguridad (baja, "
    "media, fuerte) se calculará en la etapa de la aplicación (punto 7.a.iv)."
)

# ---------------------------------------------------------------- 3.3
h2("3.3 Reglas de correos electrónicos (inciso c)")
table(
    ["ID", "Regla", "Fundamento"],
    [
        ["C0", "Todo símbolo pertenece a Σe", "RFC 5322: atext, punto, arroba"],
        ["C1", "Estructura local@dominio: exactamente una arroba y ambas partes no vacías", "RFC 5322 §3.4.1 (addr-spec)"],
        ["C2", "Parte local dot-atom: átomos de at separados por un solo punto; sin punto inicial, final ni doble", "RFC 5322 dot-atom-text"],
        ["C3", "Parte local de 1 a 64 símbolos", "RFC 5321 §4.5.3.1"],
        ["C4", "Dominio con al menos dos etiquetas separadas por punto; cada etiqueta de 1 a 63 símbolos, solo letras, números y guion, sin guion al inicio ni al final", "RFC 1035 / 1123"],
        ["C5", "TLD (última etiqueta) formado solo por letras, mínimo 2", "Criterio práctico del capítulo 2"],
        ["C6", "Dominio de 1 a 255 símbolos", "RFC 5321 §4.5.3.1"],
    ],
    [0.4, 4.2, 1.9],
)
para(
    "Exclusiones deliberadas (documentadas en el capítulo 2): parte local entrecomillada (quoted-string), dominio "
    "como dirección IP literal, comentarios y espacios plegables (CFWS), sintaxis obsoleta y correo internacionalizado "
    "con Unicode (RFC 6531)."
)

# ---------------------------------------------------------------- 3.4
h2("3.4 Expresiones regulares (inciso d)")
para(
    "Notación: | es unión, la yuxtaposición es concatenación, * es cierre de Kleene, x+ = xx*, x? = (x|ε) y "
    "x{m,n} = x^m (x|ε)^(n−m). Las abreviaturas se expanden a las tres operaciones básicas antes de aplicar Thompson."
)
h3("Contraseñas  (Σ = L | l | N | S)")
code(
    "P1  Σ{15,64}\n"
    "P2  Σ* L Σ*\n"
    "P3  Σ* l Σ*\n"
    "P4  Σ* N Σ*\n"
    "P5  Σ* S Σ*\n"
    "P6  Σ* ( aaa | bbb | ... | ~~~ ) Σ*                       94 alternativas xxx, x ∈ Σ\n"
    "P7  Σ* ( (a|A)(b|B)(c|C)(d|D) | (d|D)(c|C)(b|B)(a|A) | ... | 6789 | 9876 ) Σ*   26 letras y 14 de dígitos\n"
    "P8  Σ* ( (p|P)(a|A)(s|S)(s|S)... | ... ) Σ*             una alternativa por palabra prohibida"
)
para(
    "La composición P2–P5 también se puede escribir como una única ER, unión de las 24 permutaciones de las cuatro "
    "clases:"
)
code("COMP = Σ*LΣ*lΣ*NΣ*SΣ* | Σ*LΣ*lΣ*SΣ*NΣ* | ... | Σ*SΣ*NΣ*lΣ*LΣ*      (4! = 24 términos)")
para(
    "Reconoce el mismo lenguaje que P2 ∩ P3 ∩ P4 ∩ P5 (comprobado con el programa). Su AFD mínimo tiene 16 estados, uno "
    "por cada subconjunto de clases ya vistas, y solo acepta cuando el subconjunto es completo. Sin embargo, el método "
    "del árbol genera para esta ER 12 034 estados antes de minimizar, por eso el diseño usa las reglas P2–P5 por separado."
)
para(
    "Contraseña aceptada  =  P0 ∧ P1 ∧ P2 ∧ P3 ∧ P4 ∧ P5 ∧ ¬P6 ∧ ¬P7 ∧ ¬P8", bold=True
)
h3("Correos  (at = A | N | H | E ;  X = at | P)")
code(
    "C1  X+ @ X+\n"
    "C2  at+ ( P at+ )*\n"
    "C3  X{1,64}\n"
    "C4  et ( P et )+          con  et = (A|N) ( (A|N|H){0,61} (A|N) )?\n"
    "C5  (A|N|H|P)* P A A+\n"
    "C6  X{1,255}"
)
para("Sin las cotas de longitud, todo el correo se resume en una sola ER (es la que se usa para el diagrama final):")
code("CORREO = at+ ( P at+ )* @ ( et P )+ A A+       con  et = (A|N) ( (A|N|H)* (A|N) )?")

# ---------------------------------------------------------------- 3.5
h2("3.5 Conversión a AFND con transiciones vacías: algoritmo de Thompson (inciso d.i)")
para(
    "Thompson construye el autómata por inducción sobre la ER. Cada fragmento tiene un único estado inicial y un único "
    "estado final, y se combinan con transiciones ε:"
)
fig_grid(["th_simbolo", "th_union"], 3.0, "Figura 1a-b. Fragmentos de Thompson: símbolo y unión.")
fig_grid(["th_concat", "th_estrella"], 3.0, "Figura 1c-d. Fragmentos de Thompson: concatenación y cierre de Kleene.")
para(
    "Aplicación a P2, Σ* L Σ* (Σ = L|l|N|S): dos cierres de Kleene sobre la unión de las cuatro etiquetas, unidos "
    "por el fragmento de la etiqueta L. Resultan 34 estados."
)
figure("nfa_P2", 6.5, "Figura 2. AFND-ε de P2 (inicio: 0, final: 19).")
para(
    "Aplicación a la parte local (C2), con la etiqueta a = atext para simplificar el dibujo:  a+ ( . a+ )*."
)
figure("nfa_local", 6.5, "Figura 3. AFND-ε de la parte local a+(.a+)*.")
para("Tamaños obtenidos para todas las reglas (los tres algoritmos coinciden en el lenguaje reconocido):")
nombres = {
    "P1 longitud 15-64": "P1  Σ{15,64}", "P2 mayuscula": "P2  Σ*LΣ*", "P6 repeticion (rechazo)": "P6  repetición",
    "P7 secuencia (rechazo)": "P7  secuencias", "P8 palabra prohibida (rechazo)": "P8  palabras",
    "P* composicion (24 permutaciones)": "COMP (24 perm.)", "C1 estructura": "C1  estructura",
    "C2 parte local": "C2  parte local", "C3 longitud local": "C3  local ≤ 64", "C4 etiquetas dominio": "C4  etiquetas",
    "C5 TLD": "C5  TLD", "C6 longitud dominio": "C6  dominio ≤ 255", "C* correo completo": "CORREO completo",
}
rows = []
for k, nm in nombres.items():
    s = ST[k]
    rows.append([nm, s["leaves"], s["nfa"], s["dfa_sub"], s["dfa_tree"], s["dfa_min"]])
table(
    ["ER", "Hojas", "AFND-ε (Thompson)", "AFD subconjuntos", "AFD árbol", "AFD mínimo"],
    rows, [1.8, 0.7, 1.2, 1.1, 0.9, 0.9],
)
para(
    "P3, P4 y P5 tienen exactamente el mismo tamaño que P2 (2 estados en el AFD), porque solo cambia la etiqueta "
    "central. La ER de P1 produce un contador lineal de 65 estados (0 a 64 símbolos leídos)."
)

# ---------------------------------------------------------------- 3.6
h2("3.6 Conversión a AFD: método del árbol y subconjuntos (inciso d.ii)")
para(
    "Método del árbol (followpos). Se aumenta la ER con un marcador final #, se numera cada hoja (posición) y se "
    "calculan nullable, firstpos, lastpos y followpos con estas reglas: nullable(ε) = verdadero y nullable(hoja) = falso; "
    "en una unión, firstpos y lastpos son la unión de los hijos; en una concatenación c1·c2, firstpos(c1) se amplía con "
    "firstpos(c2) solo si c1 es anulable (análogo para lastpos con c2); en un cierre x*, firstpos y lastpos son los de x. "
    "followpos se obtiene así: en c1·c2, cada posición de lastpos(c1) recibe firstpos(c2); en x*, cada posición de "
    "lastpos(x) recibe firstpos(x). Los estados del AFD son conjuntos de posiciones: el inicial es firstpos de la raíz, "
    "y desde un estado S con el símbolo a se llega a la unión de followpos(p) para las posiciones p ∈ S etiquetadas con a. "
    "Es de aceptación todo estado que contiene la posición de #."
)
para(
    "El algoritmo de subconjuntos parte del AFND-ε de Thompson: el estado inicial es la ε-clausura del inicial y cada "
    "transición es la ε-clausura de mover(S, a). Ambos métodos se implementaron y producen AFD equivalentes en las 16 ER "
    "(verificado por búsqueda en pares de estados); en estas ER el método del árbol da un AFD más compacto porque no "
    "arrastra los estados intermedios de Thompson."
)
h3("Ejemplo 1: P2 = Σ* L Σ*")
para(
    "Posiciones: 1–4 son las etiquetas L, l, N, S del primer Σ*; 5 es la etiqueta L central; 6–9 son las etiquetas del "
    "segundo Σ*; 10 es #."
)
table(
    ["Posición", "Etiqueta", "followpos"],
    [["1", "L", "{1,2,3,4,5}"], ["2", "l", "{1,2,3,4,5}"], ["3", "N", "{1,2,3,4,5}"], ["4", "S", "{1,2,3,4,5}"],
     ["5", "L", "{6,7,8,9,10}"], ["6", "L", "{6,7,8,9,10}"], ["7", "l", "{6,7,8,9,10}"], ["8", "N", "{6,7,8,9,10}"],
     ["9", "S", "{6,7,8,9,10}"], ["10", "#", "∅"]],
    [1.0, 1.0, 2.5],
)
para(
    "firstpos(raíz) = {1,2,3,4,5} = A. Desde A con L se alcanza followpos(1) ∪ followpos(5) = {1,…,10} = B; con l, N o S "
    "se queda en followpos de la posición correspondiente, {1,…,5} = A. Desde B, cualquier símbolo regresa a B. B contiene "
    "la posición 10 (#), por lo tanto es de aceptación."
)
figure("dfa_P2", 4.3, "Figura 4. AFD de P2 obtenido con el método del árbol (2 estados; B es de aceptación).")
h3("Ejemplo 2: parte local  a+ ( . a+ )*")
table(
    ["Posición", "Etiqueta", "followpos"],
    [["1", "a", "{2,3,6}"], ["2", "a", "{2,3,6}"], ["3", ".", "{4}"], ["4", "a", "{3,5,6}"], ["5", "a", "{3,5,6}"],
     ["6", "#", "∅"]],
    [1.0, 1.0, 2.5],
)
para(
    "A = firstpos = {1}. A con a: {2,3,6} = B (aceptación, contiene 6). B con a: B; B con «.»: {4} = C. C con a: {3,5,6} = D "
    "(aceptación). D con a: D; D con «.»: C. Al minimizar, A ≡ C y B ≡ D, quedando dos estados: «esperar átomo» y «átomo leído»."
)
figure("dfa_local", 5.2, "Figura 5. AFD de la parte local por el método del árbol.")
h3("Ejemplo 3: correo completo")
para(
    "Con la ER CORREO se obtiene un AFD de 12 estados (8 al minimizar). Las aristas están rotuladas con las etiquetas; "
    "A es el estado inicial y K el único de aceptación. La parte izquierda (A–D) reconoce la parte local y la arroba, "
    "y la derecha (E–K) reconoce las etiquetas del dominio: E acepta letras y números, G es «acaba de leer un guion» "
    "(no puede terminar una etiqueta ni seguir de punto), F es «acaba de leer un punto», y K solo se alcanza cuando la "
    "última etiqueta contiene únicamente letras."
)
figure("dfa_correo", 6.5, "Figura 6. AFD del correo completo (método del árbol, 12 estados).")

# ---------------------------------------------------------------- 3.7
h2("3.7 Verificación con casos de prueba")
para(
    "Los AFD construidos se contrastaron con implementaciones independientes de cada regla (expresiones de Python y "
    "comprobaciones directas) sobre 1 500 cadenas aleatorias por regla, sin discrepancias, y la ER de composición de "
    "24 permutaciones se contrastó con P2 ∩ P3 ∩ P4 ∩ P5. A continuación, el reporte por regla que produce el "
    "validador sobre credenciales de ejemplo."
)
h3("Contraseñas")
rows = []
for s in PW:
    rep = validar_password(s, R)
    rows.append([acorta(s.replace(" ", "␣")), "Válida" if es_valido(rep) else "Inválida", resumen(rep) or "—"])
table(["Entrada", "Resultado", "Reglas incumplidas"], rows, [2.3, 0.9, 3.3], mono_cols=(0,))
h3("Correos")
rows = []
for s in EM:
    rep = validar_correo(s, R)
    rows.append([acorta(s), "Válido" if es_valido(rep) else "Inválido", resumen(rep) or "—"])
table(["Entrada", "Resultado", "Reglas incumplidas"], rows, [2.5, 0.8, 3.2], mono_cols=(0,))
para(
    "Nota: «␣» marca un espacio en la entrada. La frase «Mi perro come 7 Huesos!» se evalúa como "
    "«Miperrocome7Huesos!» al descartar los espacios. «Password12345!!» ahora se rechaza por P8 (palabra prohibida), "
    "que en el capítulo 1 se había clasificado como regla no léxica. Del mismo modo, «usuario ejemplo@correo.com» se evalúa como «usuarioejemplo@correo.com» y resulta válido, porque el enunciado indica que los espacios no se toman en cuenta."
)

# ---------------------------------------------------------------- 3.8
h2("3.8 Alcance y limitaciones")
bullet(" No se validan direcciones con comillas, IP literal, comentarios ni Unicode; son formas válidas en RFC 5322 "
       "pero no las acepta ningún servicio público (capítulo 2).", "Correos.")
bullet(" P8 usa una lista finita de palabras. Las variantes con sustitución de caracteres (p4ssw0rd) y la comparación "
       "con contraseñas filtradas no son expresables con un autómata finito.", "Contraseñas.")
bullet(" La ER general de la política requiere intersección de lenguajes; se resuelve con un AFD por regla y un "
       "autómata producto, lo que además da el reporte de la regla que falla.", "Diseño.")

h2("Referencias adicionales de este capítulo")
para(
    "Aho, A., Lam, M., Sethi, R. y Ullman, J. (2006). Compilers: Principles, Techniques, and Tools (2.ª ed.). Pearson. "
    "(Algoritmo de Thompson, construcción de subconjuntos y método del árbol o followpos.)", align="left"
)
para(
    "Hopcroft, J., Motwani, R. y Ullman, J. (2007). Introduction to Automata Theory, Languages, and Computation "
    "(3.ª ed.). Pearson. (Cierre de los lenguajes regulares bajo intersección y complemento.)", align="left"
)

doc.save(OUT)
print("guardado", os.path.abspath(OUT))
