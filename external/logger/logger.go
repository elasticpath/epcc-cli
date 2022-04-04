package logger

import (
	log "github.com/sirupsen/logrus"
	"os"
)

// ③ Map 3rd party enumeration values to their textual representations
var LoglevelIds = map[log.Level][]string{
	log.TraceLevel: {"trace"},
	log.DebugLevel: {"debug"},
	log.InfoLevel:  {"info"},
	log.WarnLevel:  {"warning", "warn"},
	log.ErrorLevel: {"error"},
	log.FatalLevel: {"fatal"},
	log.PanicLevel: {"panic"},
}

var Loglevel log.Level = log.InfoLevel

func init() {
	log.SetOutput(os.Stderr)
}
