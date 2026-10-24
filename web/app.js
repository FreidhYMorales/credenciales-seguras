'use strict';

// ------------------------------------------------------------------ helpers

const $ = (sel, root = document) => root.querySelector(sel);
const $$ = (sel, root = document) => Array.from(root.querySelectorAll(sel));

function esc(s) {
  return String(s ?? '').replace(/[&<>"']/g, (c) => ({
    '&': '&amp;', '<': '&lt;', '>': '&gt;', '"': '&quot;', "'": '&#39;',
  }[c]));
}

async function api(path, options) {
  const res = await fetch(path, options);
  let body;
  try {
    body = await res.json();
  } catch {
    throw new Error(`respuesta inválida del servidor (HTTP ${res.status})`);
  }
  if (!body.ok) throw new Error(body.error || `HTTP ${res.status}`);
  return body.datos;
}

function postJSON(path, payload) {
  return api(path, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(payload),
  });
}

function errorBox(err) {
  return `<div class="error">${esc(err.message || err)}</div>`;
}

// -------------------------------------------------------------------- state

const state = {
  tipo: 'password',
  reglas: [],
  regla: null,
  vista: 'sub',
  pagina: 1,
  centro: null,
  filtro: 'todas',
  pares: [],
  genTipo: 'password',
};

// --------------------------------------------------------------------- tabs

function showScreen(name) {
  $$('.screen').forEach((s) => s.classList.toggle('active', s.id === `screen-${name}`));
  $$('#tabs button').forEach((b) => b.setAttribute('aria-selected', String(b.dataset.screen === name)));
  if (name === 'explorador') loadReglas().then(renderSidebar);
  if (name === 'reglas') loadReglas().then(renderReglas);
  if (name === 'pares') loadPares();
}

$$('#tabs button').forEach((b) => b.addEventListener('click', () => showScreen(b.dataset.screen)));

// ------------------------------------------------------------------ validar

function setTipo(tipo) {
  state.tipo = tipo;
  $$('#tipo-seg button').forEach((b) => b.setAttribute('aria-pressed', String(b.dataset.tipo === tipo)));
  const par = tipo === 'par';
  $('#simple-campos').hidden = par;
  $('#par-campos').hidden = !par;
  if (!par) $('#valor').placeholder = tipo === 'password' ? 'Escribí una contraseña' : 'Escribí un correo';
}

$$('#tipo-seg button').forEach((b) => b.addEventListener('click', () => setTipo(b.dataset.tipo)));

$('#ver').addEventListener('click', () => {
  const input = $('#valor');
  const hidden = input.type === 'password';
  input.type = hidden ? 'text' : 'password';
  $('#ver').textContent = hidden ? 'ocultar' : 'ver';
});

async function validar() {
  const valor = $('#valor').value;
  const out = $('#reporte');
  if (!valor.trim()) {
    out.innerHTML = `<div class="card"><div class="empty">Escribí un valor para validar.</div></div>`;
    return;
  }
  out.innerHTML = `<div class="card"><div class="empty">Validando…</div></div>`;
  try {
    const rep = await postJSON('/api/validar', { tipo: state.tipo, valor });
    out.innerHTML = renderReporte(rep, valor);
  } catch (err) {
    out.innerHTML = `<div class="card">${errorBox(err)}</div>`;
  }
}

