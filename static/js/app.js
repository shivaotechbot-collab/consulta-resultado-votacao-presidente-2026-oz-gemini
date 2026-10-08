const API_BASE = '/api';

const state = {
  summary: null,
  candidates: [],
  zones: [],
  selectedCandidate: '',
  selectedZone: '',
  selectedSecao: '',
  searchQuery: '',
};

document.addEventListener('DOMContentLoaded', () => {
  initApp();
  setupEventListeners();
});

async function initApp() {
  await Promise.all([
    loadSummary(),
    loadCandidates(),
    loadZones(),
  ]);
  performSearch();
}

function setupEventListeners() {
  const searchForm = document.getElementById('searchForm');
  searchForm.addEventListener('submit', (e) => {
    e.preventDefault();
    state.selectedZone = document.getElementById('selectZone').value;
    state.selectedSecao = document.getElementById('inputSecao').value.trim();
    state.selectedCandidate = document.getElementById('selectCandidate').value;
    state.searchQuery = document.getElementById('inputLocal').value.trim();
    performSearch();
  });

  document.getElementById('btnReset').addEventListener('click', () => {
    searchForm.reset();
    state.selectedZone = '';
    state.selectedSecao = '';
    state.selectedCandidate = '';
    state.searchQuery = '';
    document.querySelectorAll('.candidate-card').forEach(c => c.classList.remove('active'));
    performSearch();
  });

  document.getElementById('modalCloseBtn').addEventListener('click', closeModal);
  document.getElementById('modalOverlay').addEventListener('click', (e) => {
    if (e.target === e.currentTarget) closeModal();
  });
  document.addEventListener('keydown', (e) => {
    if (e.key === 'Escape') closeModal();
  });
}

function formatNumber(num) {
  return new Intl.NumberFormat('pt-BR').format(num || 0);
}

function formatPercent(pct) {
  return (pct || 0).toFixed(2).replace('.', ',') + '%';
}

async function loadSummary() {
  try {
    const res = await fetch(`${API_BASE}/summary`);
    const data = await res.json();
    state.summary = data;

    document.getElementById('valAptos').textContent = formatNumber(data.total_aptos);
    document.getElementById('valComparecimento').textContent = formatNumber(data.total_comparecimento);
    document.getElementById('valAbstencoes').textContent = formatNumber(data.total_abstencoes);
    document.getElementById('valSecoes').textContent = formatNumber(data.total_secoes);
    document.getElementById('valLocais').textContent = formatNumber(data.total_locais);

    const compPct = ((data.total_comparecimento / data.total_aptos) * 100).toFixed(1);
    const abstPct = ((data.total_abstencoes / data.total_aptos) * 100).toFixed(1);
    document.getElementById('subComparecimento').textContent = `${compPct}% de presença`;
    document.getElementById('subAbstencoes').textContent = `${abstPct}% de abstenção`;
  } catch (err) {
    console.error('Erro ao carregar resumo:', err);
  }
}

async function loadCandidates() {
  try {
    const res = await fetch(`${API_BASE}/candidates`);
    const list = await res.json();
    state.candidates = list;

    const select = document.getElementById('selectCandidate');
    const container = document.getElementById('candidatesStrip');
    container.innerHTML = '';

    list.forEach((c) => {
      // Add to select
      const opt = document.createElement('option');
      opt.value = c.nr_votavel;
      opt.textContent = `${c.nr_votavel} - ${c.nm_votavel}`;
      select.appendChild(opt);

      // Add to cards strip (show top contenders)
      const card = document.createElement('div');
      card.className = 'candidate-card';
      card.dataset.num = c.nr_votavel;
      card.innerHTML = `
        <div>
          <div class="candidate-top">
            <span class="candidate-number">Nº ${c.nr_votavel}</span>
            <span class="candidate-pct">${formatPercent(c.percentual)}</span>
          </div>
          <div class="candidate-name">${c.nm_votavel}</div>
        </div>
        <div>
          <div class="candidate-votes">${formatNumber(c.votos)} votos válidos/totais</div>
          <div class="progress-bar-bg">
            <div class="progress-bar-fill" style="width: ${Math.min(c.percentual, 100)}%;"></div>
          </div>
        </div>
      `;

      card.addEventListener('click', () => {
        if (state.selectedCandidate === c.nr_votavel) {
          state.selectedCandidate = '';
          card.classList.remove('active');
          select.value = '';
        } else {
          state.selectedCandidate = c.nr_votavel;
          document.querySelectorAll('.candidate-card').forEach(el => el.classList.remove('active'));
          card.classList.add('active');
          select.value = c.nr_votavel;
        }
        performSearch();
      });

      container.appendChild(card);
    });
  } catch (err) {
    console.error('Erro ao carregar candidatos:', err);
  }
}

async function loadZones() {
  try {
    const res = await fetch(`${API_BASE}/zones`);
    const list = await res.json();
    state.zones = list;

    const select = document.getElementById('selectZone');
    list.forEach(z => {
      const opt = document.createElement('option');
      opt.value = z.nr_zona;
      opt.textContent = `Zona ${z.nr_zona} (${z.total_secoes} seções)`;
      select.appendChild(opt);
    });
  } catch (err) {
    console.error('Erro ao carregar zonas:', err);
  }
}

