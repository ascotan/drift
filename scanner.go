package drift

import (
  "io"
  "bytes"
  "errors"
)

const (
  EOF = -(iota + 1)
  UNKNOWN
)

type token struct {
  runes []rune
  width int
}

type scanner struct {
  reader *bytes.Reader
  offset int64
  lineno int
}
func NewScanner(b []byte) (*scanner) {
  return &scanner{
    reader: bytes.NewReader(b),
    offset: 0,
    lineno: 1,
  }
}

// This method reads the next rune from the buffer
func (s *scanner) next() (rune, error)  {
  ch, size, err := s.reader.ReadRune()
  if err != nil {
    if err != io.EOF {
      return rune(0), err
    }
    // ReadRune() will return io.EOF as an error when the EOF is reached
    // io.EOF is an error however, but we need a rune
    return EOF, io.EOF
  }
  if ch == '\n' {
    s.lineno++
  }
  s.offset += int64(size)  // move the read pointer
  return ch, nil
}

func (s *scanner) peek() (rune, error) {
  offset := s.offset
  ch, _, err := s.reader.ReadRune()
  if err != nil {
    if err := s.seek(offset); err != nil {
      panic("can't unroll peek")
    }
    return EOF, err
  }
  err = s.reader.UnreadRune()
  if err != nil {
    if err := s.seek(offset); err != nil {
      panic("can't unroll peek")
    }
    return EOF, err
  }
  return ch, nil
}

func (s *scanner) multipeek(count int) ([]rune, error) {
  offset := s.offset
  var runes []rune

  // O(n) of for mutlipeek is the byte size of the reader data
  if count < 1 || (int64(count) > s.reader.Size()) {
    return nil, io.EOF
  }

  for i := 1; i <= count; i++ {
    ch, _, err := s.reader.ReadRune()
    if err != nil {
      if err := s.seek(offset); err != nil {
        panic("can't unroll multipeek")
      }
      return nil, err
    }
    runes = append(runes, ch)
  }
  // unroll - can use unreadrune because it requires having called readrune
  // as the previous call
  if err := s.seek(offset); err != nil {
    panic("can't unroll multipeek")
  }
  return runes, nil
}

// we do it this way rather than bytes.Reader.Seek because we want to
// get the lineno
// if you attempt to seek past the end of file, we'll just fast forward to EOF
// and then exit
func (s *scanner) seek(offset int64) (error) {
  oldoffset := s.offset
  oldlineno := s.lineno

  // fast fail on negative seeks
  if offset < 0 {
    return io.EOF
  }
  // reset the buffer pointer to the start
  _, err := s.reader.Seek(0, io.SeekStart)
  if err != nil {
    return err
  }
  s.offset = int64(0)
  s.lineno = 1

  // now run next() for each offset
  for i := int64(0); i < offset; i++ {
    _, err := s.next()
    // if we hit an error running next() we'll revert to the pre-seek position
    // and return the error
    if err != nil {
      if err == io.EOF {
        // we've seeked to or past the EOF
        return nil
      }
      s.offset = oldoffset
      s.lineno = oldlineno
      return err
    }
  }
  return nil
}

func (s *scanner) HasMoreTokens() (bool) {
  ch, err := s.peek()
  if ch == EOF || err != nil {
    return false
  }
  return true
}

func (s *scanner) NextToken() (*token, error) {
  r, err := s.peek()
  if err != nil {
    return nil, err
  }
  // skip whitespace
  if s.isWhitespace(r) {
    s.scanForWhitespace()
    return s.NextToken()
  }
  // if s.isComment(r) {
  //   // skip the comment
  //   s.scanForComment()
  //   return s.NextToken()
  // }
  if s.isIdent(r) {
    return s.scanForIdent()
  }
  return &token{runes:[]rune{r}}, nil
}

// test for start of whitespace
func (s *scanner) isWhitespace(ch rune) bool {
  return ch == ' ' || ch == '\t' || ch == '\n' || ch == '\r'
}