// validarPar validates the two credentials at the same time and joins the two
// reports under one verdict. The pair is valid only when both halves are.
async function validarPar() {
  const email = $('#par-email').value;
  const pass = $('#par-pass').value;
  const out = $('#reporte');
  if (!email.trim() || !pass.trim()) {
    out.innerHTML = `<div class="card"><div class="empty">Escribí el correo y la contraseña del par.</div></div>`;
    return;
  }
  out.innerHTML = `<div class="card"><div class="empty">Validando el par…</div></div>`;
  try {
    const [repEmail, repPass] = await Promise.all([
      postJSON('/api/validar', { tipo: 'email', valor: email }),
      postJSON('/api/validar', { tipo: 'password', valor: pass }),
    ]);
    const valido = repEmail.valido && repPass.valido;
    const fallas = [];
    if (!repEmail.valido) fallas.push('el correo');
    if (!repPass.valido) fallas.push('la contraseña');
    const detalle = valido ? 'correo y contraseña válidos' : `falla ${fallas.join(' y ')}`;
    const veredicto = `<div class="card">
      <div class="verdict">
        <span class="pill ${valido ? 'ok' : 'bad'}">${valido ? 'Par válido' : 'Par inválido'}</span>
        <span>${esc(detalle)}</span>
      </div>
    </div>`;
    out.innerHTML = veredicto
      + renderReporte(repEmail, email, 'Correo')
      + renderReporte(repPass, pass, 'Contraseña');
  } catch (err) {
    out.innerHTML = `<div class="card">${errorBox(err)}</div>`;
  }
}

function renderReporte(rep, original, titulo) {
  const pill = rep.valido ? 'ok' : 'bad';
  const etiqueta = rep.valido ? 'VÁLIDA' : 'INVÁLIDA';
  const fallas = rep.reglas.filter((r) => r.evaluado && !r.ok).length;
  const detalle = rep.valido
    ? `${rep.longitud} caracteres`
    : `${fallas} regla${fallas === 1 ? '' : 's'} incumplida${fallas === 1 ? '' : 's'}`;

  let nivel = '';
  if (rep.nivel) {
    const cls = rep.nivel === 'Fuerte' ? 'ok' : rep.nivel === 'Media' ? 'warn' : 'bad';
    const bits = rep.entropia != null ? ` · ${rep.entropia} bits (bruto)` : '';
    nivel = `<span class="pill ${cls}">Nivel: ${esc(rep.nivel)}${bits}</span>`;
  }

  const filas = rep.reglas.map((r) => {
    if (!r.evaluado) {
      return `<tr class="skipped"><td class="mono">${esc(r.id)}</td><td>${esc(r.nombre)}</td>
        <td class="mark-skip">—</td><td>no evaluada</td></tr>`;
    }
    const marca = r.ok ? '<span class="mark-ok">✔</span>' : '<span class="mark-bad">✘</span>';
    return `<tr><td class="mono">${esc(r.id)}</td><td>${esc(r.nombre)}</td>
      <td>${marca}</td><td>${esc(r.mensaje || '')}</td></tr>`;
  }).join('');

  const normalizado = rep.normalizado !== original
    ? `<div class="meta">Normalizado: <span class="mono">${esc(rep.normalizado)}</span></div>`
    : '';

  return `<div class="card">
    ${titulo ? `<h2>${esc(titulo)}</h2>` : ''}
    <div class="verdict">
      <span class="pill ${pill}">${etiqueta}</span>
      <span>${esc(detalle)}</span>
      ${nivel}
    </div>
    ${normalizado}
    <table>
      <thead><tr><th>Regla</th><th>Nombre</th><th></th><th>Detalle</th></tr></thead>
      <tbody>${filas}</tbody>
    </table>
  </div>`;
}

$('#validar').addEventListener('click', validar);
$('#valor').addEventListener('keydown', (e) => { if (e.key === 'Enter') validar(); });
$('#validar-par').addEventListener('click', validarPar);
$('#par-ver').addEventListener('click', () => {
  const input = $('#par-pass');
  const oculto = input.type === 'password';
  input.type = oculto ? 'text' : 'password';
  $('#par-ver').textContent = oculto ? 'ocultar' : 'ver';
});
$('#par-email').addEventListener('keydown', (e) => { if (e.key === 'Enter') validarPar(); });
$('#par-pass').addEventListener('keydown', (e) => { if (e.key === 'Enter') validarPar(); });

