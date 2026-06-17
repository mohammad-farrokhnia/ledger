//nolint:staticcheck // U+200C (ZWNJ) is intentional Persian typography
package i18n

var faMessages = map[MessageCode]string{
	MsgHealthOK:             "سرویس در حال اجرا است.",
	MsgFetched:              "با موفقیت دریافت شد.",
	MsgCreated:              "با موفقیت ایجاد شد.",
	MsgBadRequest:           "درخواست نامعتبر.",
	MsgNotFound:             "منبع یافت نشد.",
	MsgValidationFailed:     "اعتبارسنجی ناموفق.",
	MsgInternalError:        "خطای داخلی سرور.",
	MsgConflict:             "منبع از قبل وجود دارد.",
	MsgAccountNotFound:      "حساب یافت نشد.",
	MsgAccountCreated:       "حساب با موفقیت ایجاد شد.",
	MsgTransactionCreated:   "تراکنش با موفقیت انجام شد.",
	MsgTransactionNotFound:  "تراکنش یافت نشد.",
	MsgInsufficientFunds:    "موجودی کافی نیست.",
	MsgCurrencyMismatch:     "واحد پولی حساب‌ها با هم مطابقت ندارد.",
	MsgDuplicateTransaction: "تراکنشی با این کلید idempotency قبلاً وجود دارد.",
	MsgInvalidAmount:        "مقدار باید بزرگ‌تر از صفر باشد.",
	MsgSameAccount:          "حساب مبدأ و مقصد باید متفاوت باشند.",
	MsgInvalidAccountType:   "نوع حساب نامعتبر است.",
	MsgInvalidCurrencyCode:  "کد ارز باید یک کد سه‌حرفی ISO 4217 باشد.",
}
