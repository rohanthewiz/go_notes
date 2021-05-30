package utils

func TruncString(str string, ln int) string {
	if len(str) > ln {
		return str[:ln]
	}
	return str
}