// ------------------------------------------------------------------ generar

$$('#gen-tipo button').forEach((b) => b.addEventListener('click', () => {
  state.genTipo = b.dataset.tipo;
  $$('#gen-tipo button').forEach((x) => x.setAttribute('aria-pressed', String(x.dataset.tipo === state.genTipo)));
  $('#gen-longitud-wrap').style.display = state.genTipo === 'password' ? '' : 'none';
}));

$('#gen-longitud').addEventListener('input', (e) => { $('#gen-longitud-val').textContent = e.target.value; });

$('#generar').addEventListener('click', async () => {
  const out = $('#gen-resultado');
  const camino = $('#gen-camino');
  const payload = { tipo: state.genTipo };
  const semillaRaw = $('#gen-semilla').value.trim();
  if (semillaRaw !== '') payload.semilla = Number(semillaRaw);
  if (state.genTipo === 'password') payload.longitud = Number($('#gen-longitud').value);

  out.innerHTML = `<div class="empty">Generando…</div>`;
  camino.innerHTML = `<span class="empty">Generando…</span>`;
  try {
    const res = await postJSON('/api/generar', payload);
    const cls = res.nivel === 'Fuerte' ? 'ok' : res.nivel === 'Media' ? 'warn' : 'bad';
    const nivel = res.nivel
      ? `<span class="pill ${cls}">${esc(res.nivel)} · ${res.longitud} caracteres · ${res.entropia} bits</span>`
      : `<span class="pill neutral">correo · ${res.longitud} caracteres</span>`;
    out.innerHTML = `
      <h2>Resultado</h2>
      <div class="result" id="gen-valor">${esc(res.valor)}</div>
      <div class="verdict" style="margin-top:12px">${nivel}</div>
      <div class="meta">semilla <b class="mono">${esc(res.semilla)}</b> · reproducir con el mismo tipo${res.tipo === 'password' ? ' y longitud' : ''}</div>
      <div class="row" style="margin-top:14px">
        <button class="ghost" id="gen-copiar">Copiar</button>
        <button class="ghost" id="gen-validar">Validar esta</button>
      </div>`;
    $('#gen-copiar').addEventListener('click', () => {
      navigator.clipboard?.writeText(res.valor);
      $('#gen-copiar').textContent = 'Copiado';
      setTimeout(() => { $('#gen-copiar').textContent = 'Copiar'; }, 1200);
    });
    $('#gen-validar').addEventListener('click', () => {
      setTipo(res.tipo === 'email' ? 'email' : 'password');
      $('#valor').value = res.valor;
      showScreen('validar');
      validar();
    });
    camino.innerHTML = renderCamino(res.camino);
  } catch (err) {
    out.innerHTML = errorBox(err);
    camino.innerHTML = '';
  }
});

function renderCamino(camino) {
  if (!camino || !camino.length) return '<span class="empty">Sin recorrido.</span>';
  const partes = [];
  partes.push(`<span class="st">${esc(camino[0].estado)}</span>`);
  for (const paso of camino) {
    partes.push(`─<span class="cls">${esc(paso.clase)}</span>→ <span class="st">${esc(paso.estado)}</span>`);
  }
  return partes.join(' ');
}

// --------------------------------------------------------------- explorador

async function loadReglas() {
  if (state.reglas.length) return state.reglas;
  state.reglas = await api('/api/reglas');
  return state.reglas;
}

function renderSidebar() {
  const cont = $('#explorador-lista');
  const grupos = [
    { titulo: 'Contraseñas', prefijo: 'P' },
    { titulo: 'Correos', prefijo: 'C' },
  ];
  const html = grupos.map((g) => {
    const filas = state.reglas
      .filter((r) => (g.prefijo === 'P' ? r.id[0] === 'P' : r.id[0] === 'C'))
      .map((r) => {
        const n = r.conteos ? r.conteos.min : '—';
        const sel = state.regla && state.regla.id === r.id;
        return `<button data-id="${esc(r.id)}" aria-pressed="${sel}">
          <span><span class="mono">${esc(r.id)}</span> ${esc(r.nombre)}</span>
          <span class="count">${n}</span></button>`;
      }).join('');
    return `<div class="group">${g.titulo}</div>${filas}`;
  }).join('');
  cont.innerHTML = html;
  $$('#explorador-lista button').forEach((b) => b.addEventListener('click', () => selectRegla(b.dataset.id)));
}

