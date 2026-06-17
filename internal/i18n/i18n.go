package i18n

import "strings"

type Lang string

const (
	LangEN Lang = "en"
	LangFA Lang = "fa"
)

var translations = map[Lang]map[MessageCode]string{
	LangEN: enMessages,
	LangFA: faMessages,
}

func Translate(acceptLang string, code MessageCode) string {
	lang := DetectLang(acceptLang)

	if msgs, ok := translations[lang]; ok {
		if msg, ok := msgs[code]; ok {
			return msg
		}
	}

	if msg, ok := translations[LangEN][code]; ok {
		return msg
	}

	return string(code)
}

func DetectLang(header string) Lang {
	if header == "" {
		return LangEN
	}

	parts := strings.Split(header, ",")
	for _, part := range parts {
		tag := strings.Split(strings.TrimSpace(part), ";")[0]
		primary := Lang(strings.ToLower(strings.Split(tag, "-")[0]))
		if _, ok := translations[primary]; ok {
			return primary
		}
	}

	return LangEN
}
