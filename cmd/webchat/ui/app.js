const $ = (selector, scope = document) => scope.querySelector(selector);
const $$ = (selector, scope = document) => Array.from(scope.querySelectorAll(selector));

const STORAGE_KEYS = {
    pinned: 'beeper.webchat.pinned-chats',
    archived: 'beeper.webchat.archived-chats',
    sidebarView: 'beeper.webchat.sidebar-view',
};

const VIEW_MODES = {
    all: 'all',
    pinned: 'pinned',
    unread: 'unread',
    archived: 'archived',
};

const preferences = loadPreferences();

const state = {
    loading: true,
    authenticated: false,
    user: null,
    chats: [],
    activeChat: null,
    messages: [],
    sending: false,
    chatFilter: '',
    loadingMessages: false,
    toast: null,
    pinnedChats: preferences.pinned,
    archivedChats: preferences.archived,
    sidebar: {
        view: preferences.view,
    },
};

let eventsBound = false;
let toastTimeout;

const api = {
    async request(path, options = {}) {
        const response = await fetch(path, {
            headers: {
                'Content-Type': 'application/json',
            },
            credentials: 'same-origin',
            ...options,
        });
        if (!response.ok) {
            const error = await response.json().catch(() => ({ message: 'Unknown error' }));
            throw new Error(error.message || 'Request failed');
        }
        return response.json();
    },

    session() {
        return this.request('/api/session', { headers: {} });
    },

    login(payload) {
        return this.request('/api/login', {
            method: 'POST',
            body: JSON.stringify(payload),
        });
    },

    logout() {
        return this.request('/api/logout', { method: 'POST', body: '{}' });
    },

    chats() {
        return this.request('/api/chats');
    },

    messages(chatId) {
        const params = new URLSearchParams({ chat_id: chatId });
        return this.request(`/api/messages?${params.toString()}`);
    },

    sendMessage(payload) {
        return this.request('/api/messages/send', {
            method: 'POST',
            body: JSON.stringify(payload),
        });
    },
};

function renderApp() {
    updatePageTitle();

    const app = $('#app');
    app.innerHTML = '';

    if (!state.authenticated) {
        app.append(renderLogin());
    } else {
        app.append(renderChatShell());
        enableSendOnEnter();
    }

    if (state.toast) {
        app.append(renderToast(state.toast));
        scheduleToastDismissal();
    }
}

function renderLogin() {
    const wrapper = document.createElement('div');
    wrapper.className = 'login';
    wrapper.innerHTML = `
        <aside class="login-hero">
            <div>
                <h1>Welcome to Beeper Web</h1>
                <p>Connect with every conversation from a single, elegant interface. Securely link your Beeper Desktop API key to start messaging instantly.</p>
            </div>
            <div>
                <p>✨ Features</p>
                <ul>
                    <li>Unified chat list with live message streams</li>
                    <li>Polished composer with rich theming</li>
                    <li>Responsive, immersive design</li>
                </ul>
            </div>
        </aside>
        <section class="login-form">
            <h2>Link your Beeper account</h2>
            <form id="login-form" class="login-fields">
                <div class="form-group">
                    <label for="token-input">Beeper API Access Token</label>
                    <input id="token-input" name="token" class="form-control" type="password" placeholder="bdp_********************************" autocomplete="off" required />
                </div>
                <div class="form-group">
                    <label for="base-url-input">Beeper API Base URL <span style="opacity:0.5">(optional)</span></label>
                    <input id="base-url-input" name="baseUrl" class="form-control" type="url" placeholder="http://localhost:23373" autocomplete="off" />
                </div>
                <button type="submit" class="button-primary" id="login-submit">Enter Beeper Web</button>
                <p class="form-helper" id="login-helper" hidden></p>
            </form>
        </section>
    `;

    const form = $('#login-form', wrapper);
    const submitBtn = $('#login-submit', wrapper);
    const helper = $('#login-helper', wrapper);

    form.addEventListener('submit', async (event) => {
        event.preventDefault();
        const formData = new FormData(form);
        const token = (formData.get('token') || '').trim();
        const baseUrl = (formData.get('baseUrl') || '').trim();

        if (!token) {
            helper.hidden = false;
            helper.textContent = 'Your Beeper API access token is required.';
            return;
        }

        submitBtn.disabled = true;
        helper.hidden = true;
        helper.textContent = '';

        try {
            const result = await api.login({ token, baseUrl });
            state.authenticated = result.authenticated;
            state.user = result.user;
            await bootstrapData();
        } catch (error) {
            helper.hidden = false;
            helper.textContent = error.message;
        } finally {
            submitBtn.disabled = false;
        }
    });

    return wrapper;
}