function selectRegla(id) {
  state.regla = state.reglas.find((r) => r.id === id) || null;
  state.pagina = 1;
  state.centro = null;
  renderSidebar();
  cargarVista();
}

$$('#explorador-vistas button').forEach((b) => b.addEventListener('click', () => {
  state.vista = b.dataset.vista;
  state.pagina = 1;
  state.centro = null;
  $$('#explorador-vistas button').forEach((x) => x.setAttribute('aria-pressed', String(x.dataset.vista === state.vista)));
  cargarVista();
}));

async function cargarVista() {
  const r = state.regla;
  if (!r) return;
  $('#explorador-titulo').textContent = `${r.id} · ${r.nombre}`;

  if (!r.tiene_er) {
    $('#explorador-conteos').innerHTML = '';
    ocultarSubcontroles();
    $('#explorador-grafo').innerHTML = `<div class="note">${esc(r.id)} es una verificación de conjunto sobre el alfabeto: no tiene autómata.</div>`;
    $('#explorador-tabla').innerHTML = '<div class="empty">—</div>';
    $('#explorador-pager').innerHTML = '';
    return;
  }

  const c = r.conteos;
  $('#explorador-conteos').innerHTML =
    `<span>AFND-ε <b>${c.nfa}</b></span><span>Subconjuntos <b>${c.sub}</b></span>` +
    `<span>Árbol <b>${c.tree}</b></span><span>Mínimo <b>${c.min}</b></span>`;

  try {
    const datos = await api(urlVista());
    if (datos.svg) {
      ocultarSubcontroles();
      $('#explorador-grafo').innerHTML = `<div class="graph">${datos.svg}</div>`;
    } else {
      // Too large for the pre-generated SVG: ask for a BFS ball and draw it.
      await cargarSubgrafo();
    }
    renderTransiciones(datos);
  } catch (err) {
    $('#explorador-grafo').innerHTML = errorBox(err);
  }
}

function urlVista() {
  const r = state.regla;
  return `/api/automatas/${encodeURIComponent(r.id)}?vista=${state.vista}&pagina=${state.pagina}&limite=50`;
}

function urlSubgrafo() {
  const profundidad = Number($('#sub-profundidad').value) || 2;
  return `${urlVista()}&profundidad=${profundidad}&centro=${encodeURIComponent(state.centro || 'q0')}`;
}

function ocultarSubcontroles() {
  $('#sub-controls').hidden = true;
  $('#sub-info').textContent = '';
}

async function cargarSubgrafo() {
  const r = state.regla;
  if (!r || !r.tiene_er) return;
  const turno = ++subTurno;
  try {
    const datos = await api(urlSubgrafo());
    if (turno !== subTurno) return; // a newer request already won
    mostrarSubgrafo(datos.subgrafo);
  } catch (err) {
    if (turno !== subTurno) return;
    $('#explorador-grafo').innerHTML = errorBox(err);
  }
}

let subTurno = 0;

