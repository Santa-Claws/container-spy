'use strict';

const POLL_INTERVAL = 5000;

function stateClass(state) {
  if (state === 'running') return 'state-running';
  if (state === 'exited' || state === 'dead') return 'state-exited';
  if (state === 'paused' || state === 'restarting') return 'state-paused';
  return 'state-other';
}

function render(containers) {
  const el = document.getElementById('container-table');

  if (!containers || containers.length === 0) {
    el.innerHTML = '<p style="color:#666;padding:16px 4px">No containers found. Add a server via the TUI.</p>';
    return;
  }

  // Group containers.
  const groups = {};
  for (const c of containers) {
    if (!groups[c.group]) groups[c.group] = [];
    groups[c.group].push(c);
  }

  let html = '';
  for (const [group, items] of Object.entries(groups)) {
    html += `<div class="group-header">▸ ${escHtml(group)}</div>`;
    html += '<table><thead><tr><th>CONTAINER</th><th>SERVER</th><th>STATUS</th><th>IMAGE</th></tr></thead><tbody>';
    for (const c of items) {
      html += `<tr>
        <td>${escHtml(c.name)}</td>
        <td title="${escHtml(c.server_host)}">${escHtml(c.server_name)}</td>
        <td class="${stateClass(c.state)}">${escHtml(c.status)}</td>
        <td>${escHtml(c.image)}</td>
      </tr>`;
    }
    html += '</tbody></table>';
  }

  el.innerHTML = html;
}

function escHtml(s) {
  return String(s || '')
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
    .replace(/"/g, '&quot;');
}

async function poll() {
  const errEl = document.getElementById('error-msg');
  const timeEl = document.getElementById('refresh-time');
  try {
    const res = await fetch('/api/containers');
    if (!res.ok) throw new Error(`HTTP ${res.status}`);
    const data = await res.json();
    render(data);
    errEl.classList.add('hidden');
    timeEl.textContent = 'refreshed ' + new Date().toLocaleTimeString();
  } catch (e) {
    errEl.textContent = 'Error fetching containers: ' + e.message;
    errEl.classList.remove('hidden');
  }
}

poll();
setInterval(poll, POLL_INTERVAL);
