// Адрес API. Если сервер запущен на другом порту, поменяйте здесь.
const API = 'http://localhost:8080';

const STATUS_LABELS = {
    new: 'Новая',
    in_progress: 'В работе',
    done: 'Выполнена',
};

// Русские тексты для ошибок, которые возвращает сервер
const ERROR_TEXTS = {
    'invalid JSON': 'Сервер не смог прочитать данные запроса',
    'invalid id': 'Некорректный номер заявки',
    'invalid ticket data': 'Заполните название, выберите автора и категорию',
    'invalid status': 'Недопустимый статус заявки',
    'related record not found': 'Автор или категория не найдены - обновите страницу',
    'ticket not found': 'Заявка не найдена - возможно, ее уже удалили',
    'internal server error': 'Ошибка на сервере, попробуйте еще раз',
};

const users = new Map();
const categories = new Map();

const alertBox = document.getElementById('alert');
const alertText = document.getElementById('alert-text');
const form = document.getElementById('ticket-form');
const submitBtn = document.getElementById('submit-btn');
const list = document.getElementById('ticket-list');
const state = document.getElementById('state');
const counter = document.getElementById('counter');

// ---------- работа с API

class ApiError extends Error {
    constructor(status, message) {
        super(message);
        this.status = status;
    }
}

// Один помощник на все запросы: отправляет JSON и превращает
// любой ответ не из диапазона 2xx в понятную ошибку.
async function request(method, path, body) {
    const options = { method, headers: {} };
    if (body !== undefined) {
        options.headers['Content-Type'] = 'application/json';
        options.body = JSON.stringify(body);
    }

    let res;
    try {
        res = await fetch(API + path, options);
    } catch {
        throw new ApiError(0, 'Сервер недоступен. Проверьте, что запущен docker compose up');
    }

    if (res.status === 204) {
        return null;
    }

    let data = null;
    try {
        data = await res.json();
    } catch {
        // тело не JSON - оставляем data пустым
    }

    if (!res.ok) {
        const serverText = data && data.error ? data.error : '';
        const text = ERROR_TEXTS[serverText] || serverText || 'Неизвестная ошибка';
        throw new ApiError(res.status, `Ошибка ${res.status}: ${text}`);
    }

    return data;
}

// ---------- сообщения об ошибках

function showError(err) {
    alertText.textContent = err instanceof ApiError ? err.message : 'Непредвиденная ошибка: ' + err;
    alertBox.hidden = false;
    console.error(err);
}

function hideError() {
    alertBox.hidden = true;
}

document.getElementById('alert-close').addEventListener('click', hideError);

// ---------- справочники для формы

function fillSelect(select, map, placeholder) {
    select.innerHTML = '';
    const empty = new Option(placeholder, '');
    select.add(empty);
    for (const [id, name] of map) {
        select.add(new Option(name, id));
    }
}

async function loadDictionaries() {
    const [usersData, categoriesData] = await Promise.all([
        request('GET', '/users'),
        request('GET', '/categories'),
    ]);

    users.clear();
    usersData.forEach((u) => users.set(u.id, u.name));
    categories.clear();
    categoriesData.forEach((c) => categories.set(c.id, c.name));

    fillSelect(document.getElementById('user-select'), users, 'Выберите автора');
    fillSelect(document.getElementById('category-select'), categories, 'Выберите категорию');
}

// ---------- список заявок

function formatDate(iso) {
    return new Date(iso).toLocaleString('ru-RU', {
        day: 'numeric',
        month: 'long',
        hour: '2-digit',
        minute: '2-digit',
    });
}

