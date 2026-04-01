package response

import "blog.alphazer01214.top/internal/entity"

type StandardAiResponse struct {
	Agent   *entity.Agent
	Content string
	Status  bool
}
