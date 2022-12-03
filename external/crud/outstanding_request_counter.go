package crud

import "sync"

var OutstandingRequestCounter = sync.WaitGroup{}
