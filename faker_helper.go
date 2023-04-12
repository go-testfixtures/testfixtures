package testfixtures // import "github.com/go-testfixtures/testfixtures/v3"

import (
	"strconv"
	"time"
	"text/template"
	"math"
	"strings"
	"github.com/jaswdr/faker"
)

var (
	fake = faker.New()
	fakerFuncMap = template.FuncMap{
            "fakePersonName": fakePersonName,
            "fakeLatitude": fakeLatitude,
            "fakeLongitude": fakeLongitude,
            "fakeTimestamp": fakeTimestamp,
            "fakeDate": fakeDate,
            "fakeAddress": fakeAddress,
            "fakeInt": fakeInt,
            "fakeIntBetween": fakeIntBetween,
            "fakeFloat": fakeFloat,
            "fakeParagraph": fakeParagraph,
            "fakeSentence": fakeSentence,
            "fakeWord": fakeWord,
            "fakePhone": fakePhone,
            "fakeDay": fakeDay,
            "fakeMonthName": fakeMonthName,
            "fakeYear": fakeYear,
            "fakeDayOfWeek": fakeDayOfWeek,
            "fakeEmail": fakeEmail,
            "fakeURL": fakeURL,
            "fakeIP": fakeIP,
            "fakePassword": fakePassword,
            "fakeCompanyName": fakeCompanyName,
            "fakeJobTitle": fakeJobTitle,
            "fakeOneOfStrings": fakeOneOfStrings,
        }
)

const (
	TIMESTAMP_FORMAT = "2006-01-02 15:04:05"
	DATE_FORMAT = "2006-01-02"
)


func fakePersonName() string {
	return escapeSpecTemplateChars(fake.Person().Name())
}

func fakeLatitude() string {
	return fakeFloat(6, -89, 89)
}

func fakeLongitude() string {
	return fakeFloat(6, -179, 179)
}

func fakeTimestamp() string {
	return time.Unix(fake.Time().Unix(time.Now()), 0).Format(TIMESTAMP_FORMAT)
}

func fakeDate() string {
	return time.Unix(fake.Time().Unix(time.Now()), 0).Format(DATE_FORMAT)
}

func fakeAddress() string {
	return escapeSpecTemplateChars(fake.Address().Address())
}

func fakeInt(size int) string {
	return strconv.Itoa(fake.RandomNumber(size))
}

func fakeIntBetween(min, max int) string {
	return strconv.Itoa(fake.IntBetween(min, max))
}

func fakeFloat(decimals, min, max int) string {
	var val = float64(fake.IntBetween(min, max)) + (float64(fake.RandomNumber(decimals)) / math.Pow10(decimals))
	return strconv.FormatFloat(val,'f', -1, 64)
}

func fakeParagraph(nSentences int) string {
	return fake.Lorem().Paragraph(nSentences)
}

func fakeSentence(nWords int) string {
	return fake.Lorem().Sentence(nWords)
}

func fakeWord() string {
	return fake.Lorem().Word()
}

func fakePhone() string {
	return fake.Phone().E164Number()
}

func fakeDay() string {
	return strconv.Itoa(fake.Time().DayOfMonth())
}

func fakeMonthName() string {
	return fake.Time().MonthName()
}

func fakeYear() string {
	return strconv.Itoa(fake.Time().Year())
}

func fakeDayOfWeek() string {
	return fake.Time().DayOfWeek().String()
}

func fakeEmail() string {
	return fake.Internet().Email()
}

func fakeURL() string {
	return fake.Internet().URL()
}

func fakeIP() string {
	return fake.Internet().Ipv4()
}

func fakePassword() string {
	return escapeSpecTemplateChars(fake.Internet().Password())
}

func fakeCompanyName() string {
	return escapeSpecTemplateChars(fake.Company().Name())
}

func fakeJobTitle() string {
	return escapeSpecTemplateChars(fake.Company().JobTitle())
}

func fakeOneOfStrings(strings ...string) string {
	return fake.RandomStringElement(strings)
}

func escapeSpecTemplateChars(val string) string {
	curated := strings.Replace(val, `"`, `\"`, -1)
	return strings.Replace(curated, `'`, `\'`, -1)
}

