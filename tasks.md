## ✅ Task Breakdown

### 🔐 Authentication (Login & Registration)
**Backend**
- [done] Create SQLite tables for users
- [testing] Handle registration: hash passwords, store user
- [testing] Handle login: validate credentials, sessions via cookies
- [testing] Middleware to check if a user is authenticated

**Frontend**
- [ ] Registration/login form handling in JS
- [ ] Store session info in cookies (send to backend)
- [ ] Toggle between login/register form and forum UI

---

### 📬 Private Messaging
**Backend**
- [testing] WebSocket hub (gorilla/websocket) to manage connections
- [testing] Store and retrieve messages in SQLite (pagination for old messages)
- [testing] Message broadcasting logic to specific users

**Frontend**
- [ ] WebSocket client for private chat
- [ ] Show online/offline users in a sidebar (sorted by last msg or name)
- [ ] Load last 10 messages when opening a chat
- [ ] Load 10 older messages on scroll (with debounce/throttle)

---

### 🧵 Posts & Comments
**Backend**
- [done] Tables for posts, categories, comments
- [switch: querying into decoding payload] APIs to create/fetch posts
- [switch: querying into decoding payload] APIs to create/fetch comments (only when viewing a post)

**Frontend**
- [ ] Render feed of posts
- [ ] Click to expand a post and view comments
- [ ] Form to submit new post/comment

---

### 🌐 Real-Time Features
- [ ] Post/comment notification system (can use websockets or periodic polling)
- [ ] Real-time messaging in private chat
- [ ] Mark users as online/offline based on websocket presence

---

### 🎨 UI & SPA Navigation
**Single HTML File:**
- [ ] func: Register
- [ ] func: Login
- [ ] func: Post feed
- [ ] func: Chat sidebar
- [ ] func: Chatbox
- [ ] func: Create post

**Navigation:**
- [ ] Js router
- [ ] Render

---

## 🛠 Suggested Development Flow

1. Setup DB schema and user auth first [done]
2. Get login/register working with sessions [pending]
3. Build post feed and comment system
4. Implement private messaging with WebSockets
5. Finalize frontend navigation and UI polish

---

Want a sample DB schema next, or want to start from the WebSocket hub?