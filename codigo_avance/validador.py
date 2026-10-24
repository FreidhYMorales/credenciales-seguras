"""Validador lexico con reporte por regla (usa los AFD construidos en motor_er)."""
from reglas import PRINT, SIGMA_E_CHARS


def _rechazo(c, s):
    return c.match(s)


def validar_password(s, R):
    s = s.replace(" ", "")  # los espacios no se toman en cuenta
    out = []
    bad = [(i, ch) for i, ch in enumerate(s) if ch not in PRINT]
    out.append(("P0 alfabeto (ASCII visible)", not bad, f"simbolo no permitido en pos. {bad[0][0] + 1}" if bad else ""))
    if bad:
        return out
    out.append(("P1 longitud 15-64", R["P1 longitud 15-64"].match(s), f"longitud = {len(s)}"))
    for k, m in (("P2 mayuscula", "falta mayuscula"), ("P3 minuscula", "falta minuscula"),
                 ("P4 digito", "falta digito"), ("P5 especial", "falta caracter especial")):
        ok = R[k].match(s)
        out.append((k, ok, "" if ok else m))
    for k, m in (("P6 repeticion (rechazo)", "3+ caracteres identicos consecutivos"),
                 ("P7 secuencia (rechazo)", "secuencia ascendente/descendente de 4"),
                 ("P8 palabra prohibida (rechazo)", "contiene palabra prohibida")):
        hit = R[k].match(s)
        out.append((k, not hit, m if hit else ""))
    return out


def validar_correo(s, R):
    s = s.replace(" ", "")
    out = []
    bad = [(i, ch) for i, ch in enumerate(s) if ch not in SIGMA_E_CHARS]
    out.append(("C0 alfabeto", not bad, f"simbolo no permitido en pos. {bad[0][0] + 1}" if bad else ""))
    if bad:
        return out
    ok1 = R["C1 estructura"].match(s)
    out.append(("C1 estructura local@dominio", ok1, f"{s.count('@')} arrobas" if not ok1 else ""))
    if not ok1:
        return out
    local, dom = s.split("@")
    for k, txt, part in (("C2 parte local", "dot-atom invalido (punto inicial/final/doble)", local),
                         ("C3 longitud local", "local > 64", local),
                         ("C4 etiquetas dominio", "etiqueta invalida (guion en borde, vacia o >63) o sin punto", dom),
                         ("C5 TLD", "TLD debe ser solo letras y >= 2", dom),
                         ("C6 longitud dominio", "dominio > 255", dom)):
        ok = R[k].match(part)
        out.append((k, ok, "" if ok else txt))
    return out


def es_valido(rep):
    return all(ok for _, ok, _ in rep)