function renderChatShell() {
    const wrapper = document.createElement('div');
    wrapper.className = 'chat-shell';
    wrapper.innerHTML = `
        <aside class="sidebar">
            ${renderSidebarHeader()}
            <div class="sidebar-controls">
                <div class="chat-search">
                    <input id="chat-filter" class="form-control" type="search" placeholder="Search chats" value="${escapeHtml(state.chatFilter)}" autocomplete="off" />
                </div>
                <nav class="view-toggle" role="tablist">
                    ${renderSidebarViewToggle()}
                </nav>
            </div>
            <div class="chat-list-wrapper" id="chat-groups">
                ${renderSidebarGroups()}
            </div>
        </aside>
        <section class="chat-main" id="chat-main">
            ${renderChatMain()}
        </section>
    `;

    return wrapper;
}

function renderSidebarHeader() {
    const initials = getInitials(state.user?.subject || 'Beeper');
    const label = formatWorkspaceLabel(state.user?.subject);
    return `
        <header>
            <div class="workspace-meta">
                <div class="brand-mark">${escapeHtml(initials)}</div>
                <div>
                    <h1>Beeper</h1>
                    <span>${escapeHtml(label)}</span>
                </div>
            </div>
            <div class="sidebar-actions">
                <button class="icon-button" data-action="open-settings" title="Open settings">⚙</button>
                <button class="logout-btn" id="logout-btn" title="Sign out">Sign out</button>
            </div>
        </header>
    `;
}

function renderSidebarViewToggle() {
    const entries = [
        { key: VIEW_MODES.all, label: 'All' },
        { key: VIEW_MODES.pinned, label: 'Pinned' },
        { key: VIEW_MODES.unread, label: 'Unread' },
        { key: VIEW_MODES.archived, label: 'Archived' },
    ];
    return entries.map((entry) => {
        const isActive = state.sidebar.view === entry.key;
        return `
            <button type="button" class="view-toggle-btn${isActive ? ' is-active' : ''}" data-action="switch-view" data-view="${entry.key}">
                ${entry.label}
            </button>
        `;
    }).join('');
}

function renderSidebarGroups() {
    const filtered = getFilteredChats();
    const segments = segmentChats(filtered);

    if (!filtered.length) {
        return `
            <div class="chat-empty" id="chat-empty">
                <p>No chats match “${escapeHtml(state.chatFilter)}”.</p>
            </div>
        `;
    }

    const markup = [];
    if (state.sidebar.view === VIEW_MODES.all) {
        if (segments.pinned.length) {
            markup.push(renderChatGroup('Pinned', segments.pinned));
        }
        if (segments.unread.length) {
            markup.push(renderChatGroup('Unread', segments.unread));
        }
        if (segments.others.length) {
            markup.push(renderChatGroup('All Chats', segments.others));
        }
        if (segments.archived.length) {
            markup.push(renderChatGroup('Archived', segments.archived));
        }
    } else if (state.sidebar.view === VIEW_MODES.pinned) {
        markup.push(renderChatGroup('Pinned', segments.pinned));
        if (!segments.pinned.length) {
            markup.push(renderEmptyState('No pinned chats yet. Pin chats to quick access them.'));
        }
    } else if (state.sidebar.view === VIEW_MODES.unread) {
        markup.push(renderChatGroup('Unread', segments.unread));
        if (!segments.unread.length) {
            markup.push(renderEmptyState('All caught up! No unread conversations.'));
        }
    } else if (state.sidebar.view === VIEW_MODES.archived) {
        markup.push(renderChatGroup('Archived', segments.archived));
        if (!segments.archived.length) {
            markup.push(renderEmptyState('Archived chats live here. Archive any chat from the sidebar overflow.'));
        }
    }

    return markup.join('');
}

function renderChatGroup(title, chats) {
    if (!chats.length) {
        return '';
    }
    return `
        <section class="chat-group">
            <header>${escapeHtml(title)}</header>
            <div class="chat-list" role="list">
                ${chats.map(renderChatCard).join('')}
            </div>
        </section>
    `;
}

