package shutdown

import "sync/atomic"

var ShutdownFlag = atomic.Bool{}
