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
  for index, value := range(map[int]struct{
      value  string
      offset int64
      lineno int
  }{
    0:{"a", 0, 1},
    2:{"bcd", 2, 1},
    6:{"1.2E+10", 6, 1},
    25:{"!@#$%^&*()_+{}:\"<>?,./;'[]\\'\"", 25, 1},
    61:{"z", 61, 1},
    68:{"役立てることができることからきている" ,68, 3},
  }) {
    s.seek(int64(index))
    ident, err := s.scanForIdent()
    if err != nil {
      t.Errorf("unexpected error %v", err)
    }
    if string(ident.runes) != value.value {
      t.Errorf("expected %v got %v", value.value, string(ident.runes))
    }
    if ident.ttype != IDENT {
      t.Errorf("expected %v got %v", IDENT, ident.ttype)
    }
    if ident.offset != value.offset {
      t.Errorf("expected %v got %v", value.offset, ident.offset)
    }
    if ident.lineno != value.lineno {
      t.Errorf("expected %v got %v", value.lineno, ident.lineno)
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
  if token.ttype != IDENT {
    t.Errorf("expected %v got %v", IDENT, token.ttype)
  }
  if token.offset != 63 {
    t.Errorf("expected %v got %v", 63, token.offset)
  }
  if token.lineno != 2 {
    t.Errorf("expected %v got %v", 2, token.lineno)
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
  if token.ttype != IDENT {
    t.Errorf("expected %v got %v", IDENT, token.ttype)
  }
  if token.offset != 1 {
    t.Errorf("expected %v got %v", 1, token.offset)
  }
  if token.lineno != 1 {
    t.Errorf("expected %v got %v", 1, token.lineno)
  }
}

func TestScanWhitespace(t *testing.T) {
  data := `a bcd`+"\t"+`1.2E+10            !@#$%^&*()_+{}:"<>?,./;'[]\'"

    役立てることができることからきている`
  s := NewScanner([]byte(data))
  for index, value := range(map[int]struct{
      value  string
      offset int64
      lineno int
    }{
    1:{" ", 1, 1},
    5:{"\t", 5, 1},
    13:{"            ", 13, 1},
    54:{"\n\n    ", 54, 1},
  }) {
    s.seek(int64(index))
    ws, err := s.scanForWhitespace()
    if err != nil {
      t.Errorf("unexpected error %v", err)
    }
    if string(ws.runes) != value.value {
      t.Errorf("expected %v got %v", value.value, string(ws.runes))
    }
    if ws.ttype != WHITESPACE {
      t.Errorf("expected %v got %v", WHITESPACE, ws.ttype)
    }
    if ws.offset != value.offset {
      t.Errorf("expected %v got %v", value.offset, ws.offset)
    }
    if ws.lineno != value.lineno {
      t.Errorf("expected %v got %v", value.lineno, ws.lineno)
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
  if token.ttype != WHITESPACE {
    t.Errorf("expected %v got %v", WHITESPACE, token.ttype)
  }
  if token.offset != 1 {
    t.Errorf("expected %v got %v", 1, token.offset)
  }
  if token.lineno != 1 {
    t.Errorf("expected %v got %v", 1, token.lineno)
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
  for index, value := range(map[int]struct{
      value  string
      offset int64
      lineno int
    }{
    1:{"// here's a comment SELECT * FROM table", 1, 1},
    57:{"// you think", 57, 3},
    96:{"-- comment at the end of line", 96, 4},
    128:{"-- sql comment", 128, 5},
    145:{"/* and well\n\n  throw in\n  a\n  multiline comment */", 145, 6},
    198:{"//", 198, 11},
  }) {
    s.seek(int64(index))
    comment, err := s.scanForComment()
    if err != nil {
      t.Errorf("unexpected error %v", err)
    }
    if string(comment.runes) != value.value {
      t.Errorf("expected '%v' got '%v'", value.value, string(comment.runes))
    }
    if comment.ttype != COMMENT {
      t.Errorf("expected %v got %v", COMMENT, comment.ttype)
    }
    if comment.offset != value.offset {
      t.Errorf("expected %v got %v", value.offset, comment.offset)
    }
    if comment.lineno != value.lineno {
      t.Errorf("expected %v got %v", value.lineno, comment.lineno)
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
  if comment.ttype != COMMENT {
    t.Errorf("expected %v got %v", COMMENT, comment.ttype)
  }
  if comment.offset != 0 {
    t.Errorf("expected %v got %v", 0, comment.offset)
  }
  if comment.lineno != 1 {
    t.Errorf("expected %v got %v", 1, comment.lineno)
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
  if comment.ttype != COMMENT {
    t.Errorf("expected %v got %v", COMMENT, comment.ttype)
  }
  if comment.offset != 0 {
    t.Errorf("expected %v got %v", 0, comment.offset)
  }
  if comment.lineno != 1 {
    t.Errorf("expected %v got %v", 1, comment.lineno)
  }
}

func TestNextToken(t *testing.T) {
  data := `a b c`
  s := NewScanner([]byte(data))
  for _, value := range([]token {
    token{runes:[]rune{'a'}, ttype:IDENT, offset:0, lineno:1},
    token{runes:[]rune{'b'}, ttype:IDENT, offset:2, lineno:1},
    token{runes:[]rune{'c'}, ttype:IDENT, offset:4, lineno:1},
  }) {
    token, err := s.NextToken()
    if err != nil {
      t.Errorf("unexpected error %v", err)
    }
    if string(token.runes) != string(value.runes) {
      t.Errorf("expected %v got %v", string(value.runes), string(token.runes))
    }
    if token.ttype != value.ttype {
      t.Errorf("expected %v got %v", value.ttype, token.ttype)
    }
    if token.offset != value.offset {
      t.Errorf("expected %v got %v", value.offset, token.offset)
    }
    if token.lineno != value.lineno {
      t.Errorf("expected %v got %v", value.lineno, token.lineno)
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
  for _, value := range([]token {
    token{runes:[]rune("-"), ttype:IDENT, offset:0, lineno:1},
    token{runes:[]rune("CREATE"), ttype:IDENT, offset:101, lineno:6},
    token{runes:[]rune("TABLE"), ttype:IDENT, offset:108, lineno:6},
    token{runes:[]rune("exercise_logs;"), ttype:IDENT, offset:114, lineno:6},
    token{runes:[]rune("-/"), ttype:IDENT, offset:144, lineno:9},
  }) {
    token, err := s.NextToken()
    if err != nil {
      t.Errorf("unexpected error %v", err)
    }
    if string(token.runes) != string(value.runes) {
      t.Errorf("expected %v got %v", string(value.runes), string(token.runes))
    }
    if token.ttype != value.ttype {
      t.Errorf("expected %v got %v", value.ttype, token.ttype)
    }
    if token.offset != value.offset {
      t.Errorf("expected %v got %v", value.offset, token.offset)
    }
    if token.lineno != value.lineno {
      t.Errorf("expected %v got %v", value.lineno, token.lineno)
    }
  }
}