function renderEmptyState(message) {
    return `
        <div class="chat-empty" id="chat-empty">
            <p>${escapeHtml(message)}</p>
        </div>
    `;
}

function renderChatMain() {
    if (state.loadingMessages) {
        return `
            ${renderChatHeader()}
            <div class="messages is-loading" id="messages">
                ${['', '', ''].map(renderMessageSkeleton).join('')}
            </div>
            ${renderComposer()}
        `;
    }

    if (!state.activeChat) {
        return `
            <div class="empty-state">
                <div class="empty-icon">💬</div>
                <h2>Select a conversation to begin</h2>
                <p>Pick a chat from the left to view messages and continue the conversation.</p>
            </div>
        `;
    }

    return `
        ${renderChatHeader()}
        ${renderMessagesSection()}
        ${renderComposer()}
    `;
}

function renderChatHeader() {
    if (!state.activeChat) {
        return `
            <header class="chat-header">
                <div class="chat-info">
                    <h2>Inbox</h2>
                    <span>Choose a chat to see the thread</span>
                </div>
            </header>
        `;
    }

    const chat = state.activeChat;
    const network = chat.network ? chat.network.toUpperCase() : 'BEEPER';
    const lastActiveLabel = formatRelativeTime(chat.lastActivity);
    const status = chat.isArchived ? 'Archived conversation' : chat.isUnread ? 'Unread messages' : 'Connected';

    return `
        <header class="chat-header">
            <div class="chat-info">
                <h2>${escapeHtml(chat.title || chat.id)}</h2>
                <span>${network}${lastActiveLabel ? ` • Active ${lastActiveLabel}` : ''}</span>
            </div>
            <div class="chat-toolbar">
                <span class="presence-indicator" data-status="${chat.isArchived ? 'archived' : 'active'}"></span>
                <span class="presence-label">${escapeHtml(status)}</span>
                <div class="chat-toolbar-actions">
                    <button class="icon-button" data-action="open-details" title="Conversation details">ⓘ</button>
                    <button class="icon-button" data-action="toggle-pin" data-chat-id="${chat.id}" title="${chat.isPinned ? 'Unpin chat' : 'Pin chat'}">${chat.isPinned ? 'Unpin' : 'Pin'}</button>
                    <button class="icon-button" data-action="toggle-archive" data-chat-id="${chat.id}" title="${chat.isArchived ? 'Move to inbox' : 'Archive chat'}">${chat.isArchived ? 'Restore' : 'Archive'}</button>
                </div>
            </div>
        </header>
    `;
}

function renderChatCard(chat) {
    const chatId = normalizeChatId(chat.id);
    const initials = getInitials(chat.title || chatId);
    const palette = getAvatarPalette(chat.title || chatId);
    const timeLabel = formatRelativeTime(chat.lastActivity);

    return `
        <button type="button" class="chat-card${state.activeChat && normalizeChatId(state.activeChat.id) === chatId ? ' active' : ''}" data-chat-id="${chatId}" role="listitem">
            <div class="chat-avatar" style="background:${palette.bg}; color:${palette.fg}; box-shadow:${palette.shadow};">${escapeHtml(initials)}</div>
            <div class="chat-meta">
                <div class="chat-meta-heading">
                    <h3>${escapeHtml(chat.title || chatId)}</h3>
                    ${chat.unreadCount ? `<span class="chip chip-unread">${chat.unreadCount}</span>` : ''}
                </div>
                <span class="chat-subline">${(chat.network || '').toUpperCase()}${timeLabel ? ` • ${timeLabel}` : ''}</span>
            </div>
            <div class="chat-card-actions">
                ${chat.isPinned ? '<span class="chip chip-muted">Pinned</span>' : ''}
                ${chat.isArchived ? '<span class="chip chip-muted">Archived</span>' : ''}
                <button type="button" class="icon-button" data-action="toggle-pin" data-chat-id="${chatId}" title="${chat.isPinned ? 'Unpin chat' : 'Pin chat'}">${chat.isPinned ? 'Unpin' : 'Pin'}</button>
                <button type="button" class="icon-button" data-action="toggle-archive" data-chat-id="${chatId}" title="${chat.isArchived ? 'Move to inbox' : 'Archive chat'}">${chat.isArchived ? 'Restore' : 'Archive'}</button>
            </div>
        </button>
    `;
}

