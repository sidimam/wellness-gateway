/* Wellness Gateway web UI: walkthrough di primo avvio, login, prenotazioni, lezioni, profili, impostazioni, registro. */
const $ = s => document.querySelector(s);
const main = $('#main'), nav = $('#nav');
let token = sessionStorage.getItem('wg.token') || '';
let me = null, profiles = [], currentProfile = '';
const fmtD = d => new Date(d).toLocaleString('it-IT', { weekday: 'short', day: 'numeric', month: 'short', hour: '2-digit', minute: '2-digit' });
const fmtT = d => new Date(d).toLocaleTimeString('it-IT', { hour: '2-digit', minute: '2-digit' });
const esc = s => String(s ?? '').replace(/[&<>"']/g, c => ({ '&': '&amp;', '<': '&lt;', '>': '&gt;', '"': '&quot;', "'": '&#39;' }[c]));
const STATE = { pending: 'In attesa', bursting: 'Prenotazione in corso', watching: 'Osservazione: piena', waitingList: "Lista d'attesa", booked: 'Prenotata', failed: 'Errore', expired: 'Scaduta', cancelled: 'Disdetta' };

async function api(path, opts = {}) {
  const r = await fetch('/api/v1' + path, { ...opts, headers: { 'Content-Type': 'application/json', ...(token ? { Authorization: 'Bearer ' + token } : {}), ...(opts.headers || {}) }, body: opts.body ? JSON.stringify(opts.body) : undefined });
  const data = await r.json().catch(() => ({}));
  if (r.status === 401 && path !== '/auth/login') { token = ''; sessionStorage.removeItem('wg.token'); render(); throw new Error('Sessione scaduta'); }
  if (!r.ok) throw new Error(data.error || ('HTTP ' + r.status));
  return data;
}
const msg = (el, text, ok) => { el.innerHTML = text ? `<div class="msg ${ok ? 'ok' : 'err'}">${esc(text)}</div>` : ''; };

async function render() {
  const st = await fetch('/api/v1/setup').then(r => r.json());
  $('#version').textContent = st.version;
  if (st.needsSetup) return walkthrough(st);
  if (!token) return loginView();
  try { me = await api('/me'); } catch { return loginView(); }
  nav.hidden = false;
  profiles = await api('/profiles');
  if (!currentProfile && profiles.length) currentProfile = profiles[0].id;
  showView(location.hash.slice(1) || 'dashboard');
}

/* ---------- walkthrough di primo avvio ---------- */
function walkthrough(st) {
  nav.hidden = true;
  let step = 0;
  const steps = [
    () => `<h1>Benvenuto in Wellness Gateway</h1>
      <p>Questo container prenota da solo le lezioni Technogym mywellness per tutta la famiglia: all'apertura delle prenotazioni e, se una classe è piena, appena si libera un posto (osservazione continua). L'app iPhone <b>Wellness Booking</b> serve per scegliere cosa prenotare e ricevere le notifiche.</p>
      <p class="mut">Quattro passi: amministratore → profili mywellness → notifiche push → app.</p>
      <button class="primary" id="next">Inizia</button>`,
    () => `<h1>1 · Amministratore</h1><p class="mut">L'utente con cui entrerai nella web UI e nell'app. Potrai aggiungere gli altri familiari dopo.</p>
      <label>Nome utente</label><input id="su" placeholder="simone" autocomplete="username">
      <label>Nome visualizzato</label><input id="sd" placeholder="Simone">
      <label>Password (almeno 6 caratteri)</label><input id="sp" type="password" autocomplete="new-password">
      <div id="m"></div><p><button class="primary" id="next">Crea amministratore</button></p>`,
    () => `<h1>2 · Profili mywellness</h1><p class="mut">Ogni profilo è un account Technogym mywellness (email e password con cui si prenota al centro). Le password restano cifrate sul NAS; il gateway fa il login e rinnova la sessione da solo. Puoi aggiungerne altri dopo da "Profili".</p>
      ${profileForm()}<div id="m"></div><div id="plist"></div>
      <p><button class="primary" id="addp">Aggiungi profilo</button> <button class="small" id="next">Avanti</button></p>`,
    () => `<h1>3 · Notifiche push</h1>
      <p>${st.push ? '✅ APNs configurato: l\'app riceverà notifiche prioritarie (prenotata, posto libero, disdetta).' : '⚠️ APNs non configurato. Per le notifiche sull\'iPhone imposta nel container le variabili <code>APNS_KEY_PATH</code> (file .p8 montato), <code>APNS_KEY_ID</code>, <code>APNS_TEAM_ID</code>. Senza, l\'app mostra comunque lo stato quando la apri.'}</p>
      <p class="mut">La chiave APNs si crea su developer.apple.com → Keys → "+" → Apple Push Notifications service.</p>
      <p><button class="primary" id="next">Avanti</button></p>`,
    () => `<h1>4 · Collega l'app</h1>
      <p>Su iPhone apri <b>Wellness Booking</b> → Altro → <b>Server</b> e inserisci:</p>
      <table><tr><th>Indirizzo</th><td><code>${esc(st.publicUrl || location.origin)}</code></td></tr><tr><th>Utente</th><td>quello creato al passo 1 (o un familiare)</td></tr></table>
      <p class="mut">Da casa puoi usare anche l'indirizzo LAN del NAS. Fuori casa serve il tunnel Cloudflare (es. booking.manieridimambro.it).</p>
      <p><button class="primary" id="next">Vai al pannello</button></p>`,
  ];
  const draw = () => {
    main.innerHTML = `<div class="card"><div class="steps">${steps.map((_, i) => `<span class="${i <= step ? 'on' : ''}"></span>`).join('')}</div>${steps[step]()}</div>`;
    $('#next')?.addEventListener('click', async () => {
      if (step === 1) {
        try {
          const r = await api('/setup', { method: 'POST', body: { username: $('#su').value, password: $('#sp').value, displayName: $('#sd').value } });
          token = r.token; sessionStorage.setItem('wg.token', token);
        } catch (e) { return msg($('#m'), e.message); }
      }
      if (step === steps.length - 1) { location.hash = 'dashboard'; return render(); }
      step++; draw();
    });
    $('#addp')?.addEventListener('click', () => addProfileFromForm($('#m'), async () => { profiles = await api('/profiles'); $('#plist').innerHTML = profiles.map(p => `<div class="badge booked">${esc(p.label)} · ${esc(p.facilityName)}</div> `).join(''); }));
    if (step === 2) api('/profiles').then(ps => { profiles = ps; $('#plist').innerHTML = ps.map(p => `<div class="badge booked">${esc(p.label)} · ${esc(p.facilityName)}</div> `).join(''); }).catch(() => {});
  };
  draw();
}

function profileForm() {
  return `<div class="grid">
    <div><label>Etichetta (es. Daniela)</label><input id="pl"></div>
    <div><label>Email mywellness</label><input id="pu" autocomplete="off"></div>
    <div><label>Password mywellness</label><input id="pp" type="password" autocomplete="new-password"></div>
    <div><label>Centro (URL widget)</label><input id="pf" value="wellnesstown"></div>
    <div><label>Massimo prenotazioni attive</label><input id="pm" type="number" value="5" min="1"></div>
    <div><label>Visibilità</label><select id="pv"><option value="0">Famiglia (tutti gli utenti)</option><option value="1">Solo io</option></select></div>
  </div>`;
}
async function addProfileFromForm(m, done) {
  msg(m, 'Verifica login mywellness…', true);
  try {
    await api('/profiles', { method: 'POST', body: { label: $('#pl').value, username: $('#pu').value, password: $('#pp').value, facilityUrl: $('#pf').value, maxBookings: +$('#pm').value, private: $('#pv').value === '1' } });
    msg(m, 'Profilo aggiunto e verificato.', true); $('#pp').value = ''; $('#pu').value = ''; $('#pl').value = '';
    await done();
  } catch (e) { msg(m, e.message); }
}

/* ---------- login ---------- */
function loginView() {
  nav.hidden = true;
  main.innerHTML = `<div class="card" style="max-width:420px;margin:40px auto"><h1>Accedi</h1>
    <label>Nome utente</label><input id="lu" autocomplete="username"><label>Password</label><input id="lp" type="password" autocomplete="current-password">
    <div id="m"></div><p><button class="primary" id="go">Entra</button></p></div>`;
  const go = async () => {
    try { const r = await api('/auth/login', { method: 'POST', body: { username: $('#lu').value, password: $('#lp').value, deviceName: 'Web UI' } }); token = r.token; sessionStorage.setItem('wg.token', token); render(); }
    catch (e) { msg($('#m'), e.message); }
  };
  $('#go').onclick = go; $('#lp').onkeydown = e => { if (e.key === 'Enter') go(); };
}
$('#logout').onclick = async () => { try { await api('/auth/logout', { method: 'POST' }); } catch {} token = ''; sessionStorage.removeItem('wg.token'); render(); };
nav.querySelectorAll('button[data-view]').forEach(b => b.onclick = () => { location.hash = b.dataset.view; showView(b.dataset.view); });

function showView(v) {
  nav.querySelectorAll('button[data-view]').forEach(b => b.classList.toggle('active', b.dataset.view === v));
  ({ dashboard, classes, profiles: profilesView, settings, log }[v] || dashboard)();
}
const profileSelect = () => profiles.length > 1 ? `<select id="psel" style="width:auto">${profiles.map(p => `<option value="${p.id}" ${p.id === currentProfile ? 'selected' : ''}>${esc(p.label)}</option>`).join('')}</select>` : `<b>${esc(profiles[0]?.label || '')}</b>`;
const bindProfileSelect = cb => { const s = $('#psel'); if (s) s.onchange = () => { currentProfile = s.value; cb(); }; };
const itemCard = (it, actions) => `<div class="item">${it.pictureUrl ? `<img src="${esc(it.pictureUrl)}" alt="">` : '<div style="width:56px"></div>'}<div class="body">
  <div class="title">${esc(it.name)} ${it.recurring ? '🔁' : ''} <span class="badge ${esc(it.state)}">${STATE[it.state] || it.state}</span></div>
  <div class="sub">${fmtD(it.start)} – ${fmtT(it.end)}${it.room ? ' · ' + esc(it.room) : ''}${it.trainer ? ' · ' + esc(it.trainer) : ''}</div>
  ${it.state === 'pending' && it.fireAt ? `<div class="sub">Prenoto ${fmtD(it.fireAt)}</div>` : ''}
  ${it.lastMessage ? `<div class="sub">${esc(it.lastMessage)}</div>` : ''}
  ${it.lastCheck && !['booked', 'expired', 'cancelled'].includes(it.state) ? `<div class="sub">Ultimo controllo ${fmtT(it.lastCheck)} · tentativi ${it.attempts}</div>` : ''}
  <div class="row" style="margin-top:6px">${actions}</div></div></div>`;

/* ---------- prenotazioni ---------- */
async function dashboard() {
  if (!profiles.length) { main.innerHTML = `<div class="card"><h1>Nessun profilo</h1><p>Aggiungi un profilo mywellness da <a href="#profiles" onclick="location.hash='profiles';showView('profiles')">Profili</a>.</p></div>`; return; }
  main.innerHTML = `<div class="card"><div class="row"><h1 style="margin:0">Prenotazioni</h1>${profileSelect()}<button class="small" id="rf">Aggiorna</button></div><div id="m"></div><div id="st" class="mut"></div></div>
    <div class="card"><h2>Lezioni seguite</h2><div id="items" class="grid"></div></div>
    <div class="card"><h2>Prenotate su mywellness</h2><p class="mut">Tutte le prenotazioni attive del profilo, fatte dal gateway o dall'app/sito Technogym. Da qui puoi disdire.</p><div id="bk" class="grid"></div></div>`;
  const load = async (refresh) => {
    const [items, bookings, status] = await Promise.all([api('/items?profile=' + currentProfile), api(`/profiles/${currentProfile}/bookings${refresh ? '?refresh=1' : ''}`), api('/status')]);
    const p = profiles.find(x => x.id === currentProfile) || {};
    $('#st').textContent = `Prenotazioni attive ${p.activeBookings ?? 0}/${p.maxBookings ?? 5} · motore attivo, prossimo controllo ${fmtT(status.nextWake)}`;
    $('#items').innerHTML = items.length ? items.map(it => itemCard(it, `
      ${it.state === 'failed' ? `<button class="small" data-retry="${it.id}">Riprova</button>` : ''}
      ${it.state === 'booked' ? `<button class="danger" data-unbook="${it.classId}|${it.partitionDate}">Disdici</button>` : ''}
      <button class="small" data-del="${it.id}">Rimuovi</button>${it.recurring ? `<button class="small" data-delrule="${it.id}">Stop ricorrenza</button>` : ''}`)).join('') : '<p class="mut">Nessuna lezione seguita: vai in Lezioni.</p>';
    $('#bk').innerHTML = bookings.length ? bookings.map(b => itemCard({ ...b, name: b.name, room: b.room, trainer: b.assignedTo, state: 'booked', lastMessage: b.tracked ? 'Prenotata dal gateway' : 'Prenotata da mywellness (app/web)' }, `<button class="danger" data-unbook="${b.id}|${b.partitionDate}">Disdici</button>`)).join('') : '<p class="mut">Nessuna prenotazione attiva.</p>';
    main.querySelectorAll('[data-del]').forEach(b => b.onclick = async () => { await api('/items/' + encodeURIComponent(b.dataset.del), { method: 'DELETE' }); load(); });
    main.querySelectorAll('[data-delrule]').forEach(b => b.onclick = async () => { await api('/items/' + encodeURIComponent(b.dataset.delrule) + '?rule=1', { method: 'DELETE' }); load(); });
    main.querySelectorAll('[data-retry]').forEach(b => b.onclick = async () => { await api('/items/' + encodeURIComponent(b.dataset.retry) + '/retry', { method: 'POST' }); load(); });
    main.querySelectorAll('[data-unbook]').forEach(b => b.onclick = async () => {
      if (!confirm('Disdire la prenotazione su mywellness?')) return;
      const [classId, pd] = b.dataset.unbook.split('|');
      try { await api(`/profiles/${currentProfile}/unbook`, { method: 'POST', body: { classId, partitionDate: +pd } }); msg($('#m'), 'Disdetta inviata.', true); load(true); } catch (e) { msg($('#m'), e.message); }
    });
  };
  bindProfileSelect(load); $('#rf').onclick = () => load(true); load();
}

/* ---------- lezioni ---------- */
async function classes() {
  if (!profiles.length) return dashboard();
  main.innerHTML = `<div class="card"><div class="row"><h1 style="margin:0">Lezioni</h1>${profileSelect()}<input id="q" placeholder="Cerca lezione, istruttore, sala" style="max-width:320px"><button class="small" id="rf">Aggiorna</button></div><div id="m"></div><div id="list"></div></div>`;
  let all = [];
  const draw = () => {
    const q = $('#q').value.toLowerCase();
    const list = all.filter(c => new Date(c.start) > new Date() && (!q || (c.name + ' ' + (c.assignedTo || '') + ' ' + (c.room || '')).toLowerCase().includes(q)));
    let html = '', day = '';
    for (const c of list) {
      const d = new Date(c.start).toLocaleDateString('it-IT', { weekday: 'long', day: 'numeric', month: 'long' });
      if (d !== day) { day = d; html += `<div class="day">${d}</div><div class="grid">`; }
      const bi = c.bookingInfo || {};
      const status = c.isParticipant ? '<span class="badge booked">Prenotata</span>' : c.isInWaitingList ? '<span class="badge waitingList">In lista d\'attesa</span>' : bi.bookingAvailable === false ? '<span class="badge">Non prenotabile online</span>' : c.opensOn && new Date(c.opensOn) > new Date() ? `<span class="badge pending">Apre ${fmtD(c.opensOn)}</span>` : c.availablePlaces <= 0 ? '<span class="badge watching">Piena · lista d\'attesa</span>' : `<span class="badge booked">${c.availablePlaces} liberi su ${c.maxParticipants}</span>`;
      const actions = c.tracked ? `<span class="badge ${esc(c.tracked.state)}">${STATE[c.tracked.state]}</span>` : `<button class="small" data-add="${c.id}|${c.partitionDate}">Prenota questa</button><button class="small" data-addr="${c.id}|${c.partitionDate}">Ogni settimana</button>`;
      html += `<div class="item">${c.pictureUrl ? `<img src="${esc(c.pictureUrl)}" alt="">` : '<div style="width:56px"></div>'}<div class="body"><div class="title">${fmtT(c.start)}–${fmtT(c.end)} · ${esc(c.name)}</div><div class="sub">${esc(c.assignedTo || '')}${c.room ? ' · ' + esc(c.room) : ''}</div><div class="row" style="margin-top:6px">${status} ${actions}</div></div></div>`;
      if (list[list.indexOf(c) + 1] && new Date(list[list.indexOf(c) + 1].start).toLocaleDateString('it-IT', { weekday: 'long', day: 'numeric', month: 'long' }) !== d) html += '</div>';
    }
    $('#list').innerHTML = html + '</div>' || '<p class="mut">Nessuna lezione.</p>';
    const add = (b, rec) => async () => { try { await api('/items', { method: 'POST', body: { profileId: currentProfile, classId: b.split('|')[0], partitionDate: +b.split('|')[1], recurring: rec } }); msg($('#m'), 'Aggiunta alle prenotazioni automatiche.', true); load(); } catch (e) { msg($('#m'), e.message); } };
    main.querySelectorAll('[data-add]').forEach(b => b.onclick = add(b.dataset.add, false));
    main.querySelectorAll('[data-addr]').forEach(b => b.onclick = add(b.dataset.addr, true));
  };
  const load = async (refresh) => { msg($('#m'), ''); try { all = await api(`/profiles/${currentProfile}/classes${refresh ? '?refresh=1' : ''}`); draw(); } catch (e) { msg($('#m'), e.message); } };
  $('#q').oninput = draw; $('#rf').onclick = () => load(true); bindProfileSelect(load); load();
}

/* ---------- profili ---------- */
async function profilesView() {
  main.innerHTML = `<div class="card"><h1>Profili mywellness</h1><div id="plist"></div></div>
    <div class="card"><h2>Aggiungi profilo</h2>${profileForm()}<div id="m"></div><p><button class="primary" id="addp">Aggiungi e verifica</button></p></div>
    ${me.user.isAdmin ? `<div class="card"><h2>Utenti del gateway</h2><div id="users"></div><div class="grid"><div><label>Nome utente</label><input id="uu"></div><div><label>Nome</label><input id="ud"></div><div><label>Password</label><input id="up" type="password"></div><div><label>Ruolo</label><select id="ua"><option value="0">Utente</option><option value="1">Amministratore</option></select></div></div><div id="um"></div><p><button class="primary" id="addu">Aggiungi utente</button></p></div>` : ''}`;
  const load = async () => {
    profiles = await api('/profiles');
    $('#plist').innerHTML = profiles.length ? `<table><tr><th>Etichetta</th><th>Account</th><th>Centro</th><th>Max</th><th>Attive</th><th>Login</th><th></th></tr>${profiles.map(p => `<tr><td>${esc(p.label)}</td><td>${esc(p.username)}<br><small class="mut">${esc(p.displayName)}</small></td><td>${esc(p.facilityName)}</td><td>${p.maxBookings}</td><td>${p.activeBookings}</td><td>${p.lastLoginError ? `<span class="lv-error">${esc(p.lastLoginError)}</span>` : p.lastLoginAt ? `<span class="lv-success">ok ${fmtD(p.lastLoginAt)}</span>` : '—'}</td><td class="row"><button class="small" data-relogin="${p.id}">Rifai login</button><button class="danger" data-delp="${p.id}">Rimuovi</button></td></tr>`).join('')}</table>` : '<p class="mut">Nessun profilo.</p>';
    main.querySelectorAll('[data-relogin]').forEach(b => b.onclick = async () => { try { await api(`/profiles/${b.dataset.relogin}/relogin`, { method: 'POST' }); load(); } catch (e) { msg($('#m'), e.message); } });
    main.querySelectorAll('[data-delp]').forEach(b => b.onclick = async () => { if (confirm('Rimuovere il profilo e le sue lezioni seguite?')) { await api('/profiles/' + b.dataset.delp, { method: 'DELETE' }); load(); } });
    if (me.user.isAdmin) {
      const users = await api('/users');
      $('#users').innerHTML = `<table>${users.map(u => `<tr><td>${esc(u.username)}</td><td>${esc(u.displayName)}</td><td>${u.isAdmin ? 'admin' : 'utente'}</td><td>${u.id !== me.user.id ? `<button class="danger" data-delu="${u.id}">Elimina</button>` : ''}</td></tr>`).join('')}</table>`;
      main.querySelectorAll('[data-delu]').forEach(b => b.onclick = async () => { if (confirm('Eliminare utente?')) { await api('/users/' + b.dataset.delu, { method: 'DELETE' }); load(); } });
      $('#addu').onclick = async () => { try { await api('/users', { method: 'POST', body: { username: $('#uu').value, displayName: $('#ud').value, password: $('#up').value, isAdmin: $('#ua').value === '1' } }); msg($('#um'), 'Utente creato.', true); load(); } catch (e) { msg($('#um'), e.message); } };
    }
  };
  $('#addp').onclick = () => addProfileFromForm($('#m'), load);
  load();
}

/* ---------- impostazioni ---------- */
async function settings() {
  const s = await api('/settings'); const st = await api('/status');
  const ro = !me.user.isAdmin;
  main.innerHTML = `<div class="card"><h1>Impostazioni</h1>
    <label><input type="checkbox" id="fs" ${s.followServerOpenTime ? 'checked' : ''} ${ro ? 'disabled' : ''} style="width:auto"> Segui l'orario di apertura comunicato dal centro</label>
    <h2 style="margin-top:14px">Regole di apertura</h2><p class="mut">Scrivi tu il testo da cercare nel nome della lezione (es. Reformer → 3 giorni); la prima regola che corrisponde decide giorni e ora, * vale per tutte le altre.</p>
    <div id="rules"></div>${ro ? '' : '<button class="small" id="addr">Aggiungi regola</button>'}
    <div class="grid" style="margin-top:14px">
      <div><label>Anticipo (ms)</label><input id="lead" type="number" value="${s.leadMilliseconds}" ${ro ? 'disabled' : ''}></div>
      <div><label>Insisti dopo l'apertura (s)</label><input id="burst" type="number" value="${s.burstSeconds}" ${ro ? 'disabled' : ''}></div>
      <div><label>Osservazione: controlla ogni (s)</label><input id="poll" type="number" value="${s.pollSeconds}" ${ro ? 'disabled' : ''}></div>
      <div><label>Giorni di calendario</label><input id="days" type="number" value="${s.daysAhead}" ${ro ? 'disabled' : ''}></div>
    </div>
    <label><input type="checkbox" id="prio" ${s.priorityNotifications ? 'checked' : ''} ${ro ? 'disabled' : ''} style="width:auto"> Notifiche prioritarie (Time Sensitive)</label>
    <div id="m"></div>${ro ? '<p class="mut">Solo l\'amministratore può modificare le impostazioni.</p>' : '<p><button class="primary" id="save">Salva</button></p>'}</div>
    <div class="card"><h2>Stato</h2><table><tr><th>Versione</th><td>${esc(st.version)}</td></tr><tr><th>Avviato</th><td>${fmtD(st.startedAt)}</td></tr><tr><th>Push APNs</th><td>${st.push ? 'attivo' : 'non configurato'}</td></tr><tr><th>Lezioni seguite</th><td>${st.activeItems} in corso su ${st.items}</td></tr><tr><th>Indirizzo pubblico</th><td>${esc(st.publicUrl || '—')}</td></tr></table>
    <p><button class="small" id="tn">Invia notifica di prova ai miei dispositivi</button> <span id="tnm" class="mut"></span></p></div>`;
  let rules = s.openRules.slice();
  const drawRules = () => {
    $('#rules').innerHTML = `<table><tr><th>Testo nel nome</th><th>Giorni prima</th><th>Ora</th><th></th></tr>${rules.map((r, i) => `<tr><td><input data-p="${i}" value="${esc(r.pattern)}" ${ro || r.pattern === '*' ? 'disabled' : ''} placeholder="es. Reformer"></td><td><input data-d="${i}" type="number" value="${r.daysBefore}" ${ro ? 'disabled' : ''} style="width:90px"></td><td><input data-t="${i}" type="time" value="${String(r.hour).padStart(2, '0')}:${String(r.minute).padStart(2, '0')}" ${ro ? 'disabled' : ''} style="width:120px"></td><td>${!ro && r.pattern !== '*' ? `<button class="small" data-x="${i}">✕</button>` : ''}</td></tr>`).join('')}</table>`;
    main.querySelectorAll('[data-p]').forEach(i => i.oninput = () => rules[+i.dataset.p].pattern = i.value);
    main.querySelectorAll('[data-d]').forEach(i => i.oninput = () => rules[+i.dataset.d].daysBefore = +i.value);
    main.querySelectorAll('[data-t]').forEach(i => i.oninput = () => { const [h, m] = i.value.split(':'); rules[+i.dataset.t].hour = +h; rules[+i.dataset.t].minute = +m; });
    main.querySelectorAll('[data-x]').forEach(b => b.onclick = () => { rules.splice(+b.dataset.x, 1); drawRules(); });
  };
  drawRules();
  $('#addr') && ($('#addr').onclick = () => { rules.splice(Math.max(0, rules.length - 1), 0, { pattern: '', daysBefore: 3, hour: 5, minute: 0 }); drawRules(); });
  $('#save') && ($('#save').onclick = async () => {
    try { await api('/settings', { method: 'PUT', body: { followServerOpenTime: $('#fs').checked, openRules: rules, leadMilliseconds: +$('#lead').value, burstSeconds: +$('#burst').value, pollSeconds: +$('#poll').value, daysAhead: +$('#days').value, priorityNotifications: $('#prio').checked } }); msg($('#m'), 'Impostazioni salvate.', true); } catch (e) { msg($('#m'), e.message); }
  });
  $('#tn').onclick = async () => { const r = await api('/devices/test-notification', { method: 'POST' }); $('#tnm').textContent = r.push ? `inviata a ${r.sent} dispositivi` : 'APNs non configurato'; };
}

/* ---------- registro ---------- */
async function log() {
  const lines = await api('/log?limit=300');
  const names = Object.fromEntries(profiles.map(p => [p.id, p.label]));
  main.innerHTML = `<div class="card"><h1>Registro attività</h1><div class="log">${lines.map(l => `<div><span class="mut">${fmtD(l.time)}</span> <span class="lv-${l.level}">●</span> ${l.profileId ? `<b>${esc(names[l.profileId] || '')}</b> ` : ''}${esc(l.text)}</div>`).join('') || '<p class="mut">Nessuna attività.</p>'}</div></div>`;
}

render();
