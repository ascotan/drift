package drift

import (
  "testing"
  "io"
)

func TestCreateScanner(t *testing.T) {
  data := `this is a 风雷动`
  s := NewScanner([]byte(data))
  if s.offset != 0 {
    t.Errorf("offset: expected %v got %v", 0, s.offset)
  }
  if s.lineno != 1 {
    t.Errorf("lineno: expected %v got %v", 1, s.lineno)
  }
  rs := s.reader.Size()
  bs := int64(len([]byte(data)))
  if rs != bs {
    t.Errorf("reader.size: expected %v got %v", bs, rs)
  }
}

func TestNext(t *testing.T) {
  data := `风`
  s := NewScanner([]byte(data))
  runes, err := s.next()
  if err != nil {
    t.Errorf("unexpected error %v", err)
  }
  if len(runes) != 1 {
    t.Errorf("expected %v got %v", 1, len(runes))
  }
  if runes[0] != '风' {
    t.Errorf("expected %v got %v", '风', runes[0])
  }
  if s.offset != 1 {  // offsets are in runes not bytes - this is a 3byte rune
    t.Errorf("expected %v got %v", 1, s.offset)
  }
}

func TestNextLineno(t *testing.T) {
  data := `

  `
  s := NewScanner([]byte(data))
  if s.lineno != 1 {
    t.Errorf("expected lineno to be %v got %v", 1, s.offset)
  }
  _, err := s.next()
  if err != nil {
    t.Errorf("next: unexpected error %v", err)
  }
  if s.lineno != 2 {
    t.Errorf("expected lineno to be %v got %v", 2, s.offset)
  }
}

func TestNextEOF(t *testing.T) {
  data := `风`
  s := NewScanner([]byte(data))
  s.next()

  // read EOF
  runes, err := s.next()
  if err != io.EOF {
    t.Errorf("unexpected error %v", err)
  }
  if len(runes) != 1 {
    t.Errorf("expected %v got %v", 1, len(runes))
  }
  if runes[0] != EOF {
    t.Errorf("expected %v got %v", EOF, runes[0])
  }

  // call again after reading EOF - we should keep getting EOF
  runes, err = s.next()
  if err != io.EOF {
    t.Errorf("next: unexpected error %v", err)
  }
  if len(runes) != 1 {
    t.Errorf("expected %v got %v", 1, len(runes))
  }
  if runes[0] != EOF {
    t.Errorf("next: expected %v got %v", EOF, runes[0])
  }
}

func TestSeek(t *testing.T) {
  data := `hello world`
  s := NewScanner([]byte(data))
  // move 1 byte forward
  err := s.seek(1)
  if err != nil {
    t.Errorf("seek: unexpected error %v", err)
  }
  if s.offset != 1 {
    t.Errorf("expected offset to be %v got %v", 1, s.offset)
  }
  if s.lineno != 1 {
    t.Errorf("expected lineno to be %v got %v", 1, s.lineno)
  }

  // read the rune
  runes, err := s.peek(1)
  if err != nil {
    t.Errorf("unexpected error %v", err)
  }
  if len(runes) != 1 {
    t.Errorf("expected %v got %v", 1, len(runes))
  }
  if runes[0] != 'e' {
    t.Errorf("expected %v got %v", 'e', runes[0])
  }
}

func TestSeekMultibyte(t *testing.T) {
  data := `役立てることができることからきている`
  s := NewScanner([]byte(data))
  // move 1 byte forward
  err := s.seek(1)
  if err != nil {
    t.Errorf("seek: unexpected error %v", err)
  }
  if s.offset != 1 {
    t.Errorf("expected offset to be %v got %v", 1, s.offset)
  }
  if s.lineno != 1 {
    t.Errorf("expected lineno to be %v got %v", 1, s.lineno)
  }

  // read the rune
  runes, err := s.peek(1)
  if err != nil {
    t.Errorf("unexpected error %v", err)
  }
  if len(runes) != 1 {
    t.Errorf("expected %v got %v", 1, len(runes))
  }
  if runes[0] != '立' {
    t.Errorf("expected %v got %v", '立', string(runes[0]))
  }
}