function renderMessagesSection() {
    if (!state.messages.length) {
        return `
            <div class="messages" id="messages">
                <div class="empty-state">
                    <div class="empty-icon">✨</div>
                    <p>No messages yet — say hello!</p>
                </div>
            </div>
        `;
    }

    return `
        <div class="messages" id="messages">
            ${renderMessageGroups(state.messages)}
        </div>
    `;
}

function renderComposer() {
    return `
        <form class="message-form" id="message-form">
            <div class="message-form-toolbar">
                <button type="button" class="icon-button" data-action="composer-attachment" title="Attach file">Attach</button>
                <button type="button" class="icon-button" data-action="composer-gif" title="Insert GIF">GIF</button>
                <button type="button" class="icon-button" data-action="composer-emoji" title="Insert emoji">Emoji</button>
            </div>
            <div class="message-input">
                <textarea id="message-input" placeholder="Write a message" rows="1" ${state.activeChat ? '' : 'disabled'} required></textarea>
            </div>
            <div class="composer-actions">
                <button class="icon-button" type="button" data-action="composer-more" title="More options">More</button>
                <button class="send-btn" type="submit" id="send-btn" ${state.activeChat ? '' : 'disabled'}>Send</button>
            </div>
        </form>
    `;
}

function renderMessageGroups(messages) {
    let markup = '';
    let currentDay = '';

    messages.forEach((message) => {
        const dayLabel = formatDayLabel(message.timestamp);
        if (dayLabel !== currentDay) {
            currentDay = dayLabel;
            markup += `<div class="message-divider"><span>${dayLabel}</span></div>`;
        }
        markup += renderMessage(message);
    });

    return markup;
}

function renderMessage(message) {
    const { senderName, text, timestamp, isSender } = message;
    const datetime = new Date(timestamp);
    const timeLabel = datetime.toLocaleTimeString([], {
        hour: '2-digit',
        minute: '2-digit',
    });

    return `
        <article class="message-bubble ${isSender ? 'outgoing' : 'incoming'}" data-message-id="${message.id}">
            <div class="message-meta">
                <strong>${escapeHtml(senderName)}</strong>
                <span>${timeLabel}</span>
            </div>
            <p class="message-text">${escapeHtml(text)}</p>
            <div class="message-actions" role="toolbar">
                <button type="button" class="icon-button" data-action="message-react" title="React">React</button>
                <button type="button" class="icon-button" data-action="message-reply" title="Reply">Reply</button>
                <button type="button" class="icon-button" data-action="message-more" title="More">More</button>
            </div>
        </article>
    `;
}

function renderMessageSkeleton() {
    return `
        <div class="message-skeleton">
            <div class="bubble"></div>
        </div>
    `;
}

function segmentChats(chats) {
    const pinned = [];
    const unread = [];
    const others = [];
    const archived = [];

    chats.forEach((chat) => {
        if (chat.isArchived) {
            archived.push(chat);
            return;
        }
        if (chat.isPinned) {
            pinned.push(chat);
        }
    });

    chats.forEach((chat) => {
        if (chat.isArchived || chat.isPinned) {
            return;
        }
        if (chat.isUnread) {
            unread.push(chat);
        } else {
            others.push(chat);
        }
    });

    return { pinned, unread, others, archived };
}

async function bootstrapData() {
    try {
        const [chatsResponse] = await Promise.all([
            api.chats(),
        ]);

        state.chats = (chatsResponse.chats || []).map(annotateChat);
        state.activeChat = state.chats.find((chat) => !chat.isArchived) || null;
        state.loading = false;
        renderApp();
        attachGlobalHandlers();

        if (state.activeChat) {
            await loadMessages();
        }
    } catch (error) {
        console.error('Failed to initialise Beeper Web:', error);
        showToast('Unable to load chats. Check your token and try again.');
    }
}

async function loadMessages() {
    if (!state.activeChat) return;
    state.loadingMessages = true;
    renderApp();
    try {
        const { messages } = await api.messages(state.activeChat.id);
        state.messages = (messages || []).map((message) => ({
            ...message,
            timestamp: message.timestamp,
        }));
    } catch (error) {
        console.error('Failed to load messages:', error);
        state.messages = [];
        showToast('Unable to load messages for this chat.');
    }
    state.loadingMessages = false;
    renderApp();
    scrollMessagesToBottom();
}

