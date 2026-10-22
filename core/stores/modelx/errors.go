package modelx

import "errors"

// ErrTransactionRequired indicates a locking read was called without a session.
// The caller must pass the session from SqlConn.TransactCtx, not a standalone connection.
var ErrTransactionRequired = errors.New("modelx: locking read requires a transaction session")
