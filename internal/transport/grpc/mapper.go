package grpc

import (
	"google.golang.org/protobuf/types/known/timestamppb"

	ledgerv1 "github.com/mohammad-farrokhnia/go-ledger/api/proto/ledger/v1"
	"github.com/mohammad-farrokhnia/go-ledger/internal/ledger"
)

func accountToProto(a ledger.Account) *ledgerv1.Wallet {
	return &ledgerv1.Wallet{
		Id:           a.ID,
		Name:         a.Name,
		Type:         accountTypeToProto(a.Type),
		CurrencyCode: a.CurrencyCode,
		Balance:      a.Balance,
		CreatedAt:    timestamppb.New(a.CreatedAt),
		UpdatedAt:    timestamppb.New(a.UpdatedAt),
	}
}

func transactionToProto(t ledger.Transaction) *ledgerv1.Transaction {
	return &ledgerv1.Transaction{
		Id:             t.ID,
		IdempotencyKey: t.IdempotencyKey,
		FromAccountId:  t.FromAccountID,
		ToAccountId:    t.ToAccountID,
		Amount:         t.Amount,
		CurrencyCode:   t.CurrencyCode,
		Status:         transactionStatusToProto(t.Status),
		CreatedAt:      timestamppb.New(t.CreatedAt),
		UpdatedAt:      timestamppb.New(t.UpdatedAt),
	}
}

func entryToProto(e ledger.Entry) *ledgerv1.Entry {
	return &ledgerv1.Entry{
		Id:            e.ID,
		AccountId:     e.AccountID,
		TransactionId: e.TransactionID,
		Amount:        e.Amount,
		CreatedAt:     timestamppb.New(e.CreatedAt),
	}
}

func accountTypeToProto(t ledger.AccountType) ledgerv1.AccountType {
	switch t {
	case ledger.AccountTypeUser:
		return ledgerv1.AccountType_ACCOUNT_TYPE_USER
	case ledger.AccountTypeSystem:
		return ledgerv1.AccountType_ACCOUNT_TYPE_SYSTEM
	case ledger.AccountTypeProvider:
		return ledgerv1.AccountType_ACCOUNT_TYPE_PROVIDER
	default:
		return ledgerv1.AccountType_ACCOUNT_TYPE_UNSPECIFIED
	}
}

func transactionStatusToProto(s ledger.TransactionStatus) ledgerv1.TransactionStatus {
	switch s {
	case ledger.TransactionStatusPending:
		return ledgerv1.TransactionStatus_TRANSACTION_STATUS_PENDING
	case ledger.TransactionStatusCompleted:
		return ledgerv1.TransactionStatus_TRANSACTION_STATUS_COMPLETED
	case ledger.TransactionStatusFailed:
		return ledgerv1.TransactionStatus_TRANSACTION_STATUS_FAILED
	default:
		return ledgerv1.TransactionStatus_TRANSACTION_STATUS_UNSPECIFIED
	}
}

func protoAccountTypeToDomain(t ledgerv1.AccountType) ledger.AccountType {
	switch t {
	case ledgerv1.AccountType_ACCOUNT_TYPE_USER:
		return ledger.AccountTypeUser
	case ledgerv1.AccountType_ACCOUNT_TYPE_SYSTEM:
		return ledger.AccountTypeSystem
	case ledgerv1.AccountType_ACCOUNT_TYPE_PROVIDER:
		return ledger.AccountTypeProvider
	default:
		return ""
	}
}