// consumes whitespace at the current scanner offset
func (s *scanner) scanForWhitespace() (*token, error) {
  var rs []rune
  offset := s.offset
  width := 0

  // back out if we are scanning starting from non whitespace
  ch, err := s.peek()
  if err != nil{
    return nil, err
  }
  if !s.isWhitespace(ch) {
    return nil, errors.New("cannot scan whitespace from non-whitespace rune")
  }

  // we can now start pulling the whitespace runes
  ch, err = s.next()
  // if we get an error we reset the offset and back out
  if err != nil {
    s.offset = offset
    return nil, err
  }
  for s.isWhitespace(ch) {
    rs = append(rs, ch)
    width += int(s.offset - (offset + int64(width)))
    var err error
    ch, err = s.next()
    if err != nil {
      // if the next character is EOF lets end this
      if err == io.EOF {
        break
      }
      s.offset = offset
      return nil, err
    }
  }
  return &token{runes:rs, width:width}, nil
}

// for our purposes, anything that isn't whitspace should be collapsed
// this allows us to reconstruct things with whitespace in the parser
func (s *scanner) isIdent(ch rune) bool {
  return !s.isWhitespace(ch) && ch != EOF
}

// consumes ident runes at the current scanner offset
func (s *scanner) scanForIdent() (*token, error) {
  var rs []rune
  offset := s.offset
  width := 0

  // back out if we are scanning starting from a non-ident
  ch, err := s.peek()
  if err != nil{
    return nil, err
  }
  if !s.isIdent(ch) {
    return nil, errors.New("cannot scan ident from non-ident rune")
  }

  // we can now start pulling the ident
  ch, err = s.next()
  // if we get an error we reset the offset and back out
  if err != nil {
    s.offset = offset
    return nil, err
  }
  for s.isIdent(ch) {
    rs = append(rs, ch)
    width += int(s.offset - (offset + int64(width)))
    var err error
    ch, err = s.next()
    if err != nil {
      // if the next character is EOF lets end this
      if err == io.EOF {
        break
      }
      s.offset = offset
      return nil, err
    }
  }
  return &token{runes:rs, width:width}, nil
}

// test for comments staring with '//' and '/*' and '--'
func (s *scanner) isComment(runes []rune) bool {
  if len(runes) >= 2 {
    if runes[0] == '/' && runes[1] == '/' ||
       runes[0] == '/' && runes[1] == '*' ||
       runes[0] == '-' && runes[1] == '-' {
      return true
    }
  }
  return false
}

// consumes comments at the current scanner offset
func (s *scanner) scanForComment() (*token, error) {
  var rs []rune
  offset := s.offset
  width := 0

  // back out if we are scanning starting from a non-comment
  runes, err := s.multipeek(2)
  if err != nil{
    return nil, err
  }
  if !s.isComment(runes) {
    return nil, errors.New("cannot scan ident from non-comment runes")
  }

  // scan a c++ style
  if (runes[0] == '/' && runes[1] == '/') {
    lineno := s.lineno
    ch, err := s.next()
    // if we get an error we reset the offset and back out
    if err != nil {
      s.offset = offset
      return nil, err
    }
    for lineno == s.lineno {
      rs = append(rs, ch)
      width += int(s.offset - (offset + int64(width)))
      var err error
      ch, err = s.next()
      if err != nil {
        // if the next character is EOF lets end this
        if err == io.EOF {
          break
        }
        s.offset = offset
        return nil, err
      }
    }
  }

  // // parse a java style comment
  // // scan until we hit a */
  // } else if ch == '/' && next == '*' {
  //   for ch != '*' || next != '/' {
  //     rs = append(rs, ch)
  //     ch, err = s.next()
  //     if err != nil {
  //       return nil, err
  //     }
  //     next, err = s.peek();
  //     if err != nil {
  //       return nil, err
  //     }
  //   }
  //   // grab the closing characters
  //   rs = append(rs, ch)
  //   ch, err = s.next()
  //   if err != nil {
  //     return nil, err
  //   }
  //   rs = append(rs, ch)
  // }
  return &token{runes:rs, width:width}, nil
}
