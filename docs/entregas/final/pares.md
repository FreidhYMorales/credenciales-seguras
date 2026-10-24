### Pares válidos

Los diez pares válidos los generó la herramienta caminando el autómata producto. Las semillas quedan registradas, así que `make pairs` los reproduce byte a byte.

| # | Correo | Contraseña | Semillas (correo / contraseña) |
|---:|---|---|---|
| 1 | `IykLbhvynB@F.mT` | `>nslx613LK2%\|#PxFbA5` | 4490 / 2026 |
| 2 | `oJbKxhbAkQ@s.RC` | `5k5"*9k06T1=g[30kzL3` | 4491 / 2027 |
| 3 | `0RtXtuXs@r.tw` | ``s{`Rx0qa6uM2Q0om;88Y`` | 4492 / 2028 |
| 4 | `OzKFbIwWcQj@y.Fb` | `)AP<1[Np)8#189rkP>XG` | 4493 / 2029 |
| 5 | `bzKbbPAe@H.xq` | `K#J=(#2n<S<O/088z:pm` | 4494 / 2030 |
| 6 | `dA4WFKM5USNTInChTydt@h.FF` | `\%=L5ak5bS0Vb1vm:yXW` | 4495 / 2031 |
| 7 | `czdnlBUwvZO@x.kV` | `97r#<$^3kz3Bf!ifPd5~` | 4496 / 2032 |
| 8 | `fCkKwufDuCfAGxIVVVcDY@r.nh` | `BM"$1\CKE=svT{30-43d` | 4497 / 2033 |
| 9 | `vAlNkVl@A.WV` | `-I:9^7WGX4yt3\|[Y7-79` | 4498 / 2034 |
| 10 | `VrvFoLu@r.yP` | `11)x5\8L0S3'-hPiP56+` | 4499 / 2035 |

### Pares inválidos

Cada fila tiene exactamente una credencial defectuosa y nombra la regla que la rechaza; la otra mitad del par es válida.

| # | Correo | Contraseña | Qué falla |
|---:|---|---|---|
| 11 | `ana.perez@correo.com` | `Abc123!` | contraseña · **P1** Longitud 15-64 |
| 12 | `ana.perez@correo.com` | `Xk9mTq2vLw8RzQe4NbA` | contraseña · **P5** Especial |
| 13 | `ana.perez@correo.com` | `Xk$mTq#vLw@RzQe&NbA` | contraseña · **P4** Dígito |
| 14 | `ana.perez@correo.com` | `xk9$mtq2#vlw8@rzqe4&` | contraseña · **P2** Mayúscula |
| 15 | `ana.perez@correo.com` | `XK9$MTQ2#VLW8@RZQE4&` | contraseña · **P3** Minúscula |
| 16 | `ana.perez@correo.com` | `Xk9$mTq2#vLw8@RzQaaa` | contraseña · **P6** Repetición (rechazo) |
| 17 | `ana.perez@correo.com` | `Xk9$mTq2#vLw8@RzQabcd` | contraseña · **P7** Secuencia (rechazo) |
| 18 | `ana.perez@correo.com` | `Xk9$mTq2#vLw8@password` | contraseña · **P8** Palabra prohibida (rechazo) |
| 19 | `ana..perez@correo.com` | `Xk9$mTq2#vLw8@RzQe4&` | correo · **C2** Parte local |
| 20 | `ana.perez@correo.c` | `Xk9$mTq2#vLw8@RzQe4&` | correo · **C5** TLD |