function mostrarSubgrafo(sg) {
  const caja = $('#explorador-grafo');
  $('#sub-controls').hidden = false;
  if (!sg || !sg.nodos || !sg.nodos.length) {
    caja.innerHTML = `<div class="note">No hay vecindad para el estado pedido.</div>`;
    $('#sub-info').textContent = '';
    return;
  }
  caja.innerHTML = `<div class="graph">${renderSubgrafo(sg)}</div>`;
  const arbol = sg.aristas_totales > sg.aristas.length
    ? ` · árbol de ${sg.aristas.length} de ${sg.aristas_totales} aristas`
    : '';
  $('#sub-info').textContent =
    `vecindad de ${sg.centro} · ${sg.nodos.length} de ${sg.total_nodos} estados${arbol}${sg.truncado ? ' · truncada' : ''}`;

  $$('#explorador-grafo .sg-node').forEach((el) => el.addEventListener('click', () => {
    if (el.dataset.id === sg.centro) return;
    state.centro = el.dataset.id;
    cargarSubgrafo();
  }));
}

$('#sub-profundidad').addEventListener('input', (e) => {
  $('#sub-profundidad-val').textContent = e.target.value;
  if (!$('#sub-controls').hidden) cargarSubgrafo();
});

// ------------------------------------------------------ subgraph rendering

const SG_RADIO = 20;
const SG_ANILLO = 115; // distance between BFS rings
const SG_MARGEN = 46;

function numeroEstado(id) {
  const m = /^q(\d+)$/.exec(String(id ?? ''));
  return m ? Number(m[1]) : 0;
}