// Создаем элементы через textContent, а не innerHTML с данными,
// чтобы текст заявки не мог выполниться как HTML.
function renderTicket(t) {
    const li = document.createElement('li');
    li.className = 'ticket';
    li.dataset.status = t.status;

    const title = document.createElement('h3');
    title.className = 'ticket-title';
    title.textContent = `#${t.id} ${t.title}`;
    li.append(title);

    if (t.description) {
        const desc = document.createElement('p');
        desc.className = 'ticket-desc';
        desc.textContent = t.description;
        li.append(desc);
    }

    const meta = document.createElement('p');
    meta.className = 'ticket-meta';
    const author = users.get(t.user_id) || `пользователь ${t.user_id}`;
    const category = categories.get(t.category_id) || `категория ${t.category_id}`;
    meta.textContent = `${author}, ${category}, ${formatDate(t.created_at)}`;
    li.append(meta);

    const actions = document.createElement('div');
    actions.className = 'ticket-actions';

    const statusSelect = document.createElement('select');
    statusSelect.setAttribute('aria-label', `Статус заявки ${t.id}`);
    for (const [value, label] of Object.entries(STATUS_LABELS)) {
        statusSelect.add(new Option(label, value, false, value === t.status));
    }
    statusSelect.addEventListener('change', () => changeStatus(t, statusSelect, li));

    const delBtn = document.createElement('button');
    delBtn.type = 'button';
    delBtn.className = 'btn-delete';
    delBtn.textContent = 'Удалить';
    delBtn.addEventListener('click', () => deleteTicket(t));

    actions.append(statusSelect, delBtn);
    li.append(actions);

    return li;
}

function renderList(tickets) {
    list.innerHTML = '';

    if (tickets.length === 0) {
        state.textContent = 'Заявок пока нет. Создайте первую в форме выше.';
        state.hidden = false;
        counter.textContent = '';
        return;
    }

    state.hidden = true;
    const open = tickets.filter((t) => t.status !== 'done').length;
    counter.textContent = `Всего: ${tickets.length}, открытых: ${open}`;

    // новые заявки сверху
    [...tickets].reverse().forEach((t) => list.append(renderTicket(t)));
}

async function loadTickets() {
    const tickets = await request('GET', '/tickets');
    renderList(tickets);
}

// ---------- действия

async function changeStatus(ticket, select, li) {
    const previous = ticket.status;
    select.disabled = true;
    try {
        const updated = await request('PATCH', `/tickets/${ticket.id}`, { status: select.value });
        ticket.status = updated.status;
        li.dataset.status = updated.status;
        hideError();
    } catch (err) {
        select.value = previous; // возвращаем прежний статус, раз сервер отказал
        showError(err);
        if (err.status === 404) {
            loadTickets().catch(showError);
        }
    } finally {
        select.disabled = false;
    }
}

async function deleteTicket(ticket) {
    if (!confirm(`Удалить заявку #${ticket.id} «${ticket.title}»?`)) {
        return;
    }
    try {
        await request('DELETE', `/tickets/${ticket.id}`);
        hideError();
    } catch (err) {
        showError(err);
    }
    await loadTickets().catch(showError);
}

form.addEventListener('submit', async (e) => {
    e.preventDefault();

    const fd = new FormData(form);
    const title = fd.get('title').trim();
    const description = fd.get('description').trim();
    const userId = Number(fd.get('user_id'));
    const categoryId = Number(fd.get('category_id'));

    // Подсветка пустых полей. Основная проверка все равно на сервере.
    form.querySelectorAll('.invalid').forEach((el) => el.classList.remove('invalid'));
    if (!title) form.elements.title.classList.add('invalid');
    if (!userId) form.elements.user_id.classList.add('invalid');
    if (!categoryId) form.elements.category_id.classList.add('invalid');

    const body = { title, user_id: userId, category_id: categoryId };
    if (description) {
        body.description = description;
    }

    submitBtn.disabled = true;
    try {
        await request('POST', '/tickets', body);
        form.reset();
        hideError();
        await loadTickets();
        document.getElementById('tickets').scrollIntoView();
    } catch (err) {
        showError(err);
    } finally {
        submitBtn.disabled = false;
    }
});

// ---------- старт

async function init() {
    try {
        await loadDictionaries();
        await loadTickets();
    } catch (err) {
        state.textContent = 'Не удалось загрузить заявки.';
        showError(err);
    }
}

init();
