## Sync vs Async — Reddit Clone

---

### Users

| Operation | Type | Why |
|---|---|---|
| Login / Register | Sync | User waits for token and session creation |
| Get / Update profile | Sync | Immediate read/write, simple request-response flow |
| Send verification email | Async | Background email job, no need to block user |
| Karma update | Async | Triggered by events, processed later without blocking |

---

### Feed

| Operation | Type | Why |
|---|---|---|
| Load feed | Sync | User needs immediate posts response from system |
| Submit a vote | Sync | Simple DB write, return success instantly |
| Not interested signal | Sync | Lightweight update, no background processing needed |

---

### Search

| Operation | Type | Why |
|---|---|---|
| Search query | Sync | User expects immediate search results response |
| Index post on creation | Async | Background indexing after post successfully stored |
| Remove post from index | Async | Cleanup handled later, not blocking deletion |

---

### Upload

| Operation | Type | Why |
|---|---|---|
| Create text post | Sync | Simple write operation, immediate confirmation needed |
| Create image post | Sync | Store metadata and return response quickly |
| Image processing | Async | Processing can run later without blocking request |
| Video transcoding | Async | Heavy long-running job, must run in background |
| Content purge on delete | Async | Delayed cleanup, not required immediately |

---

### Video

| Operation | Type | Why |
|---|---|---|
| Receive video upload | Sync | Accept file and return job reference immediately |
| Transcode variants | Async | CPU-heavy processing done in background workers |
| Get transcoding status | Sync | Client polls and receives current status immediately |
| Status push (WebSocket) | Async | Real-time updates without blocking any requests |

---

### Notifications

| Operation | Type | Why |
|---|---|---|
| Mark as read | Sync | Immediate update of notification state in database |
| Deliver to online user | Async | Push via FCM without blocking main flow |
| Queue for offline user | Async | Store temporarily and deliver on reconnect |
| Send email notification | Async | Background email worker handles delivery |

---

### Chat

| Operation | Type | Why |
|---|---|---|
| Send message | Sync | Save message and confirm instantly to sender |
| Deliver to recipient | Async | Push message without blocking sender request |
| Fetch missed messages | Sync | Client requests and receives messages immediately |

---

## Rule

Synchronous for user requests, asynchronous for background and delivery tasks.