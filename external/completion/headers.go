package completion

import (
	"strings"
	"sync"
)

// Map that specifies headers that we will complete, and how to auto-complete their values.
var supportedHeadersToCompletionRequest = map[string]*Request{
	"EP-Beta-Features": nil,
	"EP-Channel":       nil,
	"EP-Context-Tag":   nil,
	"EP-Account-Management-Authentication-Token": nil,
	"X-Moltin-Customer-Token":                    nil,
	"X-Moltin-Currency": {
		Type: CompleteCurrency,
	},
	"X-Moltin-Currencies": nil,
}

var supportedHeadersOriginalCasing = map[string]string{}

var headersMutex = &sync.RWMutex{}

func postProcessMap() {

	newSupportedHeadersToCompletionRequest := make(map[string]*Request, len(supportedHeadersToCompletionRequest))

	for k, v := range supportedHeadersToCompletionRequest {
		newSupportedHeadersToCompletionRequest[strings.ToLower(k)] = v
		supportedHeadersOriginalCasing[strings.ToLower(k)] = k
	}

	supportedHeadersToCompletionRequest = newSupportedHeadersToCompletionRequest

}

func init() {
	headersMutex.Lock()
	defer headersMutex.Unlock()
	postProcessMap()
}

func AddHeaderCompletions(hc map[string]*Request) {
	headersMutex.Lock()
	defer headersMutex.Unlock()

	for k, v := range hc {
		supportedHeadersToCompletionRequest[k] = v
	}
	postProcessMap()
}
