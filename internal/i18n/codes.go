package i18n

type MessageCode string

const (
	// System.
	MsgHealthOK MessageCode = "HEALTH_OK"

	// Generic success.
	MsgFetched MessageCode = "FETCHED"
	MsgCreated MessageCode = "CREATED"

	// Generic errors.
	MsgBadRequest       MessageCode = "BAD_REQUEST"
	MsgNotFound         MessageCode = "NOT_FOUND"
	MsgValidationFailed MessageCode = "VALIDATION_FAILED"
	MsgInternalError    MessageCode = "INTERNAL_ERROR"
	MsgConflict         MessageCode = "CONFLICT"

	// Domain — accounts.
	MsgAccountNotFound    MessageCode = "ACCOUNT_NOT_FOUND"
	MsgAccountCreated     MessageCode = "ACCOUNT_CREATED"

	// Domain — transactions.
	MsgTransactionCreated    MessageCode = "TRANSACTION_CREATED"
	MsgTransactionNotFound   MessageCode = "TRANSACTION_NOT_FOUND"
	MsgInsufficientFunds     MessageCode = "INSUFFICIENT_FUNDS"
	MsgCurrencyMismatch      MessageCode = "CURRENCY_MISMATCH"
	MsgDuplicateTransaction  MessageCode = "DUPLICATE_TRANSACTION"
	MsgInvalidAmount         MessageCode = "INVALID_AMOUNT"
	MsgSameAccount           MessageCode = "SAME_ACCOUNT"
	MsgInvalidAccountType    MessageCode = "INVALID_ACCOUNT_TYPE"
	MsgInvalidCurrencyCode   MessageCode = "INVALID_CURRENCY_CODE"
)
