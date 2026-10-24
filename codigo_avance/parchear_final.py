"""Quita la mencion a ISO 27002 (no verificada) y pasa la entropia del cap. 1 a 94 simbolos."""
import docx
d = docx.Document("docs/entregas/avance-1/Avance_1_completo.docx")
n = 0
for p in d.paragraphs:
    for r in p.runs:
        if "ISO/IEC 27002:2022" in r.text:
            r.text = ("Otro marco se alinea solo en parte con esta postura: PCI DSS 4.0.1 exige por defecto cambiar la contraseña cada 90 días en "
                      "cuentas protegidas solo con contraseña, pero permite omitir esa rotación cuando la seguridad de la cuenta se analiza de forma "
                      "dinámica y el acceso se decide en tiempo real.")
            n += 1
        if "uno de 95 caracteres imprimibles" in r.text:
            r.text = (r.text.replace("uno de 95 caracteres imprimibles", "uno de 94 caracteres visibles")
                      .replace("log2 95 / log2 26 ≈ 1.40", "log2 94 / log2 26 ≈ 1.39")
                      .replace("con los 95 imprimibles", "con los 94 visibles"))
            n += 1
assert n == 2, n
d.save("docs/entregas/avance-1/Avance_1_completo.docx")
print("ok")
