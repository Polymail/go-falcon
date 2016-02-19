package parser

import (
	"testing"
)

type mimeInvalidNameHeaderTest struct {
	From string
	To   string
}

var mimeInvalidNameHeaderTests = []mimeInvalidNameHeaderTest{
	{"Content-Type: text/plain; name=test.txt", "Content-Type: text/plain; name=test.txt"},
	{"Content-Type: image/png; name=test-with-dash.png", "Content-Type: image/png; name=test-with-dash.png"},
	{"Content-Type: text/plain; name=This is a test.txt", "Content-Type: text/plain; name=\"This is a test.txt\""},
	{"Content-Disposition: attachment;\n   filename=This is a test.txt", "Content-Disposition: attachment;\n   filename=\"This is a test.txt\""},
	//{"Content-Type: application/octet-stream; name*=iso-2022-jp'ja'01%20Quien%20Te%20Dij%8aat.%20Pitbull.mp3", "Content-Type: application/octet-stream; name=\"01 Quien Te Dijufffdat. Pitbull.mp3\""},
	//{"Content-Type: application/octet-stream; name*0=iso-2022-jp'ja'01%20Quien%20Te%20Dij%8aat.%20Pitbull.mp3 name*1=iso-2022-jp'ja'01%20Quien%20Te%20Dij%8aat.%20Pitbull.mp3", "Content-Type: application/octet-stream; name=\"01 Quien Te Dij\\ufffdat. Pitbull.mp3\" name=\"01 Quien Te Dijat. Pitbull.mp3\""},
}

func TestMimeInvalidNameHeader(t *testing.T) {
	for _, header := range mimeInvalidNameHeaderTests {
		decodedValue := FixMailEncodedHeader(header.From)
		expectEq(t, header.To, decodedValue, "Value of decoded with name header")
	}
}

type mimeHeaderTest struct {
	From string
	To   string
}

var mimeHeaderTests = []mimeHeaderTest{
	{"=?iso-8859-1?q?J=F6rg_Doe?=", "Jörg Doe"},
	{"=?utf-8?q?J=C3=B6rg_Doe?=", "Jörg Doe"},
	{"=?ISO-8859-1?Q?Andr=E9?=", "André"},
	{"=?ISO-8859-1?B?SvZyZw==?=", "Jörg"},
	{"=?UTF-8?B?SsO2cmc=?=", "Jörg"},
	{"illness notification =?8bit?Q?ALPH=C3=89E?=", "illness notification ALPHÉE"},
	{"=?UTF-8?B?44G+44G/44KA44KB44KC?=", "まみむめも"},
	{"=?utf-8?q?J=C3=B6rg_Doe?=. =?utf-8?q?J=C3=B6rg_Doe?=", "Jörg Doe. Jörg Doe"},
	{"=?iso-8859-1?Q?=A1Hola,_se=F1or!?=", "¡Hola, señor!"},
	{"=?UTF-8?B?0L/RgNC40LLQtdGCINCy0YHQtdC8?=", "привет всем"},
	{"=?UTF-8?q?=D0=BF=D1=80=D0=B8=D0=B2=D0=B5=D1=82=20=D0=BC=D0=B8=D1=80?=", "привет мир"},
	{`=?UTF-8?Q?=D0=92=D1=8B=D0=B1=D1=80=D0=B0=D0=BD?=
=?UTF-8?Q?_=D0=B8=D1=81=D0=BF=D0=BE=D0=BB=D0=BD=D0=B8=D1=82=D0=B5=D0=BB=D1=8C?=
=?UTF-8?Q?_=D0=B7=D0=B0=D0=BA=D0=B0=D0=B7=D0=B0_=D0=BD=D0=B0?=
=?UTF-8?Q?_=C2=AB=D0=A4=D1=80=D0=B8=D0=BB=D0=B0=D0=BD=D1=81=D0=B8=D0=BC=C2=BB?=`, `Выбран исполнитель заказа на «Фрилансим»`},
	{"=?UTF-8?Q?=D0=9F=D1=80=D0=B8=D0=B2=D0=B5=D1=82=20=D0=BC=D0=B8=D1=80=20=D0=B8=20=D0=BF=D0=BE=D0=B4=D1=87=D0=B5=D1=80=D0=BA=D0=B8=D0=B2=D0=B0=D0=BD=D0=B8=D0=B5?=", "Привет мир и подчеркивание"},
	{"=?utf-8?b?W0RqYW5nb10g0JfQsNC60L7QvdGH0LjQu9C40YHRjCDQutC70Y7Rh9C4INCi?= =?utf-8?b?0LDQudC8INC60LDRgNGC0LAg0L3QsCA2MCDQtNC90LXQuSBSVQ==?=", "[Django] Закончились ключи Тайм карта на 60 дней RU"},
	{"=?EUC-KR?Q?NOTE:_=C7=D1=B1=B9=B8=BB=B7=CE_=C7=CF=B4=C2_=B0=CD?=", "NOTE: 한국말로 하는 것"},
	{"Foo =?UTF-8?B?w6HDq8O0?= =?UTF-8?B?IMOuw7w=?=", "Foo áëô îü"},
	{"=?utf-8?B?0L3QvtCy0YvQuSDRgdC+0YLRgNGD0LTQvdC40Log4oCUINC00L7RgNC+0YQ=?=\n =?utf-8?B?0LXQtdCy?=", "новый сотрудник — дорофеев"},
	{"=?utf-8?B?0L3QvtCy0YvQuSDRgdC+0YLRgNGD0LTQvdC40Log4oCUINC00L7RgNC+0YQ=?==?utf-8?B?0LXQtdCy?=", "новый сотрудник — дорофеев"},
	{"=?UTF-8?Q?This_is_=E3=81=82_string?==?UTF-8?Q?This_is_=E3=81=82_string?= Some non encoded stuff =?UTF-8?B?VGhpcyBpcyDjgYIgc3RyaW5n?= \r\nMore non encoded stuff", "This is あ stringThis is あ string Some non encoded stuff This is あ string \r\nMore non encoded stuff"},
	{"=?ISO-8859-1?Q?Nicolas_Fouch=E9_has_accepted_your_invitation_to_Gmail?=", "Nicolas Fouché has accepted your invitation to Gmail"},
	{"=?Shift_JIS?Q?=93=FA=96{=8C=EA=?=", "日本語="},
	{"=?ISO-8859-2?B?dSB1bmRlcnN0YW5kIHRoZSBleGFtcGxlLg==?=", "u understand the example."},
	{"=?UTF-16?B?MEIwRDBGMEgwSg==?=", "あいうえお"},
	{"=?UTF-32?B?AAAwQgAAMEQAADBGAAAwSAAAMEo=?=", "あいうえお"},
	{"=?iso-2022-jp?B?GyRCJTAlayE8JV0lcyEmJTglYyVRJXMzdDwwMnEbKEI=?=", "グルーポン・ジャパン株式会"},
	{"=?8bit?Q?ALPH=C3=89E?=", "ALPHÉE"},
	{"=?UTF-8?B?TGluZHPjgYLjgYJy?=", "Lindsああr"},
}

func TestMimeHeaderDecode(t *testing.T) {
	for _, header := range mimeHeaderTests {
		decodedHeader := MimeHeaderDecode(header.From)
		expectEq(t, header.To, decodedHeader, "Value of decoded header")
	}
}
