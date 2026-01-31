/*
Концепция: клиент к API CommentTree
Название программы: CommentTree UI Script
Цель/назначение: загрузка дерева, создание/удаление, поиск
Аудитория: учебный проект
Входные данные: DOM-поля формы + ответы API
Выходные данные: DOM-рендер дерева/результатов
Ограничения: без фреймворков
Инвариант: дерево рендерится рекурсивно; parent_id корректный
*/

const API = {
    tree: (page, limit, sort) =>
      fetch(`/comments?page=${page}&limit=${limit}&sort=${encodeURIComponent(sort)}`)
        .then(r => r.json()),
  
    create: (payload) =>
      fetch('/comments', {
        method: 'POST',
        headers: {'Content-Type':'application/json'},
        body: JSON.stringify(payload)
      }).then(async r => {
        const t = await r.text();
        if (!r.ok) throw new Error(t || r.statusText);
        return JSON.parse(t);
      }),
  
    del: (id) =>
      fetch(`/comments/${id}`, {method:'DELETE'}).then(async r => {
        const t = await r.text();
        if (!r.ok) throw new Error(t || r.statusText);
        return JSON.parse(t);
      }),
  
    search: (q, page, limit, sort) =>
      fetch(`/comments/search?q=${encodeURIComponent(q)}&page=${page}&limit=${limit}&sort=${encodeURIComponent(sort)}`)
        .then(r => r.json())
  };
  
  let currentParentId = 0;
  
  const el = (id) => document.getElementById(id);
  
  function setStatus(msg) {
    el('status').textContent = msg || '';
  }
  
  function setParent(id) {
    currentParentId = id;
    el('parentBadge').textContent = String(id);
  }
  
  function fmtTime(iso) {
    try { return new Date(iso).toLocaleString(); } catch { return iso; }
  }
  
  function escapeHtml(s) {
    return String(s)
      .replaceAll('&','&amp;')
      .replaceAll('<','&lt;')
      .replaceAll('>','&gt;')
      .replaceAll('"','&quot;')
      .replaceAll("'","&#039;");
  }
  
  function renderNode(node, depth = 0) {
    const wrap = document.createElement('div');
    wrap.className = 'node';
    wrap.style.marginLeft = (depth * 14) + 'px';
  
    const header = document.createElement('div');
    header.className = 'node-header';
    header.innerHTML = `
      <span class="pill">#${escapeHtml(node.id)}</span>
      <span><b>${escapeHtml(node.author)}</b></span>
      <span class="muted">${escapeHtml(fmtTime(node.created_at))}</span>
    `;
  
    const text = document.createElement('div');
    text.className = 'node-text';
    text.textContent = node.text;
  
    const actions = document.createElement('div');
    actions.className = 'actions';
  
    const replyBtn = document.createElement('button');
    replyBtn.textContent = 'Ответить';
    replyBtn.onclick = () => {
      setParent(node.id);
      window.scrollTo({top:0, behavior:'smooth'});
    };
  
    const delBtn = document.createElement('button');
    delBtn.textContent = 'Удалить';
    delBtn.onclick = async () => {
      if (!confirm(`Удалить комментарий #${node.id} и всё поддерево?`)) return;
      try {
        setStatus('Удаление...');
        await API.del(node.id);
        setStatus('Удалено');
        await loadTree();
      } catch (e) {
        setStatus('Ошибка удаления: ' + e.message);
      }
    };
  
    actions.appendChild(replyBtn);
    actions.appendChild(delBtn);
  
    wrap.appendChild(header);
    wrap.appendChild(text);
    wrap.appendChild(actions);
  
    if (node.children && node.children.length) {
      for (const ch of node.children) {
        wrap.appendChild(renderNode(ch, depth + 1));
      }
    }
  
    return wrap;
  }
  
  async function loadTree() {
    const page = Number(el('page').value || 1);
    const limit = Number(el('limit').value || 10);
    const sort = el('sortMode').value || 'created_at_asc';
  
    setStatus('Загрузка дерева...');
    el('searchResults').innerHTML = '';
  
    const data = await API.tree(page, limit, sort);
  
    el('meta').textContent = `Корни: total=${data.meta.total}, page=${data.meta.page}, limit=${data.meta.limit}, sort=${data.sort}`;
  
    const tree = el('tree');
    tree.innerHTML = '';
  
    if (!data.items || !data.items.length) {
      tree.innerHTML = '<div class="muted">Пока нет комментариев.</div>';
      setStatus('');
      return;
    }
  
    for (const n of data.items) tree.appendChild(renderNode(n, 0));
    setStatus('');
  }
  
  async function doCreate() {
    const author = el('author').value || '';
    const text = el('text').value || '';
    try {
      setStatus('Создание...');
      await API.create({ parent_id: currentParentId, author, text });
      el('text').value = '';
      setStatus('Создано');
      await loadTree();
    } catch (e) {
      setStatus('Ошибка: ' + e.message);
    }
  }
  
  async function doSearch() {
    const q = (el('searchQ').value || '').trim();
    if (!q) return;
  
    const page = 1;
    const limit = 20;
    const sort = 'created_at_desc';
  
    setStatus('Поиск...');
    const data = await API.search(q, page, limit, sort);
  
    const box = el('searchResults');
    box.innerHTML = '';
  
    const title = document.createElement('div');
    title.className = 'muted';
    title.textContent = `Результаты поиска: total=${data.meta.total}, показано=${data.items.length}`;
    box.appendChild(title);
  
    if (!data.items.length) {
      const empty = document.createElement('div');
      empty.className = 'muted';
      empty.textContent = 'Ничего не найдено.';
      box.appendChild(empty);
      setStatus('');
      return;
    }
  
    const ul = document.createElement('ul');
    ul.className = 'list';
  
    for (const c of data.items) {
      const li = document.createElement('li');
      li.innerHTML = `<span class="pill">#${escapeHtml(c.id)}</span> <b>${escapeHtml(c.author)}</b>
        <span class="muted">parent=${escapeHtml(c.parent_id)} · ${escapeHtml(fmtTime(c.created_at))}</span>
        <div class="muted">${escapeHtml(c.text).slice(0, 180)}${c.text.length > 180 ? '…' : ''}</div>`;
      ul.appendChild(li);
    }
  
    box.appendChild(ul);
    setStatus('');
  }
  
  el('createBtn').addEventListener('click', doCreate);
  el('reloadBtn').addEventListener('click', loadTree);
  el('searchBtn').addEventListener('click', doSearch);
  el('clearSearchBtn').addEventListener('click', () => {
    el('searchQ').value = '';
    el('searchResults').innerHTML = '';
    setStatus('');
  });
  el('resetParentBtn').addEventListener('click', () => setParent(0));
  
  loadTree();