package shutdown

import "sync"

var OutstandingOpCounter = sync.WaitGroup{}
