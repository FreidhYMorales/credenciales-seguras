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
    out.innerHTML = veredicto + renderReporte(repEmail, email) + renderReporte(repPass, pass);
  } catch (err) {
    out.innerHTML = `<div class="card">${errorBox(err)}</div>`;
  }
}

function renderReporte(rep, original) {
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
  $('#sub-info').textContent =
    `vecindad de ${sg.centro} · ${sg.nodos.length} de ${sg.total_nodos} estados${sg.truncado ? ' · truncada' : ''}`;

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

const SG_RADIO = 18;
const SG_PASO_X = 96;
const SG_PASO_Y = 110;
const SG_MARGEN = 24;

function numeroEstado(id) {
  const m = /^q(\d+)$/.exec(String(id ?? ''));
  return m ? Number(m[1]) : 0;
}

// renderSubgrafo draws the BFS ball as an inline SVG: one row per distance,
// nodes sorted numerically, straight edges with a labelled arrowhead.
function renderSubgrafo(sg) {
  const nodos = sg.nodos.slice().sort((a, b) => numeroEstado(a.id) - numeroEstado(b.id));
  const aristas = (sg.aristas || []).slice().sort((a, b) => {
    const desde = numeroEstado(a.desde) - numeroEstado(b.desde);
    if (desde) return desde;
    if (a.simbolo !== b.simbolo) return a.simbolo < b.simbolo ? -1 : 1;
    return numeroEstado(a.hacia) - numeroEstado(b.hacia);
  });

  const filas = [];
  for (const n of nodos) {
    while (filas.length <= n.distancia) filas.push([]);
    filas[n.distancia].push(n);
  }

  // A self-loop arcs above its node, so the first row needs extra headroom.
  const primeraFila = new Set(filas[0].map((n) => n.id));
  const extraArriba = aristas.some((a) => a.desde === a.hacia && primeraFila.has(a.desde)) ? 40 : 0;

  const margen = SG_MARGEN + SG_RADIO;
  const anchoFila = Math.max(0, ...filas.map((f) => (f.length - 1) * SG_PASO_X));
  const ancho = anchoFila + 2 * margen;
  const alto = extraArriba + (filas.length - 1) * SG_PASO_Y + 2 * margen;

  const pos = new Map();
  filas.forEach((fila, d) => {
    const inicio = (ancho - (fila.length - 1) * SG_PASO_X) / 2;
    fila.forEach((n, i) => {
      pos.set(n.id, { x: inicio + i * SG_PASO_X, y: extraArriba + margen + d * SG_PASO_Y });
    });
  });

  const partes = [];
  partes.push(`<svg class="sg" viewBox="0 0 ${ancho} ${alto}" width="100%" role="img" aria-label="vecindad de ${esc(sg.centro)}">`);
  partes.push(`<defs><marker id="sg-arrow" viewBox="0 0 10 10" refX="9" refY="5" markerWidth="7" markerHeight="7" orient="auto-start-reverse"><path d="M0,0 L10,5 L0,10 z" class="sg-arrow-head"/></marker></defs>`);

  const inicial = nodos.find((n) => n.inicial);
  if (inicial) {
    const p = pos.get(inicial.id);
    const desde = Math.max(4, p.x - 50);
    partes.push(`<line class="sg-inicio" x1="${desde}" y1="${p.y}" x2="${p.x - SG_RADIO - 4}" y2="${p.y}" marker-end="url(#sg-arrow)"/>`);
  }

  for (const a of aristas) {
    const p = pos.get(a.desde);
    const q = pos.get(a.hacia);
    if (!p || !q) continue;
    if (a.desde === a.hacia) {
      const d = `M ${p.x - 10} ${p.y - 16} C ${p.x - 28} ${p.y - 48}, ${p.x + 28} ${p.y - 48}, ${p.x + 10} ${p.y - 16}`;
      partes.push(`<path class="sg-edge" d="${d}" marker-end="url(#sg-arrow)"/>`);
      partes.push(etiquetaSimbolo(p.x, p.y - 50, a.simbolo));
      continue;
    }
    const dx = q.x - p.x;
    const dy = q.y - p.y;
    const largo = Math.hypot(dx, dy) || 1;
    const ux = dx / largo;
    const uy = dy / largo;
    const x1 = p.x + ux * (SG_RADIO + 1);
    const y1 = p.y + uy * (SG_RADIO + 1);
    const x2 = q.x - ux * (SG_RADIO + 5);
    const y2 = q.y - uy * (SG_RADIO + 5);
    partes.push(`<line class="sg-edge" x1="${x1}" y1="${y1}" x2="${x2}" y2="${y2}" marker-end="url(#sg-arrow)"/>`);
    partes.push(etiquetaSimbolo((x1 + x2) / 2, (y1 + y2) / 2, a.simbolo));
  }

  for (const n of nodos) {
    const p = pos.get(n.id);
    const clases = `sg-node${n.aceptor ? ' aceptor' : ''}${n.id === sg.centro ? ' centro' : ''}`;
    partes.push(`<g class="${clases}" data-id="${esc(n.id)}" transform="translate(${p.x} ${p.y})">`);
    partes.push(`<circle r="${SG_RADIO}"/>`);
    if (n.aceptor) partes.push(`<circle r="${SG_RADIO - 4}" class="interno"/>`);
    const tam = n.id.length > 3 ? 9 : 12;
    partes.push(`<text style="font-size:${tam}px">${esc(n.id)}</text>`);
    partes.push(`</g>`);
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
