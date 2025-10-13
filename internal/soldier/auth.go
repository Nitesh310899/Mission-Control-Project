package soldier

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"sync"
	"time"
)

type TokenManager struct {
	mu          sync.RWMutex
	token       string
	expiry      time.Time
	client      *http.Client
	renewURL    string
	soldierID   string
	refreshing  bool
	refreshCond *sync.Cond
}

func NewTokenManager(soldierID, renewURL string) *TokenManager {
	tm := &TokenManager{
		client:    &http.Client{Timeout: time.Second * 5},
		renewURL:  renewURL,
		soldierID: soldierID,
	}
	tm.refreshCond = sync.NewCond(&tm.mu)
	return tm
}

func (tm *TokenManager) Token() string {
	tm.mu.RLock()
	defer tm.mu.RUnlock()
	return tm.token
}

func (tm *TokenManager) Start(initialToken string, expiresIn int) {
	tm.mu.Lock()
	tm.token = initialToken
	tm.expiry = time.Now().Add(time.Duration(expiresIn) * time.Second)
	tm.mu.Unlock()

	go tm.autoRefresh()
}

func (tm *TokenManager) autoRefresh() {
	for {
		tm.mu.Lock()
		waitDuration := time.Until(tm.expiry.Add(-5 * time.Second)) // refresh 5 sec before expiry
		if waitDuration < 0 {
			waitDuration = 0
		}
		tm.mu.Unlock()

		time.Sleep(waitDuration)

		if err := tm.refreshToken(); err != nil {
			log.Printf("[TokenManager] Token refresh failed: %v", err)
			time.Sleep(5 * time.Second) // backoff retry
		} else {
			log.Printf("[TokenManager] Token rotated successfully")
		}
	}
}

func (tm *TokenManager) refreshToken() error {
	tm.mu.Lock()
	if tm.refreshing {
		tm.refreshCond.Wait()
		tm.mu.Unlock()
		return nil
	}
	tm.refreshing = true
	tm.mu.Unlock()

	defer func() {
		tm.mu.Lock()
		tm.refreshing = false
		tm.refreshCond.Broadcast()
		tm.mu.Unlock()
	}()

	req, err := http.NewRequest("POST", tm.renewURL, nil)
	if err != nil {
		return err
	}
	tm.mu.RLock()
	req.Header.Set("Authorization", "Bearer "+tm.token)
	tm.mu.RUnlock()

	resp, err := tm.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return errors.New("renew token: unauthorized or error response")
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	var result struct {
		Token     string `json:"token"`
		ExpiresIn int    `json:"expires_in"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return err
	}

	tm.mu.Lock()
	tm.token = result.Token
	tm.expiry = time.Now().Add(time.Duration(result.ExpiresIn) * time.Second)
	tm.mu.Unlock()

	return nil
}
