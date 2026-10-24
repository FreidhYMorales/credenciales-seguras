"""Render docs/entregas/final/documento.md as .docx and .pdf.

Needs python-docx (`pip install python-docx`) and LibreOffice (`soffice`), so
this is optional tooling: the generated files are committed and nobody needs
Python to build or run the application.

pandoc is not installed and LibreOffice's HTML import drops every image in
headless mode, so the Word file is built directly with python-docx and the PDF
is exported from that .docx.

The layout mirrors the graded first delivery: A4 with 2.5 cm margins, Times New
Roman 12 pt at 1.5 line spacing, Arial headings, Courier New for code.
"""
import os
import re
import subprocess
import sys

try:
    from docx import Document
except ImportError:  # pragma: no cover - tooling path, not part of the build
    raise SystemExit("falta python-docx: pip install python-docx")
from docx.enum.section import WD_SECTION
from docx.enum.table import WD_TABLE_ALIGNMENT
from docx.enum.text import WD_ALIGN_PARAGRAPH, WD_LINE_SPACING
from docx.oxml import OxmlElement
from docx.oxml.ns import qn
from docx.shared import Cm, Pt, RGBColor

BASE = os.path.dirname(os.path.dirname(os.path.dirname(os.path.abspath(__file__))))
SRC = os.path.join(BASE, "docs/entregas/final/documento.md")
DEST = os.path.join(BASE, "docs/entregas/final")
ANCHO = Cm(15.0)          # usable width: 21 - 3 - 2.5, a little margin to spare
SERIF = "Times New Roman"
SANS = "Arial"
MONO = "Courier New"


# ------------------------------------------------------------------ inline

def add_inline(par, text, size=12, bold=False, italic=False, mono=False):
    """Append text to a paragraph honouring **bold**, *italic* and `code`."""
    piezas = re.split(r"(\*\*.+?\*\*|``.+?``|`[^`]+`|<sub>.+?</sub>|<sup>.+?</sup>"
                      r"|<https?://[^>]+>|\*[^*\s][^*]*?\*|\\[*_`\\|])", text, flags=re.S)
    for pieza in piezas:
        if not pieza:
            continue
        b, i, m = bold, italic, mono
        if pieza.startswith("**") and pieza.endswith("**") and len(pieza) > 4:
            add_inline(par, pieza[2:-2], size, True, i, m)
            continue
        if pieza.startswith("<sub>") or pieza.startswith("<sup>"):
            sub = pieza.startswith("<sub>")
            run = par.add_run(pieza[5:-6])
            run.bold, run.italic = b, i
            run.font.size = Pt(size - 1)
            run.font.name = SERIF
            if sub:
                run.font.subscript = True
            else:
                run.font.superscript = True
            continue
        if pieza.startswith("<http"):
            run = par.add_run(pieza[1:-1])
            run.bold, run.italic = b, i
            run.font.size = Pt(size)
            run.font.name = SERIF
            continue
        if pieza.startswith("``") and pieza.endswith("``") and len(pieza) > 4:
            add_inline(par, pieza[2:-2], size - 1, b, i, True)
            continue
        if pieza.startswith("`") and pieza.endswith("`") and len(pieza) > 2:
            add_inline(par, pieza[1:-1], size - 1, b, i, True)
            continue
        if pieza.startswith("*") and pieza.endswith("*") and len(pieza) > 2:
            add_inline(par, pieza[1:-1], size, b, True, m)
            continue
        if pieza.startswith("\\") and len(pieza) == 2:
            pieza = pieza[1]
        run = par.add_run(pieza)
        run.bold, run.italic = b, i
        run.font.size = Pt(size)
        run.font.name = MONO if m else SERIF
        if m:
            run._element.rPr.rFonts.set(qn("w:eastAsia"), MONO)


def shade(element, fill):
    """Add w:shd in schema order, whether element is a pPr, tcPr or trPr."""
    shd = OxmlElement("w:shd")
    shd.set(qn("w:val"), "clear")
    shd.set(qn("w:color"), "auto")
    shd.set(qn("w:fill"), fill)
    if element.tag == qn("w:pPr"):
        insertar(element, shd, P_PR_DESPUES_DE_SHD)
    else:
        element.append(shd)


# Children of w:pPr and w:tblPr that must stay after the element we insert.
P_PR_DESPUES_DE_SHD = (
    "w:tabs", "w:suppressAutoHyphens", "w:kinsoku", "w:wordWrap", "w:overflowPunct",
    "w:topLinePunct", "w:autoSpaceDE", "w:autoSpaceDN", "w:bidi", "w:adjustRightInd",
    "w:snapToGrid", "w:spacing", "w:ind", "w:contextualSpacing", "w:mirrorIndents",
    "w:suppressOverlap", "w:jc", "w:textDirection", "w:textAlignment",
    "w:textboxTightWrap", "w:outlineLvl", "w:divId", "w:cnfStyle", "w:rPr",
    "w:sectPr", "w:pPrChange",
)