// renderSubgrafo draws the BFS ball as a radial tree: the centre at the origin,
// one ring per distance, and each subtree inside its own angular sector so the
// edges never cross. A layered grid was tried first and had to be scaled down to
// fit the card, which made the labels unreadable.
function renderSubgrafo(sg) {
  const nodos = sg.nodos.slice().sort((a, b) => numeroEstado(a.id) - numeroEstado(b.id));
  const porId = new Map(nodos.map((n) => [n.id, n]));

  // One parent edge per node: the edge that comes from the previous ring.
  const padre = new Map();
  for (const a of sg.aristas || []) {
    if (a.hacia === sg.centro || a.desde === a.hacia) continue;
    const n = porId.get(a.hacia);
    const p = porId.get(a.desde);
    if (!n || !p || p.distancia !== n.distancia - 1) continue;
    if (!padre.has(a.hacia)) padre.set(a.hacia, a);
  }

  const hijos = new Map(nodos.map((n) => [n.id, []]));
  for (const [id, a] of padre) {
    if (hijos.has(a.desde)) hijos.get(a.desde).push(id);
  }
  for (const lista of hijos.values()) {
    lista.sort((x, y) => numeroEstado(x) - numeroEstado(y));
  }

  // Subtree sizes decide how much of the parent's sector each child receives.
  const tam = new Map();
  const medir = (id) => {
    let s = 1;
    for (const c of hijos.get(id) || []) s += medir(c);
    tam.set(id, s);
    return s;
  };
  medir(sg.centro);

  const angulo = new Map();
  const repartir = (id, desde, hasta) => {
    angulo.set(id, (desde + hasta) / 2);
    const cs = hijos.get(id) || [];
    if (!cs.length) return;
    const total = cs.reduce((s, c) => s + tam.get(c), 0);
    let a = desde;
    for (const c of cs) {
      const ancho = ((hasta - desde) * tam.get(c)) / total;
      repartir(c, a, a + ancho);
      a += ancho;
    }
  };
  repartir(sg.centro, -Math.PI / 2, (3 * Math.PI) / 2);

  const pos = new Map();
  let radioMax = 0;
  for (const n of nodos) {
    const a = angulo.get(n.id);
    if (a === undefined) continue; // unreachable from the centre, should not happen
    const r = n.distancia * SG_ANILLO;
    radioMax = Math.max(radioMax, r);
    pos.set(n.id, { x: r * Math.cos(a), y: r * Math.sin(a), a, r });
  }

  const lado = 2 * (radioMax + SG_MARGEN);
  // A one-ring ball would sit tiny inside the card, so it is rendered at least
  // this wide; the viewBox stays tight and the SVG scales up to fill it.
  const mostrar = Math.max(lado, 560);
  const partes = [];
  partes.push(`<svg class="sg" viewBox="${-lado / 2} ${-lado / 2} ${lado} ${lado}" width="${mostrar}" height="${mostrar}" role="img" aria-label="vecindad de ${esc(sg.centro)}">`);
  partes.push(`<defs><marker id="sg-arrow" viewBox="0 0 10 10" refX="9" refY="5" markerWidth="7" markerHeight="7" orient="auto-start-reverse"><path d="M0,0 L10,5 L0,10 z" class="sg-arrow-head"/></marker></defs>`);

  const inicial = nodos.find((n) => n.inicial);
  const pi = inicial && pos.get(inicial.id);
  if (pi && pi.r === 0) {
    // The centre has no ring to come from, so the start marker points at it
    // horizontally, the usual way to mark an initial state.
    partes.push(`<line class="sg-inicio" x1="${-SG_RADIO - 34}" y1="0" x2="${-SG_RADIO - 6}" y2="0" marker-end="url(#sg-arrow)"/>`);
  } else if (pi) {
    const desde = pi.r + SG_MARGEN - 12;
    const hasta = pi.r + SG_RADIO + 5;
    partes.push(`<line class="sg-inicio" x1="${desde * Math.cos(pi.a)}" y1="${desde * Math.sin(pi.a)}" x2="${hasta * Math.cos(pi.a)}" y2="${hasta * Math.sin(pi.a)}" marker-end="url(#sg-arrow)"/>`);
  }

  // Parallel edges share one line; their symbols stack across it. This keeps
  // "q0 reaches q1 with L, l, N and S" visible instead of picking one label.
  const pares = new Map();
  for (const a of sg.aristas || []) {
    if (a.desde === a.hacia) continue;
    const p = pos.get(a.desde);
    const q = pos.get(a.hacia);
    if (!p || !q) continue;
    const clave = `${a.desde}->${a.hacia}`;
    if (!pares.has(clave)) pares.set(clave, { p, q, simbolos: [] });
    pares.get(clave).simbolos.push(a.simbolo);
  }

  for (const { p, q, simbolos } of pares.values()) {
    const dx = q.x - p.x;
    const dy = q.y - p.y;
    const largo = Math.hypot(dx, dy) || 1;
    const ux = dx / largo;
    const uy = dy / largo;
    const x1 = p.x + ux * (SG_RADIO + 2);
    const y1 = p.y + uy * (SG_RADIO + 2);
    const x2 = q.x - ux * (SG_RADIO + 6);
    const y2 = q.y - uy * (SG_RADIO + 6);
    partes.push(`<line class="sg-edge" x1="${x1}" y1="${y1}" x2="${x2}" y2="${y2}" marker-end="url(#sg-arrow)"/>`);
    const mx = x1 + (x2 - x1) * 0.55;
    const my = y1 + (y2 - y1) * 0.55;
    const nx = -uy;
    const ny = ux;
    simbolos.forEach((simbolo, i) => {
      const corrimiento = ((simbolos.length - 1) / 2 - i) * 15;
      partes.push(etiquetaSimbolo(mx + nx * corrimiento, my + ny * corrimiento, simbolo));
    });
  }

  // Self loops sit outside the ring, away from the centre.
  for (const a of sg.aristas || []) {
    if (a.desde !== a.hacia) continue;
    const p = pos.get(a.desde);
    if (!p) continue;
    const ux = Math.cos(p.a);
    const uy = Math.sin(p.a);
    const cx = p.x + ux * (SG_RADIO + 20);
    const cy = p.y + uy * (SG_RADIO + 20);
    const nx = -uy;
    const ny = ux;
    partes.push(`<path class="sg-edge" d="M ${p.x + nx * 10} ${p.y + ny * 10} Q ${cx + nx * 14} ${cy + ny * 14}, ${p.x - nx * 10} ${p.y - ny * 10}" marker-end="url(#sg-arrow)"/>`);
    partes.push(etiquetaSimbolo(cx, cy, a.simbolo));
  }

  for (const n of nodos) {
    const p = pos.get(n.id);
    if (!p) continue;
    const clases = `sg-node${n.aceptor ? ' aceptor' : ''}${n.id === sg.centro ? ' centro' : ''}`;
    partes.push(`<g class="${clases}" data-id="${esc(n.id)}" transform="translate(${p.x} ${p.y})">`);
    partes.push(`<circle r="${SG_RADIO}"/>`);
    if (n.aceptor) partes.push(`<circle r="${SG_RADIO - 4}" class="interno"/>`);
    const tam = n.id.length > 3 ? 10 : 13;
    partes.push(`<text style="font-size:${tam}px">${esc(n.id)}</text>`);
    partes.push('</g>');
  }

  partes.push('</svg>');
  return partes.join('');
}

