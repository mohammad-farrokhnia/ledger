package i18n

import "github.com/mohammad-farrokhnia/go-ledger/internal/ledger"

type Lang string

const (
	LangEN Lang = "en"
	LangFA Lang = "fa"
)

var catalog = map[string]map[Lang]string{
	"account_not_found": {
		LangEN: "account not found",
		LangFA: "حساب یافت نشد",
	},
	"transaction_not_found": {
		LangEN: "transaction not found",
		LangFA: "تراکنش یافت نشد",
	},
	"insufficient_funds": {
		LangEN: "insufficient funds",
		LangFA: "موجودی کافی نیست",
	},
	"currency_mismatch": {
		LangEN: "currency mismatch between accounts",
		LangFA: "واحد پولی حساب‌ها با هم مطابقت ندارد",
	},
	"duplicate_transaction": {
		LangEN: "transaction with this idempotency key already exists",
		LangFA: "تراکنشی با این کلید idempotency قبلاً وجود دارد",
	},
	"invalid_amount": {
		LangEN: "amount must be greater than zero",
		LangFA: "مقدار باید بزرگتر از صفر باشد",
	},
	"same_account": {
		LangEN: "source and destination accounts must be different",
		LangFA: "حساب مبدا و مقصد باید متفاوت باشند",
	},
	"invalid_account_type": {
		LangEN: "invalid account type",
		LangFA: "نوع حساب نامعتبر است",
	},
	"invalid_currency_code": {
		LangEN: "currency code must be a 3-letter ISO 4217 code",
		LangFA: "کد ارز باید یک کد سه‌حرفی ISO 4217 باشد",
	},
	"internal_error": {
		LangEN: "an internal error occurred",
		LangFA: "خطای داخلی رخ داده است",
	},
}

var errorKey = map[error]string{
	ledger.ErrAccountNotFound:      "account_not_found",
	ledger.ErrTransactionNotFound:  "transaction_not_found",
	ledger.ErrInsufficientFunds:    "insufficient_funds",
	ledger.ErrCurrencyMismatch:     "currency_mismatch",
	ledger.ErrDuplicateTransaction: "duplicate_transaction",
	ledger.ErrInvalidAmount:        "invalid_amount",
	ledger.ErrSameAccount:          "same_account",
	ledger.ErrInvalidAccountType:   "invalid_account_type",
	ledger.ErrInvalidCurrencyCode:  "invalid_currency_code",
}

func Message(err error, lang Lang) string {
	if lang == "" {
		lang = LangEN
	}

	key, ok := errorKey[err]
	if !ok {
		key = "internal_error"
	}

	msgs, ok := catalog[key]
	if !ok {
		return err.Error()
	}

	if msg, ok := msgs[lang]; ok {
		return msg
	}

	return msgs[LangEN]
}

func Parse(acceptLanguage string) Lang {
	if len(acceptLanguage) >= 2 {
		switch acceptLanguage[:2] {
		case "fa", "ir":
			return LangFA
		case "en":
			return LangEN
		}
	}
	return LangEN
}
