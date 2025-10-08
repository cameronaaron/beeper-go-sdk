# Beeper Web Chat

A sleek web client that uses the Go SDK to deliver a polished Beeper chat experience directly in the browser.

## Features

- 🔐 Secure login with your own Beeper Desktop API access token
- 💬 Real-time style chat view with unread indicators and rich layout
- 🚀 Streamlined message composer with keyboard shortcuts
- 📱 Responsive design that adapts to tablets and smaller screens

## Getting Started

```bash
# Run from the repository root
 go run ./cmd/webchat
```

The server listens on `:8080` by default. Override the port with the `PORT` environment variable:

```bash
PORT=9090 go run ./cmd/webchat
```

Then open `http://localhost:8080` in your browser. Enter your Beeper Desktop API access token (and an optional base URL if you run the API somewhere other than `http://localhost:23373`).

## Implementation Notes

- Sessions are stored in-memory and tied to a secure cookie. The cookie is not marked `Secure` so HTTP works locally—enable TLS and the secure flag in production.
- The backend relies on the SDK’s resource clients (`Chats`, `Messages`, `Token`) for all data access.
- Static assets live in `cmd/webchat/ui/` and are embedded via `go:embed`, so the binary is self-contained.
- All outbound SDK requests carry a 10s timeout to keep the interface responsive.

## Future Enhancements

- Websocket/Server-Sent Events bridge for live updates
- Message reactions and attachments
- Multi-account switching and richer filtering
