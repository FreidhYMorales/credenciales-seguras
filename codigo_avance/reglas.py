"""Alfabetos, etiquetas y ER de cada regla (contrasenas y correos)."""
import string

from motor_er import Compiled, Leaf, alt, cat, opt, plus, rep, Star

# ============================================================ ALFABETOS
PRINT = {chr(c) for c in range(0x21, 0x7F)}  # ASCII visible; el espacio (0x20) se ignora
UP, LO, DI = set(string.ascii_uppercase), set(string.ascii_lowercase), set(string.digits)
SP = PRINT - UP - LO - DI  # 32 especiales

L = Leaf(UP, "L")  # mayuscula
l = Leaf(LO, "l")  # minuscula
N = Leaf(DI, "N")  # numero
S = Leaf(SP, "S")  # especial
SIGMA_P = alt(L, l, N, S)  # Sigma de contrasenas (94 simbolos)

ATEXT_SPECIAL = set("!#$%&'*+-/=?^_`{|}~")  # 19 simbolos (incluye '-')
A = Leaf(UP | LO, "A")  # letra
H = Leaf({"-"}, "H")  # guion
E = Leaf(ATEXT_SPECIAL - {"-"}, "E")  # otros 18 atext
P = Leaf({"."}, "P")  # punto
AT = Leaf({"@"}, "@")
N_e = Leaf(DI, "N")
SIGMA_E_CHARS = UP | LO | DI | ATEXT_SPECIAL | {".", "@"}  # 83 simbolos

# ===================================================== CONTRASENAS
COMPOSICION = [("L", L), ("l", l), ("N", N), ("S", S)]
STAR_P = Star(SIGMA_P)


def contiene(x):
    return cat(STAR_P, x, STAR_P)


P_LONG = rep(SIGMA_P, 15, 64)  # Sigma{15,64}
P_MAYUS = contiene(L)
P_MINUS = contiene(l)
P_DIGITO = contiene(N)
P_ESPECIAL = contiene(S)

# ---- filtros de inseguridad: la ER describe el caso NO aceptado
def _c(ch):
    return Leaf({ch}, ch)


def _ci(ch):
    return Leaf({ch.lower(), ch.upper()}, ch)


REPETICION = contiene(alt(*[cat(_c(c), _c(c), _c(c)) for c in sorted(PRINT)]))

_letters, _digits = string.ascii_lowercase, string.digits
_runs = []
for base in (_letters, _digits):
    for i in range(len(base) - 3):
        w = base[i : i + 4]
        _runs += [w, w[::-1]]
SECUENCIA = contiene(alt(*[cat(*[_ci(c) for c in w]) for w in _runs]))

PALABRAS = ["password", "contrasena", "admin", "qwerty", "iloveyou", "welcome", "letmein"]
PROHIBIDAS = contiene(alt(*[cat(*[_ci(c) for c in w]) for w in PALABRAS]))

# ======================================================== CORREOS
at = alt(A, N_e, H, E)  # atext
X_NO_ARROBA = alt(A, N_e, H, E, P)
label = cat(alt(A, N_e), opt(cat(Star(alt(A, N_e, H)), alt(A, N_e))))
label63 = cat(alt(A, N_e), opt(cat(rep(alt(A, N_e, H), 0, 61), alt(A, N_e))))

C_ESTRUCTURA = cat(plus(X_NO_ARROBA), AT, plus(X_NO_ARROBA))
C_LOCAL = cat(plus(at), Star(cat(P, plus(at))))
C_LOCAL_LEN = rep(X_NO_ARROBA, 1, 64)
C_DOM_ETIQ = cat(label63, plus(cat(P, label63)))
C_TLD = cat(Star(alt(A, N_e, H, P)), P, A, plus(A))
C_DOM_LEN = rep(X_NO_ARROBA, 1, 255)

# ER completa (sin cotas de longitud) para el documento
C_COMPLETA = cat(plus(at), Star(cat(P, plus(at))), AT, plus(cat(label, P)), A, plus(A))

# --------------------------------------------------- permutaciones
import itertools

_T = dict(COMPOSICION)
COMP_PERM = alt(
    *[
        cat(*sum(([STAR_P, _T[t]] for t in perm), []), STAR_P)
        for perm in itertools.permutations("LlNS")
    ]
)


def build():
    R = {
        "P1 longitud 15-64": Compiled(P_LONG),
        "P2 mayuscula": Compiled(P_MAYUS),
        "P3 minuscula": Compiled(P_MINUS),
        "P4 digito": Compiled(P_DIGITO),
        "P5 especial": Compiled(P_ESPECIAL),
        "P6 repeticion (rechazo)": Compiled(REPETICION),
        "P7 secuencia (rechazo)": Compiled(SECUENCIA),
        "P8 palabra prohibida (rechazo)": Compiled(PROHIBIDAS),
        "P* composicion (24 permutaciones)": Compiled(COMP_PERM),
        "C1 estructura": Compiled(C_ESTRUCTURA),
        "C2 parte local": Compiled(C_LOCAL),
        "C3 longitud local": Compiled(C_LOCAL_LEN),
        "C4 etiquetas dominio": Compiled(C_DOM_ETIQ),
        "C5 TLD": Compiled(C_TLD),
        "C6 longitud dominio": Compiled(C_DOM_LEN),
        "C* correo completo": Compiled(C_COMPLETA),
    }
    return R
