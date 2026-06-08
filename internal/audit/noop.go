package audit

import (
	"context"

	"github.com/mohammad-farrokhnia/go-ledger/internal/ledger"
)

type NoOp struct{}

func (n *NoOp) Log(_ context.Context, _ ledger.AuditEntry) error {
	return nil
}

var _ ledger.Auditor = (*NoOp)(nil)
