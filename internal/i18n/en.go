package i18n

var enMessages = map[MessageCode]string{
	MsgHealthOK:             "Service is up and running.",
	MsgFetched:              "Fetched successfully.",
	MsgCreated:              "Created successfully.",
	MsgBadRequest:           "Bad request.",
	MsgNotFound:             "Resource not found.",
	MsgValidationFailed:     "Validation failed.",
	MsgInternalError:        "Internal server error.",
	MsgConflict:             "Resource already exists.",
	MsgAccountNotFound:      "Account not found.",
	MsgAccountCreated:       "Account created successfully.",
	MsgTransactionCreated:   "Transaction completed successfully.",
	MsgTransactionNotFound:  "Transaction not found.",
	MsgInsufficientFunds:    "Insufficient funds.",
	MsgCurrencyMismatch:     "Currency mismatch between accounts.",
	MsgDuplicateTransaction: "Transaction with this idempotency key already exists.",
	MsgInvalidAmount:        "Amount must be greater than zero.",
	MsgSameAccount:          "Source and destination accounts must be different.",
	MsgInvalidAccountType:   "Invalid account type.",
	MsgInvalidCurrencyCode:  "Currency code must be a 3-letter ISO 4217 code.",
}
