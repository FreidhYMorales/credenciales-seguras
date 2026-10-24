import random, time
import string
from reglas import build, PRINT, PALABRAS
from validador import validar_password, validar_correo, es_valido

t = time.time(); R = build(); print("build", round(time.time() - t), "s")
random.seed(7)

def rnd(alpha, lo, hi):
    return "".join(random.choice(alpha) for _ in range(random.randint(lo, hi)))

# 1) AFD vs oraculos independientes escritos a mano con `re`/Python
import re
V = "[!-~]"
lab = "[A-Za-z0-9](?:[A-Za-z0-9-]{0,61}[A-Za-z0-9])?"
atx = "[A-Za-z0-9!#$%&'*+/=?^_`{|}~-]"
SE = "[A-Za-z0-9!#$%&'*+/=?^_`{|}~.-]"
def secuencia(s):
    s = s.lower()
    for base in (string.ascii_lowercase, string.digits):
        for i in range(len(base) - 3):
            w = base[i:i + 4]
            if w in s or w[::-1] in s: return True
    return False
ORA = {
 "P1": lambda s: bool(re.fullmatch(V + "{15,64}", s)),
 "P2": lambda s: bool(re.fullmatch(V + "*[A-Z]" + V + "*", s)),
 "P3": lambda s: bool(re.fullmatch(V + "*[a-z]" + V + "*", s)),
 "P4": lambda s: bool(re.fullmatch(V + "*[0-9]" + V + "*", s)),
 "P5": lambda s: bool(re.fullmatch(V + "*[^A-Za-z0-9]" + V + "*", s)),
 "P6": lambda s: bool(re.fullmatch(V + "*", s)) and bool(re.search(r"(.)\1\1", s)),
 "P7": lambda s: bool(re.fullmatch(V + "*", s)) and secuencia(s),
 "P8": lambda s: bool(re.fullmatch(V + "*", s)) and any(w in s.lower() for w in PALABRAS),
 "C1": lambda s: bool(re.fullmatch(SE + "+@" + SE + "+", s)),
 "C2": lambda s: bool(re.fullmatch(atx + r"+(?:\." + atx + "+)*", s)),
 "C3": lambda s: bool(re.fullmatch(SE + "{1,64}", s)),
 "C4": lambda s: bool(re.fullmatch(lab + r"(?:\." + lab + ")+", s)),
 "C5": lambda s: bool(re.fullmatch(r"[A-Za-z0-9.-]*\.[A-Za-z]{2,}", s)),
 "C6": lambda s: bool(re.fullmatch(SE + "{1,255}", s)),
 "C*": lambda s: bool(re.fullmatch(atx + r"+(?:\." + atx + "+)*@(?:[A-Za-z0-9](?:[A-Za-z0-9-]*[A-Za-z0-9])?\\.)+[A-Za-z]{2,}", s)),
}
pools = {
 "P1": ("aA1!", 8, 70), "P2": ("aA1!", 0, 12), "P3": ("aA1!", 0, 12), "P4": ("aA1!", 0, 12),
 "P5": ("aA1!", 0, 12), "P6": ("aAb1!", 0, 14), "P7": ("abcdefABCD12345!", 0, 14),
 "P8": ("passwordADMINqwerty", 0, 14), "C1": ("a.@-", 0, 10), "C2": ("a.-_", 0, 10),
 "C3": ("ab.", 55, 70), "C4": ("aB1-.", 0, 14), "C5": ("aB1-.", 0, 10), "C6": ("ab.", 250, 260),
 "C*": ("aB1-._@!", 0, 16),
}
bad = 0; tot = 0; acc = 0
for k, c in R.items():
    key = k.split()[0]
    if key == "P*": continue
    alpha, lo, hi = pools[key]
    for i in range(1500):
        s = rnd(alpha, lo, hi)
        a, b = c.match(s), ORA[key](s); tot += 1; acc += a
        if a != b: bad += 1; print("DISCREPANCIA", k, repr(s), a, b)
print(f"cruce AFD vs oraculo: {tot} cadenas ({acc} aceptadas), {bad} discrepancias")

# 2) composicion: permutaciones vs P2..P5
import itertools
c = R["P* composicion (24 permutaciones)"]; d = 0
for i in range(3000):
    s = rnd("aA1!", 0, 7)
    x = c.match(s); y = all(R[k].match(s) for k in ("P2 mayuscula","P3 minuscula","P4 digito","P5 especial"))
    d += x != y
print("composicion 24 permutaciones == P2^P3^P4^P5, discrepancias:", d)

# 3) casos de prueba
pw = ["Pl4t0n0#Guatemala2026","Mi perro come 7 Huesos!","Abc123!","contrasenasegura","CONTRASENA2026!",
      "Password12345!!","Aaaa1111!!!!bbbb","  Segura2026#xyz  ","Xk9$mTq2#vLw8@Rz","Vj7!nQe4%Ha2&Ub6(Kd",
      "Tr0ub4dor&3xYz!qW9","abcd1234EFGH!@#$","Zq8#Lp3$Mn5@Wx1^Yt","H7$kP2#mQ9@xL4!zR8"]
print("\n--- CONTRASENAS")
for p in pw:
    r = validar_password(p, R)
    print("VALIDA  " if es_valido(r) else "INVALIDA", repr(p), "|", "; ".join(f"{n}:{d}" for n, ok, d in r if not ok))
em = ["cristian.lopez@ejemplo.com","usuario+etiqueta@correo.edu.gt","a_b-c'd!e@sub.dominio.org",".usuario@ejemplo.com",
      "usuario.@ejemplo.com","usuario..nombre@ejemplo.com","usuario@@ejemplo.com","usuario@ejemplo","usuario@-ejemplo.com",
      "usuario@ejemplo..com","usuario ejemplo@correo.com","usuario@ejemplo.c",'"juan perez"@ejemplo.com',"usuario@[192.168.1.10]",
      "a"*65+"@ejemplo.com","ana@"+"b"*64+".com","ana@ejemplo.c0m","ana@ejem_plo.com"]
print("\n--- CORREOS")
for e in em:
    r = validar_correo(e, R)
    print("VALIDO  " if es_valido(r) else "INVALIDO", repr(e if len(e) < 50 else e[:20]+"..."), "|", "; ".join(f"{n}:{d}" for n, ok, d in r if not ok))
