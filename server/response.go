package server

import "universum/resp3"

func getFormattedClientMessage(data interface{}, code uint32, message string) string {
	return resp3.EncodedRESP3Response([]interface{}{
		data,
		code,
		"",
	})
}