def insertar(pPr, el, despues):
    """Insert el before the first child listed in despues (schema order)."""
    pPr.insert_element_before(el, *despues)


def border(par, sides=("top", "left", "bottom", "right")):
    pPr = par._p.get_or_add_pPr()
    pBdr = OxmlElement("w:pBdr")
    for side in sides:
        el = OxmlElement(f"w:{side}")
        el.set(qn("w:val"), "single")
        el.set(qn("w:sz"), "4")
        el.set(qn("w:space"), "4")
        el.set(qn("w:color"), "C8C8C8")
        pBdr.append(el)
    insertar(pPr, pBdr, P_PR_DESPUES_DE_SHD)


# ------------------------------------------------------------------- blocks

def parse_blocks(lines):
    """Turn the Markdown body into a flat list of blocks."""
    bloques, i = [], 0
    while i < len(lines):
        linea = lines[i]
        if not linea.strip() or re.match(r"^-{3,}$", linea.strip()):
            i += 1
            continue

        if linea.startswith("```"):
            j = i + 1
            while j < len(lines) and not lines[j].startswith("```"):
                j += 1
            bloques.append(("code", lines[i + 1:j]))
            i = j + 1
            continue

        m = re.match(r"^(#{2,3}) (.+)$", linea)
        if m:
            bloques.append(("h%d" % len(m.group(1)), m.group(2).strip()))
            i += 1
            continue

        m = re.match(r"^!\[(.*?)\]\((.*?)\)$", linea.strip())
        if m:
            bloques.append(("image", m.group(1), m.group(2)))
            i += 1
            continue

        if linea.lstrip().startswith("|"):
            filas = []
            while i < len(lines) and lines[i].lstrip().startswith("|"):
                filas.append(lines[i].strip())
                i += 1
            filas = [f for f in filas if not re.match(r"^\|[\s:|-]+\|$", f)]
            # An escaped pipe inside a code span is part of the cell, not a
            # column boundary.
            celdas = [[c.strip().replace("\\|", "|")
                       for c in re.split(r"(?<!\\)\|", f.strip("|"))] for f in filas]
            bloques.append(("table", celdas))
            continue

        if re.match(r"^\s*[-*] ", linea):
            items = []
            while i < len(lines) and (re.match(r"^\s*[-*] ", lines[i])
                                      or (lines[i].startswith("  ") and lines[i].strip())):
                if re.match(r"^\s*[-*] ", lines[i]):
                    items.append(re.sub(r"^\s*[-*] ", "", lines[i]).strip())
                else:
                    items[-1] += " " + lines[i].strip()
                i += 1
            bloques.append(("ul", items))
            continue

        if re.match(r"^\s*\d+\. ", linea):
            items = []
            while i < len(lines) and (re.match(r"^\s*\d+\. ", lines[i])
                                      or (lines[i].startswith("  ") and lines[i].strip())):
                if re.match(r"^\s*\d+\. ", lines[i]):
                    items.append(re.sub(r"^\s*\d+\. ", "", lines[i]).strip())
                else:
                    items[-1] += " " + lines[i].strip()
                i += 1
            bloques.append(("ol", items))
            continue

        parrafo = [linea.strip()]
        i += 1
        while i < len(lines) and lines[i].strip() and not re.match(
                r"^(#{2,3} |```|\||!\[|\s*[-*] |\s*\d+\. )", lines[i]):
            parrafo.append(lines[i].strip())
            i += 1
        bloques.append(("p", " ".join(parrafo)))
    return bloques


# ----------------------------------------------------------------- document

