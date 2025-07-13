import { timeAgo } from "../tools.js";

// Global variables for notification badge and chat management
let unreadCount = 0;
let updateNotifBadge;
let chatUsers = []; // Store all chat users for sorting

export const Chat = {
  html: `
    <head>
      <link rel="stylesheet" href="/public/styles/chat.css">
    </head>
    <span id="chat-btn" class="material-symbols-outlined">
      chat_bubble
      <span id="chat-notif" class="notif-badge hidden"></span>
    </span>
    <div id="chat-popup" class="hidden">
      <div class="sidebar">
        <div class="header">
          Recent
          <span class="online-count" id="online-count"></span>
        </div>
        <div id="chat-list"></div>
      </div>
      <div class="main">
        <h3>Start a conversation</h3>
        <p class="chat-instructions">Select a user from the list to start chatting</p>
      </div>
    </div>
  `,
  setup: async () => {
    const btn = document.getElementById("chat-btn");
    const popup = document.getElementById("chat-popup");
    const main = popup.querySelector(".main");

    window.ws = initWebSocket();
    window.currentConversationId = 0;
    window.chatMain = main;
    window.currentUser = await getCurrentUser();

    updateNotifBadge = () => {
      const badge = document.getElementById("chat-notif");
      if (!badge) return;
      if (unreadCount > 0) {
        badge.textContent = unreadCount > 9 ? "9+" : unreadCount;
        badge.classList.remove("hidden");
      } else {
        badge.classList.add("hidden");
      }
    };

    btn.addEventListener("click", () => {
      popup.classList.toggle("hidden");
      if (!popup.classList.contains("hidden")) {
        unreadCount = 0;
        updateNotifBadge();
        // Refresh chat list when opened
        loadRecentUsers();
      }
    });

    document.addEventListener("click", (e) => {
      const target = e.target;

      // Don't hide if the click is inside the popup or on the button
      if (popup.contains(target) || btn.contains(target)) return;

      // Also don't hide if the click is on a child of popup (like user item)
      const closestChatItem = target.closest(".chat-item");
      if (closestChatItem && popup.contains(closestChatItem)) return;

      popup.classList.add("hidden");
    });

    await loadRecentUsers();
  }
};

const getCurrentUser = async () => {
  try {
    const res = await fetch("/me", {
      method: "POST",
      headers: { "Content-Type": "application/json" }
    });
    if (res.ok) return await res.json();
    return null;
  } catch (err) {
    console.error("Error getting current user:", err);
    return null;
  }
};

const initWebSocket = () => {
  if (window.ws && window.ws.readyState === WebSocket.OPEN) return window.ws;

  const ws = new WebSocket(`ws://${location.host}/ws`);
  window.ws = ws;

  ws.onopen = () => {
    console.log("✅ WebSocket connected");
    document.body.classList.remove("ws-disconnected");
    updateOnlineStatus();
  };

  ws.onclose = () => {
    console.log("❌ WebSocket closed");
    document.body.classList.add("ws-disconnected");
    updateOnlineStatus();
    setTimeout(initWebSocket, 3000);
  };

  ws.onerror = (err) => {
    console.error("⚠️ WebSocket error:", err);
  };

  ws.onmessage = (event) => {
    try {
      const msg = JSON.parse(event.data);
      console.log("📥 WS message received:", msg);

      switch (msg.type) {
        case 'typing':
          handleTypingMessage(msg);
          break;
        case 'message':
          handleIncomingMessage(msg);
          break;
        case 'ack':
          handleAckMessage(msg);
          break;
        case 'user_status':
          handleUserStatusChange(msg);
          break;
        default:
          if (msg.content && msg.author_id) {
            console.log("🔄 Treating message without type as regular message");
            msg.type = 'message';
            handleIncomingMessage(msg);
          } else {
            console.warn("⚠️ Unknown WS message type:", msg);
          }
      }
    } catch (err) {
      console.error("❌ Failed to parse WS message:", err);
    }
  };

  return ws;
};

// Load recent chat users with proper sorting
const loadRecentUsers = async () => {
  try {
    const res = await fetch("/recent", { method: "POST" });
    const users = await res.json();
    console.log("🧾 Recent users:", users);

    // Store users globally for sorting
    chatUsers = users || [];

    // Sort users before rendering
    sortChatUsers();
    renderChatList();

  } catch (err) {
    console.error("Error loading chats:", err);
    document.getElementById("chat-list").innerHTML = `<div class="error">Unable to load conversations</div>`;
  }
};

