/* Wellness Gateway web UI: walkthrough di primo avvio, login, prenotazioni, lezioni, profili, impostazioni, registro. */
const $ = s => document.querySelector(s);
const main = $('#main'), nav = $('#nav');
let token = sessionStorage.getItem('wg.token') || '';
let me = null, profiles = [], currentProfile = '';
const fmtD = d => new Date(d).toLocaleString('it-IT', { weekday: 'short', day: 'numeric', month: 'short', hour: '2-digit', minute: '2-digit' });
const fmtT = d => new Date(d).toLocaleTimeString('it-IT', { hour: '2-digit', minute: '2-digit' });
const esc = s => String(s ?? '').replace(/[&<>"']/g, c => ({ '&': '&amp;', '<': '&lt;', '>': '&gt;', '"': '&quot;', "'": '&#39;' }[c]));
const avatar = (pr, name, size = 32) => pr && (pr.thumbUrl || pr.pictureUrl) ? `<img class="avatar" src="${esc(pr.thumbUrl || pr.pictureUrl)}" alt="" style="width:${size}px;height:${size}px">` : `<span class="avatar" style="width:${size}px;height:${size}px;font-size:${Math.round(size / 2.4)}px">${esc((name || '?').trim().slice(0, 1).toUpperCase())}</span>`;
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
  $('#version').textContent = `${st.version} · ${me.user.displayName}`;
  profiles = await api('/profiles');
  if (!profiles.some(p => p.id === currentProfile)) currentProfile = '';
  if (!currentProfile && profiles.length) currentProfile = (me.myProfileId && profiles.some(p => p.id === me.myProfileId)) ? me.myProfileId : profiles[0].id;
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
      <h2 style="margin-top:16px">Il tuo account mywellness <small class="mut">(opzionale, puoi farlo dopo)</small></h2>
      <p class="mut">Se anche tu prenoti al centro, inserisci qui le tue credenziali Technogym mywellness: diventano il tuo profilo personale.</p>
      <div class="grid"><div><label>Email mywellness</label><input id="smu" autocomplete="off"></div><div><label>Password mywellness</label><input id="smp" type="password" autocomplete="new-password"></div><div><label>Centro (URL widget)</label><input id="smf" value="wellnesstown"></div></div>
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
          msg($('#m'), 'Creazione in corso…', true);
          const mw = $('#smu').value ? { username: $('#smu').value, password: $('#smp').value, facilityUrl: $('#smf').value, maxBookings: 5 } : undefined;
          const r = await api('/setup', { method: 'POST', body: { username: $('#su').value, password: $('#sp').value, displayName: $('#sd').value, mywellness: mw } });
          token = r.token; sessionStorage.setItem('wg.token', token);
          if (r.profileError) { alert('Amministratore creato, ma il profilo mywellness no: ' + r.profileError + '. Potrai aggiungerlo al passo 2.'); }
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
    <div><label>Di chi è</label><select id="pw"><option value="1">È il mio account mywellness</option><option value="0">Di un familiare senza utente</option></select></div>
  </div>`;
}
async function addProfileFromForm(m, done) {
  msg(m, 'Verifica login mywellness…', true);
  try {
    await api('/profiles', { method: 'POST', body: { label: $('#pl').value, username: $('#pu').value, password: $('#pp').value, facilityUrl: $('#pf').value, maxBookings: +$('#pm').value, private: $('#pv').value === '1', mine: $('#pw').value === '1' } });
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

let viewId = 0, autoTimer = null;
function showView(v) {
  viewId++;
  if (autoTimer) { clearInterval(autoTimer); autoTimer = null; }
  nav.querySelectorAll('button[data-view]').forEach(b => b.classList.toggle('active', b.dataset.view === v));
  ({ dashboard, classes, profiles: profilesView, settings, log }[v] || dashboard)();
}
const stillHere = id => id === viewId;
/* autorefresh: ogni 20 s e quando la scheda torna visibile */
function autoRefresh(myView, fn) {
  if (autoTimer) clearInterval(autoTimer);
  autoTimer = setInterval(() => { if (stillHere(myView) && document.visibilityState === 'visible') fn(); }, 20000);
}
document.addEventListener('visibilitychange', () => { if (document.visibilityState === 'visible' && typeof window.__refreshNow === 'function') window.__refreshNow(); });
const profileSelect = () => `${avatar(profiles.find(p => p.id === currentProfile), '', 28)}<select id="psel" style="width:auto" title="Profilo mywellness">${profiles.map(p => `<option value="${p.id}" ${p.id === currentProfile ? 'selected' : ''}>${esc(p.label)}${p.userId === me?.user?.id ? ' (io)' : ''}</option>`).join('')}</select>`;
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
  main.innerHTML = `<div class="card"><div class="row"><h1 style="margin:0">Prenotazioni</h1>${profileSelect()}<button class="small" id="rf" title="Si aggiorna da solo ogni 20 secondi">Aggiorna</button></div><div id="m"></div><div id="st" class="mut"></div></div>
    <div class="card"><h2>Lezioni seguite</h2><div id="items" class="grid"></div></div>
    <div class="card"><h2>Prenotate su mywellness</h2><p class="mut">Tutte le prenotazioni attive del profilo, fatte dal gateway o dall'app/sito Technogym. Da qui puoi disdire.</p><div id="bk" class="grid"></div></div>
    <div class="card"><h2>Ultime attività</h2><div id="mini" class="log"></div><p><a href="#log" id="golog">Registro completo →</a></p></div>`;
  const myView = viewId;
  const load = async (refresh) => {
    let items, bookings, status, lines;
    try { [items, bookings, status, lines] = await Promise.all([api('/items?profile=' + currentProfile), api(`/profiles/${currentProfile}/bookings${refresh ? '?refresh=1' : ''}`), api('/status'), api('/log?limit=10&profile=' + currentProfile)]); }
    catch (e) { if (stillHere(myView)) msg($('#m'), e.message); return; }
    if (!stillHere(myView)) return;
    $('#mini').innerHTML = lines.map(l => `<div><span class="mut">${fmtD(l.time)}</span> <span class="lv-${l.level}">●</span> ${esc(l.text)}</div>`).join('') || '<p class="mut">Nessuna attività per questo profilo.</p>';
    $('#golog').onclick = e => { e.preventDefault(); location.hash = 'log'; showView('log'); };
    const p = profiles.find(x => x.id === currentProfile) || {};
    $('#st').textContent = `Prenotazioni attive ${p.activeBookings ?? 0}/${p.maxBookings ?? 5} · motore attivo, prossimo controllo ${fmtT(status.nextWake)}`;
    $('#items').innerHTML = items.length ? items.map(it => itemCard(it, `
      ${it.state === 'failed' ? `<button class="small" data-retry="${it.id}">Riprova</button>` : ''}
      ${it.state === 'booked' ? `<button class="danger" data-unbook="${it.classId}|${it.partitionDate}">Disdici</button>` : ''}
      ${it.state === 'waitingList' ? `<button class="danger" data-leave="${it.classId}|${it.partitionDate}">Esci dalla lista d'attesa</button>` : ''}
      <button class="small" data-del="${it.id}">Rimuovi</button>${it.recurring ? `<button class="small" data-delrule="${it.id}">Stop ricorrenza</button>` : ''}`)).join('') : '<p class="mut">Nessuna lezione seguita: vai in Lezioni.</p>';
    $('#bk').innerHTML = bookings.length ? bookings.map(b => itemCard({ ...b, name: b.name, room: b.room, trainer: b.assignedTo, state: 'booked', lastMessage: b.tracked ? 'Prenotata dal gateway' : 'Prenotata da mywellness (app/web)' }, `<button class="danger" data-unbook="${b.id}|${b.partitionDate}">Disdici</button>`)).join('') : '<p class="mut">Nessuna prenotazione attiva.</p>';
    main.querySelectorAll('[data-del]').forEach(b => b.onclick = async () => { await api('/items/' + encodeURIComponent(b.dataset.del), { method: 'DELETE' }); load(); });
    main.querySelectorAll('[data-delrule]').forEach(b => b.onclick = async () => { await api('/items/' + encodeURIComponent(b.dataset.delrule) + '?rule=1', { method: 'DELETE' }); load(); });
    main.querySelectorAll('[data-retry]').forEach(b => b.onclick = async () => { await api('/items/' + encodeURIComponent(b.dataset.retry) + '/retry', { method: 'POST' }); load(); });
    main.querySelectorAll('[data-leave]').forEach(b => b.onclick = async () => {
      if (!confirm(i18n.t("Uscire dalla lista d'attesa su mywellness? La lezione non verrà più seguita."))) return;
      const [classId, pd] = b.dataset.leave.split('|');
      try { await api(`/profiles/${currentProfile}/leave-waiting-list`, { method: 'POST', body: { classId, partitionDate: +pd, removeItem: true } }); msg($('#m'), i18n.t("Uscita dalla lista d'attesa."), true); load(true); } catch (e) { msg($('#m'), e.message); }
    });
    main.querySelectorAll('[data-unbook]').forEach(b => b.onclick = async () => {
      if (!confirm(i18n.t('Disdire la prenotazione su mywellness?'))) return;
      const [classId, pd] = b.dataset.unbook.split('|');
      try { await api(`/profiles/${currentProfile}/unbook`, { method: 'POST', body: { classId, partitionDate: +pd } }); msg($('#m'), 'Disdetta inviata.', true); load(true); } catch (e) { msg($('#m'), e.message); }
    });
  };
  bindProfileSelect(load); $('#rf').onclick = () => load(true); load();
  autoRefresh(myView, () => load(false)); window.__refreshNow = () => stillHere(myView) && load(false);
}

/* ---------- lezioni ---------- */
async function classes() {
  if (!profiles.length) return dashboard();
  main.innerHTML = `<div class="card"><div class="row"><h1 style="margin:0">Lezioni</h1>${profileSelect()}<input id="q" placeholder="Cerca lezione, istruttore, sala" style="max-width:320px"><button class="small" id="rf">Aggiorna</button></div><div id="m"></div><div id="list"></div></div>`;
  let all = [];
  const myView = viewId;
  const draw = () => {
    if (!stillHere(myView) || !$('#q')) return;
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
  const load = async (refresh) => { msg($('#m'), ''); try { all = await api(`/profiles/${currentProfile}/classes${refresh ? '?refresh=1' : ''}`); draw(); } catch (e) { if (stillHere(myView) && $('#m')) msg($('#m'), e.message); } };
  $('#q').oninput = draw; $('#rf').onclick = () => load(true); bindProfileSelect(load); load();
  autoRefresh(myView, () => load(false)); window.__refreshNow = () => stillHere(myView) && load(false);
}

/* ---------- utenti e account mywellness ---------- */
async function profilesView() {
  const admin = me.user.isAdmin;
  main.innerHTML = `<div class="card"><div class="row"><h1 style="margin:0">Utenti e account mywellness</h1>${admin ? '<button class="primary" id="newu">+ Nuovo utente</button>' : ''}</div>
    <p class="mut">Ogni utente entra nell'app con nome utente e password del gateway. Un utente normale ha il suo account Technogym mywellness e vede solo quello; un amministratore vede e gestisce tutti i profili e può anche essere solo locale, senza account mywellness. Tocca un utente per modificare tutti i suoi parametri.</p>
    <div id="m"></div><div id="ulist"></div></div>
    <div class="card"><h2>Account mywellness senza utente</h2><p class="mut">Account rimasti senza persona (per esempio dopo l'eliminazione di un utente): assegnali a un utente dalla sua finestra, oppure rimuovili.</p><div id="orph"></div></div>
    <dialog id="udlg"></dialog>`;
  let users = [];
  const load = async () => {
    profiles = await api('/profiles');
    users = admin ? await api('/users') : [me.user];
    $('#ulist').innerHTML = `<table><tr><th>Nome utente</th><th>Nome</th><th>Ruolo</th><th>Account mywellness</th><th>Centro</th><th>Attive</th><th>Login mywellness</th></tr>${users.map(u => { const pr = profiles.find(p => p.userId === u.id); return `<tr class="urow" data-u="${u.id}" style="cursor:pointer"><td><div class="row" style="gap:8px">${avatar(pr, u.displayName)}<b>${esc(u.username)}</b></div></td><td>${esc(u.displayName)}${pr && pr.displayName ? `<br><small class="mut">${esc(pr.displayName)}${pr.nickName ? ' · ' + esc(pr.nickName) : ''}</small>` : ''}</td><td>${u.isAdmin ? '<span class="badge bursting">admin</span>' : '<span class="badge">utente</span>'}</td>
      <td>${pr ? `${esc(pr.username)}${pr.ownerUserIds && pr.ownerUserIds.length ? ' 🔒' : ''}` : '<span class="badge failed">nessuno</span>'}</td><td>${pr ? esc(pr.facilityName) : ''}</td><td>${pr ? `${pr.activeBookings}/${pr.maxBookings}` : ''}</td>
      <td>${pr ? (pr.lastLoginError ? `<span class="lv-error">${esc(pr.lastLoginError)}</span>` : pr.lastLoginAt ? `<span class="lv-success">ok ${fmtD(pr.lastLoginAt)}</span>` : '—') : ''}</td></tr>`; }).join('')}</table>`;
    main.querySelectorAll('.urow').forEach(r => r.onclick = () => openUser(users.find(u => u.id === r.dataset.u)));
    const orph = profiles.filter(p => !p.userId);
    $('#orph').innerHTML = orph.length ? `<table>${orph.map(p => `<tr><td>${esc(p.label)}</td><td>${esc(p.username)}</td><td>${esc(p.facilityName)}</td><td class="row"><button class="small" data-relogin="${p.id}">Rifai login</button><button class="danger" data-delp="${p.id}">Rimuovi</button></td></tr>`).join('')}</table>` : '<p class="mut">Nessuno.</p>';
    main.querySelectorAll('[data-relogin]').forEach(b => b.onclick = async () => { try { await api(`/profiles/${b.dataset.relogin}/relogin`, { method: 'POST' }); load(); } catch (e) { msg($('#m'), e.message); } });
    main.querySelectorAll('[data-delp]').forEach(b => b.onclick = async () => { if (confirm(i18n.t('Rimuovere il profilo e le sue lezioni seguite?'))) { await api('/profiles/' + b.dataset.delp, { method: 'DELETE' }); load(); } });
  };

  /* finestra utente: accesso + ruolo + account mywellness */
  const openUser = (u) => {
    const isNew = !u; u = u || { username: '', displayName: '', isAdmin: false };
    const pr = isNew ? null : profiles.find(p => p.userId === u.id);
    const orph = profiles.filter(p => !p.userId);
    const canEdit = admin || (!isNew && u.id === me.user.id);
    const dlg = $('#udlg');
    dlg.innerHTML = `<form method="dialog" class="dlg"><div class="row" style="gap:12px">${isNew ? '' : avatar(pr, u.displayName, 56)}<div><h2 style="margin:0">${isNew ? 'Nuovo utente' : esc(u.displayName || u.username)}</h2>${pr ? `<small class="mut">${esc(pr.displayName || '')}${pr.email ? ' · ' + esc(pr.email) : ''}</small>` : ''}</div></div>
      <h3>Accesso al gateway</h3><div class="grid">
        <div><label>Nome utente (per entrare)</label><input id="d-u" value="${esc(u.username)}" autocomplete="off" ${canEdit ? '' : 'disabled'}></div>
        <div><label>Nome</label><input id="d-d" value="${esc(u.displayName)}" ${canEdit ? '' : 'disabled'}></div>
        <div><label>${isNew ? 'Password gateway (almeno 6 caratteri)' : 'Nuova password gateway (vuoto = invariata)'}</label><input id="d-p" type="password" autocomplete="new-password" ${canEdit ? '' : 'disabled'}></div>
        <div><label>Ruolo</label><select id="d-a" ${admin ? '' : 'disabled'}><option value="0" ${u.isAdmin ? '' : 'selected'}>Utente</option><option value="1" ${u.isAdmin ? 'selected' : ''}>Amministratore</option></select></div>
      </div>
      <h3>Account mywellness</h3>
      ${pr ? `<div class="grid">
        <div><label>Email mywellness</label><input value="${esc(pr.username)}" disabled></div>
        <div><label>Nuova password mywellness (vuoto = invariata)</label><input id="d-mp" type="password" autocomplete="new-password"></div>
        <div><label>Etichetta</label><input id="d-ml" value="${esc(pr.label)}"></div>
        <div><label>Massimo prenotazioni attive</label><input id="d-mm" type="number" min="1" value="${pr.maxBookings}"></div>
        <div><label>Visibilità</label><select id="d-mv"><option value="0" ${pr.ownerUserIds && pr.ownerUserIds.length ? '' : 'selected'}>Famiglia (tutti gli utenti)</option><option value="1" ${pr.ownerUserIds && pr.ownerUserIds.length ? 'selected' : ''}>Solo questa persona</option></select></div>
        <div><label>Centro</label><input value="${esc(pr.facilityName)}" disabled></div></div>
        <p class="row"><button type="button" class="small" id="d-relogin">Rifai login mywellness</button><button type="button" class="small" id="d-unlink">Scollega account</button><button type="button" class="danger" id="d-delp">Rimuovi account</button></p>`
      : `<p class="mut">${isNew ? "Obbligatorio per gli utenti normali (senza account non vedrebbero nulla); facoltativo per un amministratore solo locale, che vede e gestisce tutti i profili." : 'Nessun account collegato.'}</p>
        ${orph.length && !isNew ? `<label>Assegna un account esistente</label><select id="d-orph"><option value="">— nuovo account qui sotto —</option>${orph.map(p => `<option value="${p.id}">${esc(p.label)} · ${esc(p.username)}</option>`).join('')}</select>` : ''}
        <div class="grid" id="d-newmw">
        <div><label>Email mywellness</label><input id="d-mu" autocomplete="off"></div><div><label>Password mywellness</label><input id="d-mpw" type="password" autocomplete="new-password"></div>
        <div><label>Centro (URL widget)</label><input id="d-mf" value="wellnesstown"></div><div><label>Massimo prenotazioni attive</label><input id="d-mm2" type="number" min="1" value="5"></div>
        <div><label>Visibilità</label><select id="d-mv2"><option value="0">Famiglia (tutti gli utenti)</option><option value="1">Solo questa persona</option></select></div></div>`}
      <div id="d-msg"></div>
      <p class="row"><button type="button" class="primary" id="d-save">${isNew ? 'Crea utente' : 'Salva'}</button><button type="button" class="small" id="d-close">Chiudi</button>${!isNew && admin && u.id !== me.user.id ? `<span style="flex:1"></span><button type="button" class="danger" id="d-del">Elimina utente</button>` : ''}</p>
    </form>`;
    i18n.apply(dlg); i18n.addEyes(dlg);
    dlg.showModal();
    $('#d-close').onclick = () => dlg.close();
    const m = $('#d-msg');
    if ($('#d-orph')) $('#d-orph').onchange = () => { $('#d-newmw').style.display = $('#d-orph').value ? 'none' : ''; };
    $('#d-save').onclick = async () => {
      try {
        msg(m, i18n.t('Salvataggio…'), true);
        let userId = u.id;
        if (isNew) {
          const hasMw = $('#d-mu').value && $('#d-mpw').value;
          if (!hasMw && $('#d-a').value !== '1') { msg(m, i18n.t('Un utente normale deve avere il suo account mywellness (altrimenti non vedrebbe nulla). Per un account solo locale scegli il ruolo Amministratore.')); return; }
          const r = await api('/users', { method: 'POST', body: { username: $('#d-u').value, displayName: $('#d-d').value, password: $('#d-p').value, isAdmin: $('#d-a').value === '1',
            mywellness: hasMw ? { label: $('#d-d').value, username: $('#d-mu').value, password: $('#d-mpw').value, facilityUrl: $('#d-mf').value, maxBookings: +$('#d-mm2').value, private: $('#d-mv2').value === '1' } : undefined } });
          if (r.profileError) { msg(m, i18n.t('Utente creato, ma profilo mywellness non aggiunto:') + ' ' + r.profileError); await load(); return; }
        } else {
          if (canEdit) {
            const body = { username: $('#d-u').value, displayName: $('#d-d').value };
            if (admin) body.isAdmin = $('#d-a').value === '1';
            if ($('#d-p').value) body.password = $('#d-p').value;
            if (admin) await api('/users/' + userId, { method: 'PUT', body });
            else if ($('#d-p').value) { msg(m, i18n.t('Per cambiare la tua password usa l\'app o chiedi a un amministratore.')); }
          }
          if (pr) {
            const body = { label: $('#d-ml').value, maxBookings: +$('#d-mm').value, private: $('#d-mv').value === '1' };
            if ($('#d-mp').value) body.password = $('#d-mp').value;
            await api('/profiles/' + pr.id, { method: 'PUT', body });
          } else if ($('#d-orph') && $('#d-orph').value) {
            await api('/profiles/' + $('#d-orph').value, { method: 'PUT', body: { userId } });
          } else if ($('#d-mu') && $('#d-mu').value) {
            const body = { label: $('#d-d').value || u.displayName, username: $('#d-mu').value, password: $('#d-mpw').value, facilityUrl: $('#d-mf').value, maxBookings: +$('#d-mm2').value, private: $('#d-mv2').value === '1' };
            if (userId === me.user.id) body.mine = true; else body.userId = userId;
            await api('/profiles', { method: 'POST', body });
          }
        }
        dlg.close(); await load();
      } catch (e) { msg(m, e.message); }
    };
    if ($('#d-del')) $('#d-del').onclick = async () => { if (!confirm(i18n.t('Eliminare utente?'))) return; try { await api('/users/' + u.id, { method: 'DELETE' }); dlg.close(); load(); } catch (e) { msg(m, e.message); } };
    if ($('#d-relogin')) $('#d-relogin').onclick = async () => { try { await api(`/profiles/${pr.id}/relogin`, { method: 'POST' }); msg(m, i18n.t('Login mywellness riuscito.'), true); } catch (e) { msg(m, e.message); } };
    if ($('#d-unlink')) $('#d-unlink').onclick = async () => { try { await api('/profiles/' + pr.id, { method: 'PUT', body: { userId: '' } }); dlg.close(); load(); } catch (e) { msg(m, e.message); } };
    if ($('#d-delp')) $('#d-delp').onclick = async () => { if (!confirm(i18n.t('Rimuovere il profilo e le sue lezioni seguite?'))) return; try { await api('/profiles/' + pr.id, { method: 'DELETE' }); dlg.close(); load(); } catch (e) { msg(m, e.message); } };
  };
  if (admin) $('#newu').onclick = () => openUser(null);
  load();
}

/* ---------- impostazioni ---------- */
async function settings() {
  const s = await api('/settings'); const st = await api('/status');
  const ro = !me.user.isAdmin;
  main.innerHTML = `<div class="card"><h1>Impostazioni</h1>
    <h2>Aspetto e lingua</h2><div class="grid">
      <div><label>Tema</label><select id="theme"><option value="system">Sistema</option><option value="light">Chiaro</option><option value="dark">Scuro</option></select></div>
      <div><label>Lingua</label><select id="lang"><option value="it">Italiano</option><option value="en">English</option><option value="es">Español</option><option value="fr">Français</option><option value="de">Deutsch</option></select></div>
    </div><p class="mut">Valgono per questo browser.</p></div>
    <div class="card"><h2>Motore</h2>
    <label><input type="checkbox" id="fs" ${s.followServerOpenTime ? 'checked' : ''} ${ro ? 'disabled' : ''} style="width:auto"> Segui l'orario di apertura comunicato dal centro</label>
    <h2 style="margin-top:14px">Regole di prenotazione (quanti giorni prima apre ogni tipo di lezione)</h2>
    <p class="mut">Puoi avere tutte le regole che vuoi: in ogni riga scrivi il testo da cercare nel nome della lezione (es. <b>Reformer</b> → 3 giorni prima alle 05:00); la prima regola che corrisponde decide, la riga <b>*</b> vale per tutte le altre (es. 7 giorni). Con "Segui l'orario del centro" attivo queste regole servono solo quando mywellness non comunica l'apertura di una lezione.</p>
    <div id="rules"></div>${ro ? '' : '<p><button class="primary" id="addr">+ Aggiungi regola</button></p>'}
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
  i18n.bindPrefs();
  let rules = s.openRules.slice();
  if (!ro && rules.length === 1) rules.unshift({ pattern: '', daysBefore: 3, hour: 5, minute: 0 });   // riga d'esempio da completare
  let classNames = [];
  try { if (currentProfile) classNames = [...new Set((await api(`/profiles/${currentProfile}/classes`)).map(c => c.name))]; } catch {}
  const matches = (pat) => { pat = pat.trim().toLowerCase(); if (!pat || pat === '*') return []; return classNames.filter(n => n.toLowerCase().includes(pat)); };
  const drawRules = () => {
    $('#rules').innerHTML = `<table><tr><th>Testo nel nome</th><th>Giorni prima</th><th>Ora</th><th>Lezioni intercettate</th><th></th></tr>${rules.map((r, i) => `<tr><td><input data-p="${i}" value="${esc(r.pattern)}" ${ro || r.pattern === '*' ? 'disabled' : ''} placeholder="es. Reformer"></td><td><input data-d="${i}" type="number" min="0" value="${r.daysBefore}" ${ro ? 'disabled' : ''} style="width:90px"></td><td><input data-t="${i}" type="time" value="${String(r.hour).padStart(2, '0')}:${String(r.minute).padStart(2, '0')}" ${ro ? 'disabled' : ''} style="width:120px"></td><td class="mut" style="font-size:12px">${r.pattern === '*' ? 'tutte le altre' : (matches(r.pattern).join(', ') || (r.pattern.trim() ? 'nessuna lezione in calendario' : 'scrivi un testo'))}</td><td>${!ro && r.pattern !== '*' ? `<button class="danger" data-x="${i}">Elimina</button>` : ''}</td></tr>`).join('')}</table>`;
    main.querySelectorAll('[data-p]').forEach(i => i.oninput = () => { rules[+i.dataset.p].pattern = i.value; const td = i.closest('tr').children[3]; td.textContent = matches(i.value).join(', ') || (i.value.trim() ? 'nessuna lezione in calendario' : 'scrivi un testo'); });
    main.querySelectorAll('[data-d]').forEach(i => i.oninput = () => rules[+i.dataset.d].daysBefore = +i.value);
    main.querySelectorAll('[data-t]').forEach(i => i.oninput = () => { const [h, m] = i.value.split(':'); rules[+i.dataset.t].hour = +h; rules[+i.dataset.t].minute = +m; });
    main.querySelectorAll('[data-x]').forEach(b => b.onclick = () => { rules.splice(+b.dataset.x, 1); drawRules(); });
  };
  drawRules();
  $('#addr') && ($('#addr').onclick = () => { rules.splice(Math.max(0, rules.length - 1), 0, { pattern: '', daysBefore: 3, hour: 5, minute: 0 }); drawRules(); });
  $('#save') && ($('#save').onclick = async () => {
    try { await api('/settings', { method: 'PUT', body: { followServerOpenTime: $('#fs').checked, openRules: rules.filter(r => r.pattern.trim()), leadMilliseconds: +$('#lead').value, burstSeconds: +$('#burst').value, pollSeconds: +$('#poll').value, daysAhead: +$('#days').value, priorityNotifications: $('#prio').checked } }); msg($('#m'), 'Impostazioni salvate.', true); } catch (e) { msg($('#m'), e.message); }
  });
  $('#tn').onclick = async () => { const r = await api('/devices/test-notification', { method: 'POST' }); $('#tnm').textContent = r.push ? `inviata a ${r.sent} dispositivi` : 'APNs non configurato'; };
}

/* ---------- registro ---------- */
async function log() {
  const names = Object.fromEntries(profiles.map(p => [p.id, p.label]));
  main.innerHTML = `<div class="card"><div class="row"><h1 style="margin:0">Registro attività</h1>
    <select id="lsel" style="width:auto"><option value="">Tutti i profili</option>${profiles.map(p => `<option value="${p.id}" ${p.id === currentProfile ? 'selected' : ''}>${esc(p.label)}</option>`).join('')}</select>
    <select id="llev" style="width:auto"><option value="">Tutti i livelli</option><option value="success">Successi</option><option value="warn">Avvisi</option><option value="error">Errori</option></select>
    <button class="small" id="rf">Aggiorna</button></div><div id="loglist" class="log" style="margin-top:12px"></div></div>`;
  const myView = viewId;
  const load = async () => {
    const pid = $('#lsel').value, lev = $('#llev').value;
    let lines; try { lines = await api('/log?limit=500' + (pid ? '&profile=' + pid : '')); } catch (e) { return; }
    if (!stillHere(myView)) return;
    lines = lines.filter(l => !lev || l.level === lev);
    $('#loglist').innerHTML = lines.map(l => `<div><span class="mut">${fmtD(l.time)}</span> <span class="lv-${l.level}">●</span> ${l.profileId ? `<b>${esc(names[l.profileId] || '')}</b> ` : ''}${esc(l.text)}</div>`).join('') || '<p class="mut">Nessuna attività.</p>';
  };
  $('#lsel').onchange = load; $('#llev').onchange = load; $('#rf').onclick = load; load();
  autoRefresh(myView, load); window.__refreshNow = () => stillHere(myView) && load();
}

render();