// etiquetaSimbolo keeps the edge label legible over the line with a white rect.
function etiquetaSimbolo(x, y, simbolo) {
  const texto = String(simbolo ?? '');
  const ancho = texto.length * 7 + 8;
  return `<g class="sg-etiqueta">
    <rect x="${x - ancho / 2}" y="${y - 9}" width="${ancho}" height="18" rx="4"/>
    <text x="${x}" y="${y}">${esc(texto)}</text>
  </g>`;
}

function renderTransiciones(datos) {
  const filas = datos.transiciones.map((t) =>
    `<tr><td class="mono">${esc(t.desde)}</td><td class="mono">${esc(t.simbolo)}</td><td class="mono">${esc(t.hacia)}</td></tr>`
  ).join('');
  $('#explorador-tabla').innerHTML = `
    <table><thead><tr><th>Desde</th><th>Símbolo</th><th>Hacia</th></tr></thead>
    <tbody>${filas || '<tr><td colspan="3" class="empty">Sin transiciones.</td></tr>'}</tbody></table>`;

  const paginas = Math.max(1, Math.ceil(datos.total / datos.limite));
  $('#explorador-pager').innerHTML = `
    <button class="ghost small" id="pag-prev" ${datos.pagina <= 1 ? 'disabled' : ''}>←</button>
    <span>página ${datos.pagina} de ${paginas} · ${datos.total} transiciones</span>
    <button class="ghost small" id="pag-next" ${datos.pagina >= paginas ? 'disabled' : ''}>→</button>`;
  $('#pag-prev')?.addEventListener('click', () => { state.pagina--; cargarVista(); });
  $('#pag-next')?.addEventListener('click', () => { state.pagina++; cargarVista(); });
}

$('#simular').addEventListener('click', async () => {
  const r = state.regla;
  const out = $('#sim-traza');
  if (!r) { out.innerHTML = '<span class="empty">Elegí una regla primero.</span>'; return; }
  if (!r.tiene_er) { out.innerHTML = '<span class="empty">Esta regla no tiene autómata.</span>'; return; }
  const valor = $('#sim-valor').value;
  try {
    const res = await postJSON('/api/simular', { id: r.id, valor, vista: state.vista });
    const marca = res.acepta ? '<span class="mark-ok">✔ acepta</span>' : '<span class="mark-bad">✘ rechaza</span>';
    const cadena = res.traza.map((p) =>
      `<span class="st">${esc(p.estado)}</span>─<span class="cls">${esc(p.simbolo)}</span>→`
    ).join(' ');
    out.innerHTML = `${cadena} ${marca}${res.fallo != null ? ` <span class="meta">(se corta en la posición ${res.fallo + 1})</span>` : ''}`;
  } catch (err) {
    out.innerHTML = errorBox(err);
  }
});

// ------------------------------------------------------------------- reglas

