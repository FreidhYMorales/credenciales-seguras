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
  $('#valor').placeholder = tipo === 'password' ? 'Escribí una contraseña' : 'Escribí un correo';
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
    out.innerHTML = renderReporte(rep);
  } catch (err) {
    out.innerHTML = `<div class="card">${errorBox(err)}</div>`;
  }
}

function renderReporte(rep) {
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

  const normalizado = rep.normalizado !== $('#valor').value
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
  renderSidebar();
  cargarVista();
}

$$('#explorador-vistas button').forEach((b) => b.addEventListener('click', () => {
  state.vista = b.dataset.vista;
  state.pagina = 1;
  $$('#explorador-vistas button').forEach((x) => x.setAttribute('aria-pressed', String(x.dataset.vista === state.vista)));
  cargarVista();
}));

async function cargarVista() {
  const r = state.regla;
  if (!r) return;
  $('#explorador-titulo').textContent = `${r.id} · ${r.nombre}`;

  if (!r.tiene_er) {
    $('#explorador-conteos').innerHTML = '';
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
    const datos = await api(`/api/automatas/${encodeURIComponent(r.id)}?vista=${state.vista}&pagina=${state.pagina}&limite=50`);
    if (datos.svg) {
      $('#explorador-grafo').innerHTML = `<div class="graph">${datos.svg}</div>`;
    } else {
      $('#explorador-grafo').innerHTML = `<div class="note">${esc(datos.motivo || 'Esta vista no se dibuja.')}</div>`;
    }
    renderTransiciones(datos);
  } catch (err) {
    $('#explorador-grafo').innerHTML = errorBox(err);
  }
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
