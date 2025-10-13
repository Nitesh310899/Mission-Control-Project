package commander

import (
	"encoding/json"
	"log"
	"mission-control/pkg/models"
	"net/http"
	"strings"

	"github.com/google/uuid"
)

type Handler struct {
	store *MissionStore
	queue *QueueManager
}

func NewHandler(store *MissionStore, queue *QueueManager) *Handler {
	return &Handler{store: store, queue: queue}
}

func (h *Handler) PostMission(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Payload string `json:"payload"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Payload == "" {
		http.Error(w, "Invalid payload", http.StatusBadRequest)
		return
	}
	id := uuid.New().String()
	m := h.store.CreateMission(id, body.Payload)
	if err := h.queue.PublishOrder(m); err != nil {
		log.Printf("Publish error: %v", err)
		http.Error(w, "Queue error", http.StatusInternalServerError)
		return
	}
	resp := struct {
		MissionID string `json:"mission_id"`
	}{
		MissionID: id,
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(resp)
}

func (h *Handler) GetMissionStatus(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Path[len("/missions/"):]
	m, ok := h.store.GetMission(id)
	if !ok {
		http.Error(w, "Mission not found", http.StatusNotFound)
		return
	}
	resp := struct {
		MissionID string               `json:"mission_id"`
		Status    models.MissionStatus `json:"status"`
	}{
		MissionID: m.ID,
		Status:    m.Status,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *Handler) RenewToken(w http.ResponseWriter, r *http.Request) {
	// Soldier must provide current token as Bearer Authorization to prove identity
	auth := r.Header.Get("Authorization")
	if auth == "" {
		log.Printf("[RenewToken] Missing Authorization header")
		http.Error(w, "Missing Authorization header", http.StatusUnauthorized)
		return
	}
	tokenStr := strings.TrimPrefix(auth, "Bearer ")
	log.Printf("[RenewToken] Received token for renewal: %s", tokenStr)

	claims, err := VerifyToken(tokenStr)
	if err != nil {
		log.Printf("[RenewToken] Token verification failed: %v", err)
		http.Error(w, "Invalid token", http.StatusUnauthorized)
		return
	}

	newToken, err := GenerateToken(claims.SoldierID)
	if err != nil {
		log.Printf("[RenewToken] Failed to generate new token: %v", err)
		http.Error(w, "Failed to generate token", http.StatusInternalServerError)
		return
	}

	log.Printf("Token rotation for soldier %s", claims.SoldierID) // Log rotation

	resp := struct {
		Token     string `json:"token"`
		ExpiresIn int    `json:"expires_in"` // seconds
	}{
		Token:     newToken,
		ExpiresIn: 30,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// func (h *Handler) RenewToken(w http.ResponseWriter, r *http.Request) {
// 	// Soldier must provide current token as Bearer Authorization to prove identity
// 	auth := r.Header.Get("Authorization")
// 	if auth == "" {
// 		http.Error(w, "Missing Authorization header", http.StatusUnauthorized)
// 		return
// 	}
// 	tokenStr := strings.TrimPrefix(auth, "Bearer ")
// 	claims, err := VerifyToken(tokenStr)
// 	if err != nil {
// 		http.Error(w, "Invalid token", http.StatusUnauthorized)
// 		return
// 	}

// 	newToken, err := GenerateToken(claims.SoldierID)
// 	if err != nil {
// 		http.Error(w, "Failed to generate token", http.StatusInternalServerError)
// 		return
// 	}

// 	log.Printf("Token rotation for soldier %s", claims.SoldierID) // Log rotation

// 	resp := struct {
// 		Token     string `json:"token"`
// 		ExpiresIn int    `json:"expires_in"` // seconds
// 	}{
// 		Token:     newToken,
// 		ExpiresIn: 30,
// 	}
// 	w.Header().Set("Content-Type", "application/json")
// 	json.NewEncoder(w).Encode(resp)
// }