$$('#reglas-filtro button').forEach((b) => b.addEventListener('click', () => {
  state.filtro = b.dataset.filtro;
  $$('#reglas-filtro button').forEach((x) => x.setAttribute('aria-pressed', String(x.dataset.filtro === state.filtro)));
  renderReglas();
}));

function renderReglas() {
  const cont = $('#reglas-tabla');
  if (!state.reglas.length) { cont.innerHTML = '<div class="empty">Cargando…</div>'; return; }

  const filas = state.reglas
    .filter((r) => {
      if (state.filtro === 'todas') return true;
      if (state.filtro === 'password') return r.id[0] === 'P';
      return r.id[0] === 'C';
    })
    .map((r) => {
      const c = r.conteos
        ? `${r.conteos.nfa} / ${r.conteos.sub} / ${r.conteos.tree} / ${r.conteos.min}`
        : '—';
      return `<tr>
        <td class="mono">${esc(r.id)}</td>
        <td>${esc(r.nombre)}</td>
        <td>${r.tipo === 'accept' ? 'aceptación' : 'rechazo'}</td>
        <td class="er">${r.er ? esc(r.er) : '<span class="empty">conjunto</span>'}</td>
        <td class="er">${r.re2 ? esc(r.re2) : '—'}</td>
        <td class="num">${c}</td>
        <td><button class="ghost small" data-open="${esc(r.id)}">ver</button></td>
      </tr>`;
    }).join('');

  cont.innerHTML = `<table>
    <thead><tr><th>Id</th><th>Regla</th><th>Tipo</th><th>ER formal</th><th>Traducción RE2</th>
      <th class="num">AFND / sub / árbol / mín</th><th></th></tr></thead>
    <tbody>${filas}</tbody></table>`;

  $$('[data-open]').forEach((b) => b.addEventListener('click', () => {
    showScreen('explorador');
    selectRegla(b.dataset.open);
  }));
}

// -------------------------------------------------------------------- pares

async function loadPares() {
  if (state.pares.length) { renderPares(); return; }
  const cont = $('#pares-tabla');
  try {
    state.pares = await api('/api/pares');
    renderPares();
  } catch (err) {
    cont.innerHTML = errorBox(err);
  }
}

function renderPares() {
  const filas = state.pares.map((p) => {
    const marcaEmail = p.email_esperado
      ? '<span class="mark-ok">válido</span>' : '<span class="mark-bad">inválido</span>';
    const marcaPass = p.password_esperado
      ? '<span class="mark-ok">válido</span>' : '<span class="mark-bad">inválido</span>';
    const semillas = [p.seed_email != null ? `e:${p.seed_email}` : '', p.seed_password != null ? `p:${p.seed_password}` : '']
      .filter(Boolean).join(' · ');
    return `<tr>
      <td class="num">${p.id}</td>
      <td class="mono">${esc(p.email)}</td>
      <td>${marcaEmail}</td>
      <td class="mono">${esc(p.password)}</td>
      <td>${marcaPass}</td>
      <td class="mono">${esc(p.regla_fallida || '—')}</td>
      <td class="mono">${esc(semillas || '—')}</td>
      <td><button class="ghost small" data-pair="${p.id}">probar</button></td>
    </tr>`;
  }).join('');

  $('#pares-tabla').innerHTML = `<table>
    <thead><tr><th>#</th><th>Correo</th><th></th><th>Contraseña</th><th></th>
      <th>Regla fallida</th><th>Semillas</th><th></th></tr></thead>
    <tbody>${filas}</tbody></table>`;

  $$('[data-pair]').forEach((b) => b.addEventListener('click', () => {
    const p = state.pares.find((x) => String(x.id) === b.dataset.pair);
    if (!p) return;
    showScreen('validar');
    setTipo(p.password_esperado ? 'password' : 'email');
    $('#valor').value = p.password_esperado ? p.password : p.email;
    validar();
  }));
}

// --------------------------------------------------------------------- init

setTipo('password');
loadReglas().catch(() => {});