def main():
    crudo = open(SRC, encoding="utf-8").read().splitlines()
    corte = next(n for n, l in enumerate(crudo) if l.strip() == "---")
    titulo = crudo[0].lstrip("# ").strip()
    meta = [l.strip() for l in crudo[1:corte] if l.strip()]
    bloques = parse_blocks(crudo[corte + 1:])

    doc = Document()
    doc.core_properties.title = titulo
    doc.core_properties.author = "Cristian Antonio Mendoza Mendoza"

    # Page setup and the default paragraph style.
    sec = doc.sections[0]
    sec.page_width, sec.page_height = Cm(21), Cm(29.7)
    sec.top_margin, sec.bottom_margin = Cm(2.5), Cm(2.5)
    sec.left_margin, sec.right_margin = Cm(3), Cm(2.5)

    normal = doc.styles["Normal"]
    normal.font.name = SERIF
    normal.font.size = Pt(12)
    normal.element.rPr.rFonts.set(qn("w:eastAsia"), SERIF)
    normal.paragraph_format.space_after = Pt(6)
    normal.paragraph_format.line_spacing = 1.5
    normal.paragraph_format.alignment = WD_ALIGN_PARAGRAPH.JUSTIFY

    def parrafo(align=WD_ALIGN_PARAGRAPH.JUSTIFY, space_after=6, space_before=0,
                spacing=1.5, keep=False):
        p = doc.add_paragraph()
        pf = p.paragraph_format
        pf.alignment = align
        pf.space_after = Pt(space_after)
        pf.space_before = Pt(space_before)
        if spacing == 1:
            pf.line_spacing_rule = WD_LINE_SPACING.SINGLE
        else:
            pf.line_spacing = spacing
        pf.keep_with_next = keep
        return p

    # ------------------------------------------------------------- portada
    p = parrafo(WD_ALIGN_PARAGRAPH.CENTER, space_after=18, space_before=60, spacing=1)
    add_inline(p, titulo, size=16, bold=True)
    for run in p.runs:
        run.font.name = SANS
        run._element.rPr.rFonts.set(qn("w:eastAsia"), SANS)

    p = parrafo(WD_ALIGN_PARAGRAPH.CENTER, space_after=24, spacing=1)
    border(p, sides=("top",))

    for linea in meta:
        p = parrafo(WD_ALIGN_PARAGRAPH.CENTER, space_after=4, spacing=1)
        add_inline(p, linea, size=12)

    # --------------------------------------------------------------- cuerpo
    visto_h2 = False
    for bloque in bloques:
        tipo = bloque[0]

        if tipo == "h2":
            p = parrafo(WD_ALIGN_PARAGRAPH.LEFT, space_after=10, space_before=12, spacing=1, keep=True)
            # The body flows like the graded first delivery; only the abstract
            # starts a page, so the cover stands alone.
            if not visto_h2:
                p.paragraph_format.page_break_before = True
                visto_h2 = True
            run = p.add_run(bloque[1])
            run.bold = True
            run.font.size = Pt(13)
            run.font.name = SANS
            run.font.color.rgb = RGBColor(0, 0, 0)

        elif tipo == "h3":
            p = parrafo(WD_ALIGN_PARAGRAPH.LEFT, space_after=6, space_before=12, spacing=1, keep=True)
            run = p.add_run(bloque[1])
            run.bold = True
            run.font.size = Pt(12)
            run.font.name = SANS
            run.font.color.rgb = RGBColor(0, 0, 0)

        elif tipo == "p":
            p = parrafo()
            add_inline(p, bloque[1])

        elif tipo in ("ul", "ol"):
            # Manual markers with a hanging indent: Word's List Number style
            # keeps counting across the document, so the second list would
            # start at 11 instead of 1.
            for n, item in enumerate(bloque[1], start=1):
                p = parrafo(space_after=4)
                pf = p.paragraph_format
                pf.left_indent = Cm(1.0)
                pf.first_line_indent = Cm(-1.0)
                pf.tab_stops.add_tab_stop(Cm(1.0))
                marca = "\u2022" if tipo == "ul" else f"{n}."
                run = p.add_run(marca + "\t")
                run.font.size = Pt(12)
                run.font.name = SERIF
                add_inline(p, item)

        elif tipo == "code":
            p = parrafo(WD_ALIGN_PARAGRAPH.LEFT, space_after=10, space_before=6, spacing=1)
            p.paragraph_format.keep_together = True
            shade(p._p.get_or_add_pPr(), "F5F5F5")
            border(p)
            for n, linea in enumerate(bloque[1]):
                run = p.add_run(linea)
                run.font.name = MONO
                run._element.rPr.rFonts.set(qn("w:eastAsia"), MONO)
                run.font.size = Pt(9.5)
                if n < len(bloque[1]) - 1:
                    run.add_break()

        elif tipo == "image":
            _, alt, ruta = bloque
            p = parrafo(WD_ALIGN_PARAGRAPH.CENTER, space_after=2, space_before=6, spacing=1, keep=True)
            p.add_run().add_picture(os.path.join(DEST, ruta), width=ANCHO)

        elif tipo == "table":
            filas = bloque[1]
            columnas = max(len(f) for f in filas)
            tabla = doc.add_table(rows=len(filas), cols=columnas)
            tabla.style = "Table Grid"
            tabla.alignment = WD_TABLE_ALIGNMENT.CENTER
            # Column widths: every column gets room for its longest
            # unbreakable token (a password cannot wrap), and whatever is left
            # is shared in proportion to the longest cell. Equal columns used to
            # squeeze a 20-character password into three lines.
            fs = 9 if columnas >= 4 else 10
            em = fs * 0.03528                      # centimetres per em
            util = 15.5

            def ancho_cm(texto, fs=fs):
                limpio = texto.replace("`", "").replace("**", "").replace("*", "")
                mono = "`" in texto
                return len(limpio) * ((0.62 if mono else 0.47) * fs * 0.03528) + 0.25

            def piso_cm(texto, fs=fs):
                limpio = texto.replace("`", "").replace("**", "").replace("*", "")
                mono = "`" in texto
                trozos = [limpio] if mono else limpio.split()
                largo = max((len(t) for t in trozos), default=1)
                return largo * ((0.62 if mono else 0.47) * fs * 0.03528) + 0.25

            techo, piso = [], []
            for c in range(columnas):
                celdas_col = [f[c] for f in filas if c < len(f)]
                techo.append(max(ancho_cm(x) for x in celdas_col))
                piso.append(max(piso_cm(x) for x in celdas_col))

            if sum(techo) <= util:
                sobra = (util - sum(techo)) / columnas
                anchos = [Cm(t + sobra) for t in techo]
            else:
                holgura = util - sum(piso)
                if holgura <= 0:
                    k = util / sum(piso)
                    anchos = [Cm(t * k) for t in piso]
                else:
                    extra = [t - b for t, b in zip(techo, piso)]
                    total_extra = sum(extra) or 1
                    anchos = [Cm(b + holgura * e / total_extra)
                              for b, e in zip(piso, extra)]
            # LibreOffice honours the table grid, not the cell widths.
            for c, ancho in enumerate(anchos):
                tabla.columns[c].width = ancho
                for fila in tabla.rows:
                    fila.cells[c].width = ancho
            tabla.autofit = False
            # Tighter cell margins buy space for the 20-character passwords.
            mar = OxmlElement("w:tblCellMar")
            for lado in ("left", "right"):
                el = OxmlElement(f"w:{lado}")
                el.set(qn("w:w"), "57")
                el.set(qn("w:type"), "dxa")
                mar.append(el)
            tabla._tbl.tblPr.insert_element_before(
                mar, "w:tblLook", "w:tblCaption", "w:tblDescription", "w:tblPrChange")
            for f, fila in enumerate(filas):
                for c in range(columnas):
                    celda = tabla.cell(f, c)
                    celda.text = ""
                    p = celda.paragraphs[0]
                    p.paragraph_format.space_after = Pt(2)
                    p.paragraph_format.line_spacing = 1
                    p.paragraph_format.alignment = WD_ALIGN_PARAGRAPH.LEFT
                    texto = fila[c] if c < len(fila) else ""
                    add_inline(p, texto, size=fs, bold=(f == 0))
                fila_pr = tabla.rows[f]._tr.get_or_add_trPr()
                fila_pr.append(OxmlElement("w:cantSplit"))
                if f == 0:
                    trPr = tabla.rows[0]._tr.get_or_add_trPr()
                    trPr.append(OxmlElement("w:tblHeader"))
                    for c in range(columnas):
                        shade(tabla.cell(0, c)._tc.get_or_add_tcPr(), "EDEDED")
            # Keep the table with the text that follows it out of reach: a
            # paragraph mark after a table is mandatory in Word.
            parrafo(space_after=10, spacing=1)

    # ------------------------------------------------------------- footer
    pie = sec.footer.paragraphs[0]
    pie.alignment = WD_ALIGN_PARAGRAPH.CENTER
    run = pie.add_run("Página ")
    run.font.size = Pt(10)
    run.font.name = SERIF
    fld = OxmlElement("w:fldSimple")
    fld.set(qn("w:instr"), "PAGE")
    r = OxmlElement("w:r")
    t = OxmlElement("w:t")
    t.text = "1"
    r.append(t)
    fld.append(r)
    pie._p.append(fld)

    salida = os.path.join(DEST, "Credenciales_seguras.docx")
    doc.save(salida)
    print("generado", salida)

    subprocess.run(["soffice", "--headless", "--norestore", "--convert-to", "pdf",
                    "--outdir", DEST, salida],
                   check=True, stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL)
    print("generado", os.path.join(DEST, "Credenciales_seguras.pdf"))


if __name__ == "__main__":
    sys.exit(main())