function attachGlobalHandlers() {
    if (eventsBound) {
        return;
    }
    eventsBound = true;

    document.addEventListener('click', async (event) => {
        if (event.target.closest('#logout-btn')) {
            await handleLogout();
            return;
        }

        const viewBtn = event.target.closest('[data-action="switch-view"]');
        if (viewBtn) {
            const targetView = viewBtn.dataset.view;
            if (targetView && state.sidebar.view !== targetView) {
                state.sidebar.view = targetView;
                persistSidebarView(targetView);
                renderApp();
            }
            return;
        }

        const chatCard = event.target.closest('.chat-card');
        if (chatCard) {
            const { chatId } = chatCard.dataset;
            if (!chatId || normalizeChatId(state.activeChat?.id) === chatId) {
                return;
            }
            const nextChat = findChatById(chatId);
            if (!nextChat) {
                showToast('Unable to open that chat right now.');
                return;
            }
            await selectChat(nextChat);
            return;
        }

        const pinBtn = event.target.closest('[data-action="toggle-pin"]');
        if (pinBtn) {
            const chatId = pinBtn.dataset.chatId;
            if (chatId) {
                togglePinned(chatId);
            }
            return;
        }

        const archiveBtn = event.target.closest('[data-action="toggle-archive"]');
        if (archiveBtn) {
            const chatId = archiveBtn.dataset.chatId;
            if (chatId) {
                toggleArchived(chatId);
            }
            return;
        }

        if (event.target.closest('[data-action="open-settings"]')) {
            showToast('Settings are not available in this demo yet.', 'info');
            return;
        }

        if (event.target.closest('[data-action^="composer-"]')) {
            showToast('This composer action is coming soon.', 'info');
            return;
        }

        if (event.target.closest('[data-action^="message-"]')) {
            showToast('Message actions will arrive in a future update.', 'info');
        }
    });

    document.addEventListener('submit', async (event) => {
        if (event.target.matches('#message-form')) {
            event.preventDefault();
            if (state.sending) return;

            const textarea = $('#message-input');
            const sendBtn = $('#send-btn');
            const text = textarea.value.trim();
            if (!text || !state.activeChat) return;

            state.sending = true;
            sendBtn.disabled = true;

            try {
                await api.sendMessage({ chatId: state.activeChat.id, text });
                textarea.value = '';
                await loadMessages();
            } catch (error) {
                console.error('Failed to send message:', error);
                showToast('Message could not be sent.');
            } finally {
                state.sending = false;
                sendBtn.disabled = false;
            }
        }
    });

    document.addEventListener('input', (event) => {
        if (event.target.matches('#chat-filter')) {
            state.chatFilter = event.target.value;
            renderApp();
        }
    });
}

function enableSendOnEnter() {
    const textarea = $('#message-input');
    if (!textarea) return;

    textarea.addEventListener('keydown', (event) => {
        if (event.key === 'Enter' && !event.shiftKey) {
            event.preventDefault();
            $('#message-form').requestSubmit();
        }
    });

    scrollMessagesToBottom();
}

function scrollMessagesToBottom() {
    const container = $('#messages');
    if (!container) return;
    requestAnimationFrame(() => {
        container.scrollTop = container.scrollHeight;
    });
}

function getFilteredChats() {
    let chats = state.chats;
    if (state.sidebar.view === VIEW_MODES.pinned) {
        chats = chats.filter((chat) => chat.isPinned && !chat.isArchived);
    } else if (state.sidebar.view === VIEW_MODES.unread) {
        chats = chats.filter((chat) => chat.isUnread && !chat.isArchived);
    } else if (state.sidebar.view === VIEW_MODES.archived) {
        chats = chats.filter((chat) => chat.isArchived);
    }

    if (!state.chatFilter) {
        return chats;
    }
    const query = state.chatFilter.toLowerCase();
    return chats.filter((chat) => {
        const target = `${chat.title || ''} ${chat.network || ''} ${chat.id || ''}`.toLowerCase();
        return target.includes(query);
    });
}

function findChatById(chatId) {
    const targetId = normalizeChatId(chatId);
    return state.chats.find((chat) => normalizeChatId(chat.id) === targetId) || null;
}

async function handleLogout() {
    try {
        await api.logout();
        Object.assign(state, {
            authenticated: false,
            user: null,
            chats: [],
            messages: [],
            activeChat: null,
            chatFilter: '',
        });
        renderApp();
    } catch (error) {
        console.error('Failed to log out:', error);
        showToast('Logout failed. Please try again.');
    }
}

