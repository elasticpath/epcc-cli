package clictx

import "context"

var Ctx context.Context = nil

var Cancel context.CancelFunc = nil

func init() {
	Ctx, Cancel = context.WithCancel(context.Background())
}
