package entity

import "blog.alphazer01214.top/internal/config"

type Agent struct {
	Name      string
	LLMConfig *config.LLM
}
