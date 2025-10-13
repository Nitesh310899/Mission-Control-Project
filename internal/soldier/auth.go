package soldier

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
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

type TokenResponse struct {
	Token     string `json:"token"`
	ExpiresIn int    `json:"expires_in"`
}

func GetInitialToken(commanderURL, soldierID string) (string, int, error) {
	reqBody, err := json.Marshal(map[string]string{
		"soldier_id": soldierID,
	})
	if err != nil {
		return "", 0, err
	}

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Post(fmt.Sprintf("%s/tokens/issue", commanderURL), "application/json", bytes.NewBuffer(reqBody))
	if err != nil {
		return "", 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", 0, fmt.Errorf("token issuance failed with status %d", resp.StatusCode)
	}

	var tokenResp TokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return "", 0, err
	}

	return tokenResp.Token, tokenResp.ExpiresIn, nil
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
	backoff := 5 * time.Second
	maxBackoff := 1 * time.Minute

	for {
		tm.mu.Lock()
		waitDuration := time.Until(tm.expiry.Add(-5 * time.Second))
		if waitDuration < 0 {
			waitDuration = 0
		}
		tm.mu.Unlock()

		time.Sleep(waitDuration)

		err := tm.refreshToken()
		if err != nil {
			log.Printf("[TokenManager] Token refresh failed: %v", err)
			time.Sleep(backoff)
			backoff *= 2
			if backoff > maxBackoff {
				backoff = maxBackoff
			}
		} else {
			log.Printf("[TokenManager] Token rotated successfully")
			backoff = 5 * time.Second
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

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		log.Printf("[TokenManager] Token renew failed. HTTP %d: %s", resp.StatusCode, string(body))
		return errors.New("renew token: unauthorized or error response")
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
