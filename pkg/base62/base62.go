package base62

import "strings"

const alphabet = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
const base = 62

//DB ID into base62 converstion

func Encode(id int64) string {
	if id == 0 {
		return "0"
	}
	var sb strings.Builder
	for id > 0 {
		remainder := id % base
		sb.WriteByte(alphabet[remainder])
		id = id / base
	}
	//Reverse the string
	return reverse(sb.String())
}
func reverse(s string) string {
	runes := []rune(s)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)

}
