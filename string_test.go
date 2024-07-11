package hubur

import (
	"github.com/stretchr/testify/suite"
	"testing"
)

func TestStrUtils(t *testing.T) {
	suite.Run(t, new(StrUtilsTest))
}

type StrUtilsTest struct{ suite.Suite }

func (s *StrUtilsTest) TestChunkString() {
	assert := s.Require()
	assert.Equal(ChunkString("hello world", 3), []string{"hel", "lo ", "wor", "ld"})
	assert.Equal(ChunkString("富强、民主、文明、和谐、自由，平等，公正，法治", 3),
		[]string{"富强、", "民主、", "文明、", "和谐、", "自由，", "平等，", "公正，", "法治"})
}

func (s *StrUtilsTest) TestSha256() {
	assert := s.Require()
	v := "c3ab8ff13720e8ad9047dd39466b3c8974e592c2fa383d4a3960714caef0c4f2"
	assert.Equal(Sha256([]byte("foobar")), v)
	assert.Equal(Sha256String("foobar"), v)
}

func (s *StrUtilsTest) TestEscapeInvalidUTF8Byte() {
	assert := s.Require()
	origin := `\\xdb\xdb` + "\xbd\x20\xe2\x8c\x98你\u597d'\"\n\x00"
	excepted := `\\xdb\xdb\xbd ⌘你好'"` + "\n\\x00"
	output := EscapeInvalidUTF8Byte([]byte(origin))
	assert.Equal([]byte(excepted), []byte(output))
}