// Sort chat users by last message time and online status
const sortChatUsers = () => {
  chatUsers.sort((a, b) => {
    // First, prioritize online users
    const aOnline = a.is_online || false;
    const bOnline = b.is_online || false;

    if (aOnline !== bOnline) {
      return bOnline ? 1 : -1; // Online users first
    }

    // Then sort by last message time
    const aTime = getLastMessageTime(a);
    const bTime = getLastMessageTime(b);

    if (aTime && bTime) {
      return new Date(bTime) - new Date(aTime); // Most recent first
    }

    if (aTime && !bTime) return -1; // Users with messages first
    if (!aTime && bTime) return 1;

    // Finally, sort alphabetically by name
    const aName = `${a.first_name} ${a.last_name}`.toLowerCase();
    const bName = `${b.first_name} ${b.last_name}`.toLowerCase();
    return aName.localeCompare(bName);
  });
};

// Helper function to get last message time
const getLastMessageTime = (user) => {
  if (typeof user.last_message_at === "string" && user.last_message_at) {
    return user.last_message_at;
  }
  if (user.last_message_at?.Valid && user.last_message_at?.Time) {
    return user.last_message_at.Time;
  }
  return null;
};

// Render the sorted chat list
const renderChatList = () => {
  const list = document.getElementById("chat-list");
  list.innerHTML = "";

  if (!chatUsers || chatUsers.length === 0) {
    list.innerHTML = `<div class="no-users">No recent conversations</div>`;
    return;
  }

  // Create sections for online and offline users
  const onlineUsers = chatUsers.filter(user => user.is_online);
  const offlineUsers = chatUsers.filter(user => !user.is_online);

  // Update online count
  updateOnlineCount(onlineUsers.length);

  // Render online users
  if (onlineUsers.length > 0) {
    const onlineHeader = document.createElement("div");
    onlineHeader.className = "chat-section-header";
    onlineHeader.innerHTML = `<span class="online-indicator"></span> Online (${onlineUsers.length})`;
    list.appendChild(onlineHeader);

    onlineUsers.forEach(user => {
      const chatItem = createChatItem(user);
      chatItem.addEventListener("click", () => openChat(user));
      list.appendChild(chatItem);
    });
  }

  // Render offline users
  if (offlineUsers.length > 0) {
    const offlineHeader = document.createElement("div");
    offlineHeader.className = "chat-section-header";
    offlineHeader.innerHTML = `<span class="offline-indicator"></span> Offline (${offlineUsers.length})`;
    list.appendChild(offlineHeader);

    offlineUsers.forEach(user => {
      const chatItem = createChatItem(user);
      chatItem.addEventListener("click", () => openChat(user));
      list.appendChild(chatItem);
    });
  }
};

// Build a single chat item with enhanced features
const createChatItem = (user) => {
  const item = document.createElement("div");
  item.className = "chat-item";
  item.setAttribute("data-user-id", user.id);

  // Add unread indicator if there are unread messages
  if (user.unread_count > 0) {
    item.classList.add("has-unread");
  }

  const avatar = document.createElement("div");
  avatar.className = "avatar";
  avatar.style.backgroundImage = `url('${user.profile_img || ""}')`;

  // Add online status indicator
  const statusIndicator = document.createElement("div");
  statusIndicator.className = `status-indicator ${user.is_online ? 'online' : 'offline'}`;
  avatar.appendChild(statusIndicator);

  const chatInfo = document.createElement("div");
  chatInfo.className = "chat-info";

  // Add last message preview
  const lastMessagePreview = user.last_message_content
    ? `<div class="last-message">${truncateMessage(user.last_message_content)}</div>`
    : '';

  chatInfo.innerHTML = `
    <div class="user-name">
      <strong>${user.first_name} ${user.last_name}</strong>
      ${user.unread_count > 0 ? `<span class="unread-badge">${user.unread_count}</span>` : ''}
    </div>
    <small class="username">@${user.username || ""}</small>
    ${lastMessagePreview}
  `;

  const time = document.createElement("div");
  time.className = "chat-time";
  const lastMessageTime = getLastMessageTime(user);
  if (lastMessageTime) {
    time.textContent = timeAgo(lastMessageTime);
  } else {
    time.textContent = "";
  }

  item.appendChild(avatar);
  item.appendChild(chatInfo);
  item.appendChild(time);

  return item;
};

// Update a specific user in the chat list (for real-time updates)
const updateUserInChatList = (userId, updates) => {
  const userIndex = chatUsers.findIndex(user => user.id === userId);
  if (userIndex !== -1) {
    // Update user data
    chatUsers[userIndex] = { ...chatUsers[userIndex], ...updates };

    // Re-sort and re-render
    sortChatUsers();
    renderChatList();

    // Maintain active state if this user is currently selected
    if (window.currentConversationId && updates.conversation_id === window.currentConversationId) {
      document.querySelector(`[data-user-id="${userId}"]`)?.classList.add("active");
    }
  }
};