func TestSeekLineno(t *testing.T) {
  data := `hello
  world`
  s := NewScanner([]byte(data))
  err := s.seek(8)
  if err != nil {
    t.Errorf("seek: unexpected error %v", err)
  }
  if s.offset != 8 {
    t.Errorf("expected offset to be %v got %v", 8, s.offset)
  }
  if s.lineno != 2 {
    t.Errorf("expected lineno to be %v got %v", 2, s.lineno)
  }

  // read the rune
  runes, err := s.peek(1)
  if err != nil {
    t.Errorf("unexpected error %v", err)
  }
  if len(runes) != 1 {
    t.Errorf("expected %v got %v", 1, len(runes))
  }
  if runes[0] != 'w' {
    t.Errorf("expected %v got %v", 'w', runes[0])
  }
}

func TestSeekBad(t *testing.T) {
  data := `hello
  world`
  s := NewScanner([]byte(data))
  // seeking past the end of file will just seek to EOF
  err := s.seek(1000)
  if err != nil{
    t.Errorf("unexpected error %v", err)
  }
  // because seek failed, we should not have moved the scanner index
  if s.offset != 13 {
    t.Errorf("expected offset to be %v got %v", 13, s.offset)
  }
  if s.lineno != 2 {
    t.Errorf("expected lineno to be %v got %v", 2, s.lineno)
  }

  // invalid
  err = s.seek(-1000)
  if err != io.EOF {
    t.Errorf("expected io.EOF to be returned")
  }
  // scanner keeps its state after an illegal seek
  if s.offset != 13 {
    t.Errorf("expected offset to be %v got %v", 13, s.offset)
  }
  if s.lineno != 2 {
    t.Errorf("expected lineno to be %v got %v", 2, s.lineno)
  }
}

// func TestPeek(t *testing.T) {
//   data := `风`
//   s := NewScanner([]byte(data))
//   ch, err := s.peek()
//   if err != nil {
//     t.Errorf("peek: unexpected error %v", err)
//   }
//   if ch != '风' {
//     t.Errorf("peek: expected %v got %v", '风', ch)
//   }
//   if s.offset != 0 {  // peeking should not fast forward the offset
//     t.Errorf("peek: expected %v got %v", 0, s.offset)
//   }
//
//   // peeking again should give me the same character
//   ch, err = s.peek()
//   if err != nil {
//     t.Errorf("peek: unexpected error %v", err)
//   }
//   if ch != '风' {
//     t.Errorf("peek: expected %v got %v", '风', ch)
//   }
// }
//
// func TestPeekEOF(t *testing.T) {
//   data := `a`
//   s := NewScanner([]byte(data))
//   err := s.seek(1)
//   if err != nil {
//     t.Errorf("peek: unexpected error %v", err)
//   }
//   ch, err := s.peek()
//   if err != io.EOF && ch != EOF {
//     t.Error("expected io.EOF to be thrown and ch set to EOF")
//   }
//   if s.offset != 1 {  // peeking should not fast forward the offset
//     t.Errorf("peek: expected %v got %v", 1, s.offset)
//   }
// }

func TestPeek(t *testing.T) {
  data := `风雷动`
  s := NewScanner([]byte(data))

  for index, value := range(map[int]string{
    1:"风",
    2:"风雷",
    3:"风雷动",
  }) {
    runes, err := s.peek(index)
    if err != nil {
      t.Errorf("peek: unexpected error %v", err)
    }
    if string(runes) != value {
      t.Errorf("peek: expected %v got %v", value, runes)
    }
  }

  // peek past the EOF but less then bytes in the reader
  // there are 9 bytes in this reader because of the ckj unicode charcters
  offset := s.offset
  runes, err := s.peek(7)
  if err != io.EOF {
    t.Error("peek: expected io.EOF to be throw")
  }
  // peeking past the EOF returns all the runes upto the EOF
  if len(runes) != 3 {
    t.Errorf("peek: expected rune length to be %v got %v", 3, len(runes))
  }
  if string(runes) != data {
    t.Errorf("peek: expected %v got %v", data, runes)
  }
  if s.offset != offset {
    t.Errorf("peek: expected offset to be %v got %v", offset, s.offset)
  }

  // way way past the EOF
  runes, err = s.peek(1000)
  if err != io.EOF {
    t.Error("peek: expected io.EOF to be throw")
  }
  if len(runes) != 3 {
    t.Errorf("peek: expected rune length to be %v got %v", 3, len(runes))
  }
  if string(runes) != data {
    t.Errorf("peek: expected %v got %v", data, runes)
  }
  if s.offset != offset {
    t.Errorf("peek: expected offset to be %v got %v", offset, s.offset)
  }

  // seek n' peek at EOF
  s.seek(3) // EOF
  runes, err = s.peek(1)
  if err != io.EOF {
    t.Error("peek: expected io.EOF to be throw")
  }
  // there should be no runes left to peek
  if len(runes) != 0 {
    t.Errorf("peek: expected rune length to be %v got %v", 0, len(runes))
  }
  if s.offset != 3 {
    t.Errorf("peek: expected offset to be %v got %v", 3, s.offset)
  }
}

