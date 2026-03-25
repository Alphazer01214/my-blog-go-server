package utils

import (
	"sync"
)

// TokenBlacklist 用于存储已登出的 token（黑名单）
type TokenBlacklist struct {
	mu     sync.RWMutex
	tokens map[string]bool
}

var blacklist = &TokenBlacklist{
	tokens: make(map[string]bool),
}

// AddToBlacklist 将 token 加入黑名单
func AddToBlacklist(token string) {
	blacklist.mu.Lock()
	defer blacklist.mu.Unlock()
	blacklist.tokens[token] = true
}

// IsBlacklisted 检查 token 是否在黑名单中
func IsBlacklisted(token string) bool {
	blacklist.mu.RLock()
	defer blacklist.mu.RUnlock()
	return blacklist.tokens[token]
}

// RemoveFromBlacklist 从黑名单中移除 token（可选，用于清理）
func RemoveFromBlacklist(token string) {
	blacklist.mu.Lock()
	defer blacklist.mu.Unlock()
	delete(blacklist.tokens, token)
}
