package util

import (
	"encoding/json"
	"log"

	"github.com/gin-gonic/gin"
)

type LogType int

const (
	LogRequest LogType = iota
	LogResponse
)

// Helper function to handle response logging and JSON formatting
func LogAndHandleResponse(ctx *gin.Context, status int, data any) {
	logData(data, LogResponse)
	ctx.JSON(status, data)
}

// Helper function to handle request logging
func LogIncomingRequest(data any) {
	logData(data, LogRequest)
}

func logData(data any, logType LogType) {
	promptMessage := "request"

	if logType == LogResponse {
		promptMessage = "response"
	}

	jsonResponseData, _ := json.Marshal(data)
	log.Printf("%s => %s\n", promptMessage, jsonResponseData)
}