function togglePinned(chatId) {
    const id = normalizeChatId(chatId);
    if (!id) return;

    if (state.pinnedChats.has(id)) {
        state.pinnedChats.delete(id);
        showToast('Chat removed from pinned.', 'info');
    } else {
        state.pinnedChats.add(id);
        showToast('Chat pinned for quick access.', 'success');
    }

    persistPreferenceSet(STORAGE_KEYS.pinned, state.pinnedChats);
    refreshChatAnnotations();
    renderApp();
}

function toggleArchived(chatId) {
    const id = normalizeChatId(chatId);
    if (!id) return;

    if (state.archivedChats.has(id)) {
        state.archivedChats.delete(id);
        showToast('Chat restored to inbox.', 'info');
    } else {
        state.archivedChats.add(id);
        showToast('Chat archived.', 'info');
    }

    persistPreferenceSet(STORAGE_KEYS.archived, state.archivedChats);
    refreshChatAnnotations();
    if (state.activeChat && normalizeChatId(state.activeChat.id) === id && state.archivedChats.has(id)) {
        state.activeChat = null;
        state.messages = [];
    }
    renderApp();
}

function refreshChatAnnotations() {
    state.chats = state.chats.map(annotateChat);
    syncActiveChat();
}

function annotateChat(chat) {
    const id = normalizeChatId(chat.id);
    const unreadCount = typeof chat.unreadCount === 'number' ? chat.unreadCount : 0;
    return {
        ...chat,
        id,
        isPinned: state.pinnedChats.has(id),
        isArchived: state.archivedChats.has(id),
        isUnread: unreadCount > 0,
    };
}

function syncActiveChat() {
    if (!state.activeChat) return;
    const activeId = normalizeChatId(state.activeChat.id);
    const next = state.chats.find((chat) => normalizeChatId(chat.id) === activeId);
    if (next) {
        state.activeChat = next;
    } else {
        state.activeChat = null;
        state.messages = [];
    }
}

function normalizeChatId(value) {
    return String(value ?? '').trim();
}

function getInitials(label) {
    if (!label) return 'B';
    const words = label.trim().split(/\s+/).filter(Boolean);
    if (words.length === 0) return 'B';
    if (words.length === 1) {
        return words[0].slice(0, 2).toUpperCase();
    }
    return `${words[0][0] || ''}${words[1][0] || ''}`.toUpperCase();
}

function getAvatarPalette(seed) {
    const palettes = [
        { bg: 'linear-gradient(135deg, #4f6ef7, #415be4)', fg: '#f5f7ff', shadow: '0 8px 20px rgba(66, 94, 226, 0.22)' },
        { bg: 'linear-gradient(135deg, #f97373, #f0527c)', fg: '#fff6f5', shadow: '0 8px 18px rgba(240, 82, 124, 0.22)' },
        { bg: 'linear-gradient(135deg, #38bdf8, #3a82f6)', fg: '#0b1a32', shadow: '0 8px 18px rgba(58, 130, 246, 0.22)' },
        { bg: 'linear-gradient(135deg, #facc15, #fb8c2f)', fg: '#3b2f01', shadow: '0 8px 18px rgba(250, 140, 47, 0.22)' },
        { bg: 'linear-gradient(135deg, #34d399, #14b88a)', fg: '#052e16', shadow: '0 8px 18px rgba(20, 184, 138, 0.22)' },
        { bg: 'linear-gradient(135deg, #5aa6ff, #5d7dff)', fg: '#0b1220', shadow: '0 8px 18px rgba(93, 125, 255, 0.22)' },
    ];
    const index = Math.abs(hashString(seed || 'beeper')) % palettes.length;
    return palettes[index];
}

function hashString(value) {
    let hash = 0;
    for (let i = 0; i < value.length; i += 1) {
        hash = ((hash << 5) - hash) + value.charCodeAt(i);
        hash |= 0;
    }
    return hash;
}

function formatWorkspaceLabel(subject) {
    if (!subject) return 'Connected';
    if (subject.length <= 18) return subject;
    return `${subject.slice(0, 6)}…${subject.slice(-4)}`;
}

