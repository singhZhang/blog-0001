package api

import (
	"encoding/json"
	"net/http"
)

// writeJSON 以 JSON 写出响应。编码失败通常意味着客户端已断开，无法补救，忽略即可。
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

// errorBody 是错误响应的统一结构：{"error":{"code":"...","message":"..."}}。
type errorBody struct {
	Error struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

// writeError 写出统一格式的错误响应。
func writeError(w http.ResponseWriter, status int, code, message string) {
	var body errorBody
	body.Error.Code = code
	body.Error.Message = message
	writeJSON(w, status, body)
}
