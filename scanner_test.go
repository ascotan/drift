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
  ch, err := s.next()
  if err != nil {
    t.Errorf("next: unexpected error %v", err)
  }
  if ch != '风' {
    t.Errorf("next: expected %v got %v", '风', ch)
  }
  if s.offset != 3 {  // since this is in the cjk plane we have a 3 byte rune
    t.Errorf("offset: expected %v got %v", 3, s.offset)
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
  ch, err := s.next()
  if err != io.EOF {
    t.Errorf("next: unexpected error %v", err)
  }
  if ch != EOF {
    t.Errorf("next: expected %v got %v", EOF, ch)
  }

  // call again after reading EOF - we should keep getting EOF
  ch, err = s.next()
  if err != io.EOF {
    t.Errorf("next: unexpected error %v", err)
  }
  if ch != EOF {
    t.Errorf("next: expected %v got %v", EOF, ch)
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
  ch, err := s.peek()
  if err != nil {
    t.Errorf("seek: unexpected error %v", err)
  }
  if ch != 'e' {
    t.Errorf("seek: expected %v got %v", 'e', ch)
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
  ch, err := s.peek()
  if err != nil {
    t.Errorf("seek: unexpected error %v", err)
  }
  if ch != 'w' {
    t.Errorf("seek: expected %v got %v", 'w', ch)
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
  data := `风`
  s := NewScanner([]byte(data))
  ch, err := s.peek()
  if err != nil {
    t.Errorf("peek: unexpected error %v", err)
  }
  if ch != '风' {
    t.Errorf("peek: expected %v got %v", '风', ch)
  }
  if s.offset != 0 {  // peeking should not fast forward the offset
    t.Errorf("peek: expected %v got %v", 0, s.offset)
  }

  // peeking again should give me the same character
  ch, err = s.peek()
  if err != nil {
    t.Errorf("peek: unexpected error %v", err)
  }
  if ch != '风' {
    t.Errorf("peek: expected %v got %v", '风', ch)
  }
}

func TestPeekEOF(t *testing.T) {
  data := `a`
  s := NewScanner([]byte(data))
  err := s.seek(1)
  if err != nil {
    t.Errorf("peek: unexpected error %v", err)
  }
  ch, err := s.peek()
  if err != io.EOF && ch != EOF {
    t.Error("expected io.EOF to be thrown and ch set to EOF")
  }
  if s.offset != 1 {  // peeking should not fast forward the offset
    t.Errorf("peek: expected %v got %v", 1, s.offset)
  }
}

func TestMultiPeek(t *testing.T) {
  data := `风雷动`
  s := NewScanner([]byte(data))

  for index, value := range(map[int]string{
    1:"风",
    2:"风雷",
    3:"风雷动",
  }) {
    runes, err := s.multipeek(index)
    if err != nil {
      t.Errorf("multipeek: unexpected error %v", err)
    }
    if string(runes) != value {
      t.Errorf("multipeek: expected %v got %v", value, runes)
    }
  }

  // multipeek past the EOF but less then bytes in the reader
  // there are 9 bytes in this reader because of the ckj unicode charcters
  offset := s.offset
  runes, err := s.multipeek(7)
  if err != io.EOF {
    t.Error("multipeek: expected io.EOF to be throw")
  }
  if len(runes) != 0 {
    t.Errorf("multipeek: expected rune length to be %v got %v", 0, len(runes))
  }
  if s.offset != offset {
    t.Errorf("multipeek: expected offset to be %v got %v", offset, s.offset)
  }

  // way way past the EOF
  runes, err = s.multipeek(1000)
  if err != io.EOF {
    t.Error("multipeek: expected io.EOF to be throw")
  }
  if len(runes) != 0 {
    t.Errorf("multipeek: expected rune length to be %v got %v", 0, len(runes))
  }
  if s.offset != offset {
    t.Errorf("multipeek: expected offset to be %v got %v", offset, s.offset)
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
  for value, expected := range(map[rune]bool {
    ' '  : true,
    '\n' : true,
    '\r' : true,
    '\t' : true,
    'a'  : false,
  }) {
    if s.isWhitespace(value) != expected {
      t.Errorf("isWhitespace: expected %v got %v", expected, s.isWhitespace(value))
    }
  }
}

func TestIsIdent(t *testing.T) {
  data := `a`
  s := NewScanner([]byte(data))
  for value, expected := range(map[rune]bool {
    'a'  : true,
    'A' : true,
    '0' : true,
    '.' : true,
    '呼' : true,
    'б' : true,
    ' ' : false,
  }) {
    if s.isIdent(value) != expected {
      t.Errorf("isIdent: expected %v got %v", expected, s.isIdent(value))
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
      t.Errorf("scanIdent: unexpected error %v", err)
    }
    if string(ident.runes) != value {
      t.Errorf("scanIdent: expected %v got %v", value, string(ident.runes))
    }
    if ident.width != len([]byte(value)) {
      t.Errorf("scanIdent: expected %v got %v", len([]byte(value)), ident.width)
    }
  }
}

func TestScanIdentBadStart(t *testing.T) {
  data := `a bcd 1.2E+10            !@#$%^&*()_+{}:"<>?,./;'[]\'"       z

    役立てることができることからきている`
  s := NewScanner([]byte(data))

  s.seek(63) // seeking into whitespace
  _, err := s.scanForIdent()
  if err == nil {
    t.Errorf("scanIdent: expected error to be thrown")
  }
}

func TestScanIdentEOF(t *testing.T) {
  data := `a`
  s := NewScanner([]byte(data))

  s.seek(1) // seeking to EOF
  _, err := s.scanForIdent()
  if err == nil {
    t.Errorf("scanIdent: expected error to be thrown")
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
    if ws.width != len([]byte(value)) {
      t.Errorf("expected %v got %v", len([]byte(value)), ws.width)
    }
  }
}

func TestScanWhitespaceEOF(t *testing.T) {
  data := `a`
  s := NewScanner([]byte(data))

  s.seek(1) // seeking to EOF
  _, err := s.scanForWhitespace()
  if err == nil {
    t.Errorf("expected error to be thrown")
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
    198:"//",
  }) {
    s.seek(int64(index))
    comment, err := s.scanForComment()
    if err != nil {
      t.Errorf("unexpected error %v", err)
    }
    if string(comment.runes) != value {
      t.Errorf("expected %v got %v", value, string(comment.runes))
    }
    if comment.width != len([]byte(value)) {
      t.Errorf("expected %v got %v", len([]byte(value)), comment.width)
    }
  }
}