function formatRelativeTime(value) {
    if (!value) return '';
    const date = new Date(value);
    if (Number.isNaN(date.getTime())) return '';

    const diff = Date.now() - date.getTime();
    const seconds = Math.round(diff / 1000);
    if (seconds < 45) return 'just now';
    const minutes = Math.round(seconds / 60);
    if (minutes < 60) return `${minutes}m ago`;
    const hours = Math.round(minutes / 60);
    if (hours < 24) return `${hours}h ago`;
    const days = Math.round(hours / 24);
    if (days < 7) return `${days}d ago`;
    return date.toLocaleDateString();
}

function formatDayLabel(value) {
    const date = new Date(value);
    if (Number.isNaN(date.getTime())) return '';

    const today = new Date();
    today.setHours(0, 0, 0, 0);

    const target = new Date(date);
    target.setHours(0, 0, 0, 0);

    const diffMs = today - target;
    const diffDays = Math.round(diffMs / (24 * 60 * 60 * 1000));

    if (diffDays === 0) return 'Today';
    if (diffDays === 1) return 'Yesterday';
    return date.toLocaleDateString(undefined, { weekday: 'long', month: 'short', day: 'numeric' });
}

function showToast(message, tone = 'error') {
    clearTimeout(toastTimeout);
    state.toast = { message, tone };
    renderApp();
    toastTimeout = setTimeout(() => {
        state.toast = null;
        renderApp();
    }, 4200);
}

function renderToast(toast) {
    const elem = document.createElement('div');
    elem.className = `toast toast-${toast.tone}`;
    elem.id = 'global-toast';
    elem.innerHTML = `
        <span>${escapeHtml(toast.message)}</span>
        <button type="button" aria-label="Dismiss" class="toast-dismiss">✕</button>
    `;

    elem.querySelector('.toast-dismiss').addEventListener('click', () => {
        clearTimeout(toastTimeout);
        state.toast = null;
        renderApp();
    });

    return elem;
}

function scheduleToastDismissal() {
    const toast = $('#global-toast');
    if (!toast) return;
    requestAnimationFrame(() => {
        toast.classList.add('is-visible');
    });
}

function loadPreferences() {
    const pinned = readPreferenceSet(STORAGE_KEYS.pinned);
    const archived = readPreferenceSet(STORAGE_KEYS.archived);
    const view = readSidebarView();
    return {
        pinned,
        archived,
        view,
    };
}

function readPreferenceSet(key) {
    try {
        const value = window.localStorage.getItem(key);
        if (!value) return new Set();
        const parsed = JSON.parse(value);
        if (Array.isArray(parsed)) {
            return new Set(parsed.map(normalizeChatId));
        }
        return new Set();
    } catch (error) {
        console.warn('Failed to read preference set', key, error);
        return new Set();
    }
}

function persistPreferenceSet(key, set) {
    try {
        const value = JSON.stringify(Array.from(set));
        window.localStorage.setItem(key, value);
    } catch (error) {
        console.warn('Failed to persist preference', key, error);
    }
}

function readSidebarView() {
    try {
        const value = window.localStorage.getItem(STORAGE_KEYS.sidebarView);
        if (value && Object.values(VIEW_MODES).includes(value)) {
            return value;
        }
    } catch (error) {
        console.warn('Failed to read sidebar view preference', error);
    }
    return VIEW_MODES.all;
}

function persistSidebarView(view) {
    try {
        window.localStorage.setItem(STORAGE_KEYS.sidebarView, view);
    } catch (error) {
        console.warn('Failed to persist sidebar view', error);
    }
}

async function selectChat(chat) {
    state.activeChat = chat;
    await loadMessages();
}

function updatePageTitle() {
    if (state.activeChat) {
        const title = state.activeChat.title || state.activeChat.id;
        document.title = `${title} • Beeper Web`;
    } else {
        document.title = 'Beeper • Web Chat';
    }
}

function escapeHtml(input = '') {
    return input
        .replace(/&/g, '&amp;')
        .replace(/</g, '&lt;')
        .replace(/>/g, '&gt;')
        .replace(/"/g, '&quot;')
        .replace(/'/g, '&#039;');
}

(async function init() {
    try {
        const session = await api.session();
        state.authenticated = session.authenticated;
        state.user = session.user || null;
        state.loading = false;

        if (state.authenticated) {
            await bootstrapData();
        } else {
            renderApp();
        }
    } catch (error) {
        console.error('Failed to bootstrap application:', error);
        state.loading = false;
        state.authenticated = false;
        renderApp();
    }
})();
