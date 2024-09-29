package secret

import (
	"fmt"
	"math/rand"
	"strings"
	"time"
)

func MakeLetter() []string {
	var letter []string
	for i := 'a'; i <= 'z'; i++ {
		letter = append(letter, fmt.Sprintf("%c", i))
	}
	return letter
}

func MakeLetterUppercase() []string {
	var letter []string
	for i := 'a'; i <= 'z'; i++ {
		item := fmt.Sprintf("%c", i)
		letter = append(letter, strings.ToUpper(item))
	}
	return letter
}

func MakeNumber() []int {
	Number := make([]int, 10)
	for i := range Number {
		Number[i] = 0 + i
	}
	return Number
}

func MakeLetterNumber() []string {
	number := MakeNumber()
	letter := MakeLetter()
	var strNumber []string
	for i := 0; i < len(number); i++ {
		strNumber = append(strNumber, fmt.Sprintf("%d", rune(number[i])))
	}
	return append(letter, strNumber...)
}

func MakeLetterNumberUppercase() []string {
	number := MakeLetterNumber()
	letter := MakeLetterUppercase()
	var strNumber []string
	for i := 0; i < len(number); i++ {
		strNumber = append(strNumber, fmt.Sprintf("%s", number[i]))
	}
	return append(letter, strNumber...)
}

func MakeSecretSimple(length int) string {
	var passwd []string
	number := strings.ReplaceAll(fmt.Sprintf("%s", fmt.Sprintf("%d", MakeNumber()[:])[1:len(fmt.Sprintf("%d", MakeNumber()[:]))-1]), " ", "")
	rand.Seed(time.Now().UnixNano())
	for i := 0; i < length; i++ {
		passwd = append(passwd, string(number[rand.Intn(len(number))]))
	}
	return strings.Join(passwd, "")
}

func MakeSecretPrimary(length int) string {
	var passwd []string
	number := strings.ReplaceAll(fmt.Sprintf("%s", fmt.Sprintf("%d", MakeNumber()[:])[1:len(fmt.Sprintf("%d", MakeNumber()[:]))-1]), " ", "")
	letter := strings.Join(MakeLetter(), "")
	rawStr := []string{number, letter}
	rand.Seed(time.Now().UnixNano())
	for i := 0; i < 2; i++ {
		passwd = append(passwd, string(rawStr[i][rand.Intn(len(rawStr[i]))]))
	}
	for ii := 0; ii < (length - 2); ii++ {
		passwd = append(passwd, string(strings.Join(rawStr, "")[rand.Intn(len(strings.Join(rawStr, "")))]))
	}
	return strings.Join(passwd, "")
}

func MakeSecretAdvanced(length int) string {
	var passwd []string
	number := strings.ReplaceAll(fmt.Sprintf("%s", fmt.Sprintf("%d", MakeNumber()[:])[1:len(fmt.Sprintf("%d", MakeNumber()[:]))-1]), " ", "")
	letter := strings.Join(MakeLetter(), "")
	letterUppercase := strings.Join(MakeLetterUppercase(), "")
	rawStr := []string{number, letter, letterUppercase}
	rand.Seed(time.Now().UnixNano())
	for i := 0; i < 3; i++ {
		passwd = append(passwd, string(rawStr[i][rand.Intn(len(rawStr[i]))]))
	}
	for ii := 0; ii < (length - 3); ii++ {
		passwd = append(passwd, string(strings.Join(rawStr, "")[rand.Intn(len(strings.Join(rawStr, "")))]))
	}
	return strings.Join(passwd, "")
}

func MakeSecretComplex(length int, salt string) string {
	if salt == "" {
		salt = "!@#$%&()"
	}
	number := strings.ReplaceAll(fmt.Sprintf("%s", fmt.Sprintf("%d", MakeNumber()[:])[1:len(fmt.Sprintf("%d", MakeNumber()[:]))-1]), " ", "")
	letter := MakeLetter()
	letterUppercase := MakeLetterUppercase()
	rawStr := []string{
		number,
		strings.Join(letterUppercase, ""),
		strings.Join(letter, ""),
		salt,
	}
	var passwd []string
	rand.Seed(time.Now().UnixNano())
	for i := 0; i < 4; i++ {
		passwd = append(passwd, string(rawStr[i][rand.Intn(len(rawStr[i]))]))
	}
	for ii := 0; ii < (length - 4); ii++ {
		passwd = append(passwd, string(strings.Join(rawStr, "")[rand.Intn(len(strings.Join(rawStr, "")))]))
	}
	return strings.Join(passwd, "")
}