// Handle incoming messages and update chat list
const handleIncomingMessage = (msg) => {
  const messagesContainer = document.getElementById("messages");
  const isCurrentConv = msg.conversation_id?.Int64 === window.currentConversationId?.Int64;

  if (messagesContainer && isCurrentConv) {
    // Remove any existing typing indicators
    const typing = messagesContainer.querySelector(".typing");
    if (typing) typing.remove();

    const bubble = document.createElement("div");
    const isOutgoing = msg.author_id === window.currentUser?.id;
    bubble.className = `bubble ${isOutgoing ? "outgoing" : "incoming"}`;
    bubble.textContent = msg.content;

    const timestamp = document.createElement("div");
    timestamp.className = "message-time";
    timestamp.textContent = "just now";
    bubble.appendChild(timestamp);

    messagesContainer.appendChild(bubble);
    messagesContainer.scrollTop = messagesContainer.scrollHeight;
  } else {
    // Message is for a different conversation
    if (document.getElementById("chat-popup").classList.contains("hidden")) {
      unreadCount++;
      updateNotifBadge();
    }
  }

  // Update the chat list with new message info
  const senderId = msg.author_id;
  const senderUser = chatUsers.find(user => user.id === senderId);

  if (senderUser) {
    const updates = {
      last_message_at: new Date().toISOString(),
      last_message_content: msg.content,
      unread_count: isCurrentConv ? 0 : (senderUser.unread_count || 0) + 1
    };

    updateUserInChatList(senderId, updates);
  }
};

// Handle user status changes (online/offline)
const handleUserStatusChange = (user) => {


  updateUserInChatList(user.ID, { is_online: user.isOnline });
  console.log(`👤 User ${user.ID} is now ${user.isOnline ? 'online' : 'offline'}`);
};

// Handle typing indicators
const handleTypingMessage = (msg) => {
  const messagesContainer = document.getElementById("messages");
  if (compareConversationIds(msg.conversation_id, window.currentConversationId) && messagesContainer) {
    let typing = messagesContainer.querySelector(".typing");
    if (!typing) {
      typing = document.createElement("div");
      typing.className = "typing";
      typing.innerHTML = `<div class="typing-indicator"><span></span><span></span><span></span></div>`;
      messagesContainer.appendChild(typing);
    }
    clearTimeout(typing._timeout);
    typing._timeout = setTimeout(() => typing.remove(), 3000);
  }
};

// Handle acknowledgment messages
const handleAckMessage = (msg) => {
  document.querySelectorAll(".bubble.outgoing.pending").forEach(bubble => {
    if (bubble.dataset.tempid === String(msg.temp_id)) {
      window.currentConversationId = msg.conversation_id
      bubble.classList.remove("pending");
      bubble.classList.add("delivered");
      bubble.title = "Delivered";
    }
  });
};

// Helper functions
const compareConversationIds = (id1, id2) => {
  const getId = (id) => {
    if (typeof id === 'object' && id !== null) {
      return id.Int64 || id.int64 || id.ID || id.id;
    }
    return id;
  };
  return getId(id1) === getId(id2);
};

const truncateMessage = (message, maxLength = 30) => {
  if (message.length <= maxLength) return message;
  return message.substring(0, maxLength) + '...';
};

const updateOnlineCount = (count) => {
  const onlineCountEl = document.getElementById("online-count");
  if (onlineCountEl) {
    onlineCountEl.textContent = count > 0 ? `(${count})` : '';
  }
};

const updateOnlineStatus = () => {
  const isConnected = window.ws?.readyState === WebSocket.OPEN;
  document.body.classList.toggle("ws-disconnected", !isConnected);
};

// Open chat window for selected user
const openChat = async (user) => {
  console.log("📂 Opening chat with:", user);

  // Clear unread count for this user
  updateUserInChatList(user.id, { unread_count: 0 });

  document.querySelectorAll(".chat-item.active").forEach(el =>
    el.classList.remove("active")
  );
  document.querySelector(`[data-user-id="${user.id}"]`)?.classList.add("active");

  const main = window.chatMain;
  const hasConversation = user.conversation_id != 0 && user.conversation_id != null;
  window.currentConversationId = user.conversation_id || 0;

  main.innerHTML = `
    <div class="conv-header">
      <div class="user-info">
        <div class="avatar" style="background-image: url('${user.profile_img || ''}')">
          <div class="status-indicator ${user.is_online ? 'online' : 'offline'}"></div>
        </div>
        <div>
          <strong>${user.first_name} ${user.last_name}</strong>
          <small>@${user.username || ''}</small>
          <div class="user-status">${user.is_online ? 'Online' : 'Offline'}</div>
        </div>
      </div>
    </div>
    <button id="load-more" style="display: none;">Load more</button>
    <div id="messages" class="messages"></div>
    <div id="chat-form">
      <input id="msg-input" type="text" placeholder="Type a message..." required />
      <button>Send</button>
    </div>
  `;

  if (!hasConversation) {
    main.querySelector("#messages").innerHTML = `<div class="no-chat">No conversation yet. Send a message to start!</div>`;
    setupMessageForm(user);
    return;
  }

  await loadMessages(user, -1);
  setupMessageForm(user);

  unreadCount = 0;
  if (updateNotifBadge) updateNotifBadge();
};