func TestHasMoreTokens(t *testing.T) {
  data := `风`
  s := NewScanner([]byte(data))
  more := s.HasMoreTokens()
  if more != true {
    t.Errorf("HasMoreTokens: expected %v got %v", true, more)
  }

  // advance to the end - should have no more tokens
  s.next()
  more = s.HasMoreTokens()
  if more != false {
    t.Errorf("HasMoreTokens: expected %v got %v", false, more)
  }
}

func TestIsWhitespace(t *testing.T) {
  data := `a`
  s := NewScanner([]byte(data))
  for value, expected := range(map[*[]rune]bool {
    &[]rune{' '}  : true,
    &[]rune{'\n'} : true,
    &[]rune{'\r'} : true,
    &[]rune{'\t'} : true,
    &[]rune{'a'} : false,
  }) {
    if s.isWhitespace(*value) != expected {
      t.Errorf("isWhitespace: expected %v got %v", expected, s.isWhitespace(*value))
    }
  }
}

func TestIsIdent(t *testing.T) {
  data := `a`
  s := NewScanner([]byte(data))
  for value, expected := range(map[*[]rune]bool {
    &[]rune{'a'}  : true,
    &[]rune{'A'} : true,
    &[]rune{'0'} : true,
    &[]rune{'.'} : true,
    &[]rune{'呼'} : true,
    &[]rune{'б'} : true,
    &[]rune{' '} : false,
  }) {
    if s.isIdent(*value) != expected {
      t.Errorf("isIdent: expected %v got %v", expected, s.isIdent(*value))
    }
  }
}

func TestIsComment(t *testing.T) {
  data := `a`
  s := NewScanner([]byte(data))
  for value, expected := range(map[*[]rune]bool {
    &[]rune{'/', '/'}       : true,
    &[]rune{'/', '*'}       : true,
    &[]rune{'/', '-'}       : false,
    &[]rune{'/'}            : false,
    &[]rune{'/', '/', '/'}  : true,
    &[]rune{'/', '*', ' '}  : true,
    &[]rune{' ', '/', '/'}  : false,
    &[]rune{'风'}           : false,
    &[]rune{'-', '-', ' '}  : true,
    &[]rune{'-', '-', '风'}  : true,
    &[]rune{'-', '-'}       : true,
    &[]rune{'-'}            : false,
    &[]rune{'-', '-', '+'}  : true,
    &[]rune{'-', '-', '-'}  : true,
  }) {
    if s.isComment(*value) != expected {
      t.Errorf("isComment: expected %v got %v for value '%v'", expected, s.isComment(*value), string(*value))
    }
  }
}

func TestScanIdent(t *testing.T) {
  data := `a bcd 1.2E+10            !@#$%^&*()_+{}:"<>?,./;'[]\'"       z

    役立てることができることからきている`
  s := NewScanner([]byte(data))

  for index, value := range(map[int]string{
    0:"a",
    2:"bcd",
    6:"1.2E+10",
    25:"!@#$%^&*()_+{}:\"<>?,./;'[]\\'\"",
    61:"z",
    68:"役立てることができることからきている",
  }) {
    s.seek(int64(index))
    ident, err := s.scanForIdent()
    if err != nil {
      t.Errorf("unexpected error %v", err)
    }
    if string(ident.runes) != value {
      t.Errorf("expected %v got %v", value, string(ident.runes))
    }
  }
}

func TestScanIdentBadStart(t *testing.T) {
  data := `a bcd 1.2E+10            !@#$%^&*()_+{}:"<>?,./;'[]\'"       z

    役立てることができることからきている`
  s := NewScanner([]byte(data))

  s.seek(63) // seeking into whitespace - should return an empty token
  token, err := s.scanForIdent()
  if err != nil {
    t.Errorf("unexpected error %v", err)
  }
  if len(token.runes) != 0 {
    t.Errorf("expected %v got %v", 0, len(token.runes))
  }
}

