package handlers

import (
	"database/sql"
	"log"
	"net/http"
	"sync"
	"time"

	"echohub/models"

	"github.com/gorilla/websocket"
)

type WSHub struct {
	Upgrader  websocket.Upgrader
	Clients   map[*websocket.Conn]*models.User
	Broadcast chan models.Message
	Lock      sync.Mutex
}

func (app *WebApp) HTTPtoWS(w http.ResponseWriter, r *http.Request) {
	wsConn, err := app.Hub.Upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("WebSocket upgrade error:", err)
		return
	}
	defer wsConn.Close()

	user, ok := r.Context().Value(contextKeyUser).(*models.User)
	if !ok {
		log.Println("Unauthorized WebSocket connection attempt")
		wsConn.WriteJSON(map[string]string{"error": "Unauthorized"})
		return
	}

	app.Hub.Lock.Lock()
	app.Hub.Clients[wsConn] = user
	app.Hub.Lock.Unlock()

	log.Printf("✅ WebSocket connected: User ID %d\n", user.ID)

	defer func() {
		app.Hub.Lock.Lock()
		delete(app.Hub.Clients, wsConn)
		app.Hub.Lock.Unlock()
		log.Printf("❌ WebSocket disconnected: User ID %d\n", user.ID)
	}()

	for {
		var message models.Message
		if err := wsConn.ReadJSON(&message); err != nil {
			log.Println("⚠️ Read message error:", err)
			break
		}

		log.Printf("📩 Received message from user %d: %+v\n", user.ID, message)
		message.AuthorID = user.ID

		switch message.Type {
		case "message":
			if err := app.handleChatMessage(wsConn, user, &message); err != nil {
				log.Printf("❌ Error handling chat message: %v", err)
				break
			}
		case "typing":
			app.handleTypingMessage(user, &message)
		default:
			log.Printf("⚠️ Unknown message type: %s", message.Type)
		}
	}
}

func (app *WebApp) BroadcastMessages() {
	for msg := range app.Hub.Broadcast {
		log.Printf("📡 Broadcasting message to clients: %+v\n", msg)

		app.Hub.Lock.Lock()
		for client, user := range app.Hub.Clients {
			shouldSend := false

			switch msg.Type {
			case "message":
				// Send to receiver only (not back to sender)
				shouldSend = user.ID == msg.RecieverID
			case "typing":
				// Send typing indicators to all participants except sender
				shouldSend = user.ID != msg.AuthorID 
			}

			if shouldSend {
				log.Printf("➡️ Sending %s to user %d\n", msg.Type, user.ID)
				if err := client.WriteJSON(msg); err != nil {
					log.Printf("❌ Write error to user %d: %v\n", user.ID, err)
					client.Close()
					delete(app.Hub.Clients, client)
				}
			} else {
				log.Printf("⏭️ Skipping user %d (not target for %s)\n", user.ID, msg.Type)
			}
		}
		app.Hub.Lock.Unlock()
	}
}

func (app *WebApp) handleChatMessage(wsConn *websocket.Conn, user *models.User, message *models.Message) error {
	// Get or create conversation
	conv, err := app.Conversations.GetConversation(user.ID, message.RecieverID)
	if err != nil {
		log.Println("❌ Error checking conversation:", err)
		return err
	}

	if conv == nil {
		log.Printf("🆕 Creating new conversation between %d and %d\n", user.ID, message.RecieverID)
		convID, err := app.Conversations.InsertConversation(user.ID, message.RecieverID)
		if err != nil {
			log.Println("❌ Failed to create conversation:", err)
			return err
		}
		message.ConversationID.Int64 = int64(convID)
		message.ConversationID.Valid = true
	} else {
		log.Printf("🔁 Using existing conversation ID %d\n", conv.ID)
		message.ConversationID.Int64 = int64(conv.ID)
		message.ConversationID.Valid = true
	}

	// Set timestamp
	message.SentAt = time.Now()

	// Insert message into database
	if err := app.Messages.InsertMessage(*message); err != nil {
		log.Println("❌ Failed to insert message:", err)
		return err
	}

	log.Printf("✅ Message inserted in DB: %+v\n", message)

	// Send ACK to sender
	ack := models.Message{
		Type:    "ack",
		ConversationID: sql.NullInt64{
			Int64: int64(conv.ID),
			Valid: true,
		},
		TempID:  message.TempID,
		Content: "Message delivered",
	}
	if err := wsConn.WriteJSON(ack); err != nil {
		log.Printf("❌ Failed to send ACK to user %d: %v", user.ID, err)
	}

	// Update conversation timestamp
	if err := app.Conversations.UpdateLastMessageAt(message.ConversationID.Int64); err != nil {
		log.Println("❌ Failed to update last_message_at:", err)
		return err
	}

	// Prepare message for broadcast (ensure type is set)
	broadcastMessage := *message
	broadcastMessage.Type = "message" // Ensure type is set for broadcast

	log.Printf("📤 Broadcasting message: %+v\n", broadcastMessage)
	 wsConn.WriteJSON(ack)
	app.Hub.Broadcast <- broadcastMessage

	return nil
}

func (app *WebApp) handleTypingMessage(user *models.User, message *models.Message) {
	// Forward typing indicator to other participants
	typingMessage := models.Message{
		Type:           "typing",
		AuthorID:       user.ID,
		ConversationID: message.ConversationID,
		SentAt:         time.Now(),
	}

	log.Printf("📤 Broadcasting typing indicator: %+v\n", typingMessage)
	app.Hub.Broadcast <- typingMessage
}


