package reddit

import (
	"fmt"
)

var (
	ErrPermissionDenied   = fmt.Errorf("unauthorized access to endpoint")
	ErrBusy               = fmt.Errorf("reddit is busy right now")
	ErrRateLimit          = fmt.Errorf("reddit is rate limiting requests")
	ErrGateway            = fmt.Errorf("502 bad gateway code from Reddit")
	ErrGatewayTimeout     = fmt.Errorf("504 gateway timeout from Reddit")
	ErrThreadDoesNotExist = fmt.Errorf("the requested post does not exist")
)