func TestScanIdentEOF(t *testing.T) {
  data := `a`
  s := NewScanner([]byte(data))

  s.seek(1) // seeking to EOF
  token, err := s.scanForIdent()
  if err != nil {
    t.Errorf("unexpected error %v", err)
  }
  if len(token.runes) != 0 {
    t.Errorf("expected %v got %v", 0, len(token.runes))
  }
}

func TestScanWhitespace(t *testing.T) {
  data := `a bcd`+"\t"+`1.2E+10            !@#$%^&*()_+{}:"<>?,./;'[]\'"

    役立てることができることからきている`
  s := NewScanner([]byte(data))

  for index, value := range(map[int]string{
    1:" ",
    5:"\t",
    13:"            ",
    54:"\n\n    ",
  }) {
    s.seek(int64(index))
    ws, err := s.scanForWhitespace()
    if err != nil {
      t.Errorf("unexpected error %v", err)
    }
    if string(ws.runes) != value {
      t.Errorf("expected %v got %v", value, string(ws.runes))
    }
  }
}

func TestScanWhitespaceEOF(t *testing.T) {
  data := `a`
  s := NewScanner([]byte(data))

  s.seek(1) // seeking to EOF
  token, err := s.scanForWhitespace()
  if err != nil {
    t.Errorf("unexpected error %v", err)
  }
  if len(token.runes) != 0 {
    t.Errorf("expected %v got %v", 0, len(token.runes))
  }
}

func TestScanComment(t *testing.T) {
  data := ` // here's a comment SELECT * FROM table

  hello world  // you think
  and here's some idests  -- comment at the end of line
  -- sql comment
  /* and well

  throw in
  a
  multiline comment */
  //`
  s := NewScanner([]byte(data))

  for index, value := range(map[int]string{
    1:"// here's a comment SELECT * FROM table",
    57:"// you think",
    96:"-- comment at the end of line",
    128:"-- sql comment",
    145:"/* and well\n\n  throw in\n  a\n  multiline comment */",
    198:"//",
  }) {
    s.seek(int64(index))
    comment, err := s.scanForComment()
    if err != nil {
      t.Errorf("unexpected error %v", err)
    }
    if string(comment.runes) != value {
      t.Errorf("expected '%v' got '%v'", value, string(comment.runes))
    }
  }
}
func TestScanJavaCommentEOF(t *testing.T) {
  data := `/*comment*/`
  s := NewScanner([]byte(data))
  comment, err := s.scanForComment()
  value := "/*comment*/"
  if err != nil {
    t.Errorf("unexpected error %v", err)
  }
  if string(comment.runes) != value {
    t.Errorf("expected %v got %v", value, string(comment.runes))
  }
}

func TestScanJavaCommentUnterminated(t *testing.T) {
  data := `/*comment`
  s := NewScanner([]byte(data))
  comment, err := s.scanForComment()
  value := "/*comment"
  if err != nil {
    t.Errorf("unexpected error %v", err)
  }
  if string(comment.runes) != value {
    t.Errorf("expected %v got %v", value, string(comment.runes))
  }
}

func TestNextToken(t *testing.T) {
  data := `a b c`
  s := NewScanner([]byte(data))
  for _, value := range([]string {
    "a",
    "b",
    "c",
  }) {
    token, err := s.NextToken()
    if err != nil {
      t.Errorf("unexpected error %v", err)
    }
    if string(token.runes) != value {
      t.Errorf("expected %v got %v", value, string(token.runes))
    }
  }
}

// next token on EOF returns io.EOF
func TestNextTokenEmpty(t *testing.T) {
  data := ``
  s := NewScanner([]byte(data))
  token, err := s.NextToken()
  if err != io.EOF {
    t.Errorf("unexpected error %v", err)
  }
  if token != nil {
    t.Errorf("unexpected token %v", token)
  }
}

func TestNextTokenMoreComplex(t *testing.T) {
  data := `-/* here's a multiline comment
     that spans multiple lines */

  // hello
  --- +changeset id:2
  CREATE TABLE exercise_logs;
  /*
  end*/
  -/`
  s := NewScanner([]byte(data))
  for _, value := range([]string {
    "-",
    "CREATE",
    "TABLE",
    "exercise_logs;",
    "-/",
  }) {
    token, err := s.NextToken()
    if err != nil {
      t.Errorf("unexpected error %v", err)
    }
    if string(token.runes) != value {
      t.Errorf("expected %v got %v", value, string(token.runes))
    }
  }
}