async function performSearch() {
  const tbody = document.getElementById('resultsBody');
  const countSpan = document.getElementById('resultsCount');
  tbody.innerHTML = `<tr><td colspan="7" class="empty-state"><span class="spinner"></span><br>Consultando seções...</td></tr>`;

  const params = new URLSearchParams();
  if (state.selectedZone) params.append('zona', state.selectedZone);
  if (state.selectedSecao) params.append('secao', state.selectedSecao);
  if (state.selectedCandidate) params.append('candidate', state.selectedCandidate);
  if (state.searchQuery) params.append('local', state.searchQuery);
  params.append('limit', '100');

  try {
    const res = await fetch(`${API_BASE}/search?${params.toString()}`);
    const results = await res.json();

    if (!results || results.length === 0) {
      tbody.innerHTML = `<tr><td colspan="7" class="empty-state">Nenhum resultado encontrado para os filtros selecionados.</td></tr>`;
      countSpan.textContent = '0 registros encontrados';
      return;
    }

    countSpan.textContent = `${results.length} seções exibidas`;
    tbody.innerHTML = '';

    results.forEach((row) => {
      const tr = document.createElement('tr');
      tr.innerHTML = `
        <td><span class="tag-zona">Zona ${row.nr_zona}</span></td>
        <td><span class="tag-secao">Seção ${row.nr_secao}</span></td>
        <td>
          <span class="local-name">${row.nm_local_votacao}</span>
          <span class="local-address">${row.ds_local_votacao_endereco}</span>
        </td>
        <td><strong>${row.nm_votavel}</strong> <small>(${row.nr_votavel})</small></td>
        <td><strong>${formatNumber(row.qt_votos)}</strong></td>
        <td><span style="font-weight: 600; color: var(--accent-zen);">${formatPercent(row.percentual_na_secao)}</span></td>
        <td style="text-align: right;">
          <button class="btn-detail" onclick="openSectionModal(${row.nr_zona}, ${row.nr_secao})">Detalhes</button>
        </td>
      `;
      tbody.appendChild(tr);
    });
  } catch (err) {
    console.error('Erro ao pesquisar:', err);
    tbody.innerHTML = `<tr><td colspan="7" class="empty-state" style="color: #c53030;">Falha ao comunicar com o servidor.</td></tr>`;
  }
}

async function openSectionModal(zona, secao) {
  const modal = document.getElementById('modalOverlay');
  const detailsDiv = document.getElementById('modalSectionDetails');
  detailsDiv.innerHTML = '<div class="empty-state"><span class="spinner"></span></div>';
  modal.classList.add('open');

  try {
    const res = await fetch(`${API_BASE}/section?zona=${zona}&secao=${secao}`);
    if (!res.ok) throw new Error('Não encontrado');
    const data = await res.json();

    document.getElementById('modalTitle').textContent = `Zona ${data.nr_zona} • Seção ${data.nr_secao}`;
    document.getElementById('modalSubtitle').textContent = `${data.nm_local_votacao} — ${data.ds_local_votacao_endereco}`;

    let rowsHtml = '';
    (data.resultados || []).forEach(r => {
      rowsHtml += `
        <tr>
          <td><span class="candidate-number" style="font-size:0.75rem;">${r.nr_votavel}</span></td>
          <td><strong>${r.nm_votavel}</strong></td>
          <td style="text-align: right;">${formatNumber(r.qt_votos)}</td>
          <td style="text-align: right; font-weight:600; color:var(--accent-zen);">${formatPercent(r.percentual)}</td>
        </tr>
      `;
    });

    detailsDiv.innerHTML = `
      <div class="section-stats-bar">
        <div>
          <div class="stat-box-title">Eleitores Aptos</div>
          <div class="stat-box-num">${formatNumber(data.qt_aptos)}</div>
        </div>
        <div>
          <div class="stat-box-title">Comparecimento</div>
          <div class="stat-box-num">${formatNumber(data.qt_comparecimento)}</div>
        </div>
        <div>
          <div class="stat-box-title">Abstenções</div>
          <div class="stat-box-num">${formatNumber(data.qt_abstencoes)}</div>
        </div>
        <div>
          <div class="stat-box-title">Votos Nominais</div>
          <div class="stat-box-num">${formatNumber(data.qt_votos_nominais)}</div>
        </div>
      </div>
      <table class="zen-table" style="margin-top: 1rem;">
        <thead>
          <tr>
            <th style="width: 45px;">Nº</th>
            <th>Candidato / Opção</th>
            <th style="text-align: right;">Votos</th>
            <th style="text-align: right;">% Seção</th>
          </tr>
        </thead>
        <tbody>
          ${rowsHtml}
        </tbody>
      </table>
    `;
  } catch (err) {
    detailsDiv.innerHTML = '<div class="empty-state" style="color:#c53030;">Não foi possível carregar os detalhes da seção.</div>';
  }
}

function closeModal() {
  document.getElementById('modalOverlay').remove('open');
  document.getElementById('modalOverlay').classList.remove('open');
}
