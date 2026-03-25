package utils

import "blog.alphazer01214.top/internal/global"

func IsGlobalVarLoaded() bool {
	return global.DB != nil && global.Config != nil && global.Log != nil && global.Redis != nil
}