// Load chat messages from server
const loadMessages = async (user, startID = -1) => {
  console.log(`🔄 Loading messages for conversation ${window.currentConversationId}, startID=${startID}`);
  const messagesContainer = window.chatMain.querySelector("#messages");
  const loadMoreBtn = window.chatMain.querySelector("#load-more");

  try {
    const res = await fetch("/conversation", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({
        conversation_id: window.currentConversationId,
        start_id: startID,
        n_message: 10 // Load 10 messages as specified in requirements
      }),
    });

    const messages = await res.json();
    if (!messages || messages.length === 0) {
      loadMoreBtn.style.display = "none";
      return;
    }

    const wasAtBottom = messagesContainer.scrollTop + messagesContainer.clientHeight >= messagesContainer.scrollHeight - 1;
    const oldScrollHeight = messagesContainer.scrollHeight;
    const frag = document.createDocumentFragment();

    messages.reverse().forEach(msg => {
      const bubble = document.createElement("div");
      bubble.className = `bubble ${msg.is_outgoing ? "outgoing" : "incoming"}`;
      bubble.textContent = msg.content;
      if (msg.id) bubble.dataset.id = msg.id;

      const timestamp = document.createElement("div");
      timestamp.className = "message-time";
      timestamp.textContent = timeAgo(msg.sent_at);
      bubble.appendChild(timestamp);

      frag.appendChild(bubble);
    });

    if (startID === -1) {
      messagesContainer.appendChild(frag);
      messagesContainer.scrollTop = messagesContainer.scrollHeight;
    } else {
      messagesContainer.prepend(frag);
      messagesContainer.scrollTop = messagesContainer.scrollHeight - oldScrollHeight;
    }

    loadMoreBtn.style.display = "block";
    loadMoreBtn.onclick = () => {
      const firstBubble = messagesContainer.querySelector(".bubble");
      if (firstBubble && firstBubble.dataset.id) {
        loadMessages(user, parseInt(firstBubble.dataset.id));
      }
    };

    // Mark messages as seen
    fetch("/mark-seen", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ conversation_id: window.currentConversationId })
    });

  } catch (err) {
    console.error("Error loading messages:", err);
    messagesContainer.innerHTML = `<div class="error">Failed to load messages</div>`;
  }
};

// Setup message input form and sending logic
const setupMessageForm = (user) => {
  const form = document.getElementById("chat-form");
  const input = document.getElementById("msg-input");
  const messagesContainer = document.getElementById("messages");

  if (!form || !input) return;

  input.addEventListener("input", () => {
    console.log('typing');
    
    if (window.ws?.readyState === WebSocket.OPEN) {
      window.ws.send(JSON.stringify({
        type: "typing",
        conversation_id: window.currentConversationId
      }));
    }
  });

  const sendBtn = form.querySelector("button");
  input.addEventListener("keydown", (e) => {
    if (e.key === "Enter" && !e.shiftKey) {
      sendBtn.click();
      e.preventDefault();
    }
  });

  sendBtn.addEventListener("click", async (e) => {
    e.preventDefault();
    const content = input.value.trim();
    if (!content) return;

    input.value = "";
    const tempId = Date.now();

    const msg = {
      type: "message",
      content,
      conversation_id: window.currentConversationId,
      reciever_id: user.id,
      temp_id: tempId,
    };

    // Create pending message bubble
    const bubble = document.createElement("div");
    bubble.className = window.ws?.readyState === WebSocket.OPEN
      ? "bubble outgoing pending"
      : "bubble outgoing error";
    bubble.dataset.tempid = tempId;
    bubble.textContent = content;
    bubble.title = window.ws?.readyState === WebSocket.OPEN ? "Sending..." : "Failed to send";

    const timestamp = document.createElement("div");
    timestamp.className = "message-time";
    timestamp.textContent = window.ws?.readyState === WebSocket.OPEN ? "sending..." : "failed";
    bubble.appendChild(timestamp);

    messagesContainer.appendChild(bubble);
    messagesContainer.scrollTop = messagesContainer.scrollHeight;

    // Send message via WebSocket
    if (window.ws?.readyState === WebSocket.OPEN) {
      console.log("📤 Sending message:", msg);
      window.ws.send(JSON.stringify(msg));

      // Update chat list with new message
      updateUserInChatList(user.id, {
        last_message_at: new Date().toISOString(),
        last_message_content: content
      });
    } else {
      console.warn("⚠️ WebSocket is not open. Message not sent.");
    }
  });
};