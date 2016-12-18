package drift

import (
  "io"
  "bytes"
  "errors"
)

const (
  EOF = -(iota + 1)
  IDENT
  COMMENT
  WHITESPACE
)

type token struct {
  runes  []rune
  ttype  int
  offset int64  // starting offset (in runes not bytes) token was consumed at
  lineno int    // starting lineno the token was consumed from
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
// if we read EOF then we will return EOF, io.EOF,
// if it's an error we will return NUL, and the error
// otherwise we'll return the next rune (which can be 1-4 bytes)
// This method will also keep track of the current linenumber in the buffer
// and the byte offset in the buffer
func (s *scanner) next() ([]rune, error)  {
  ch, _, err := s.reader.ReadRune()
  if err != nil {
    if err != io.EOF {
      return []rune{rune(0)}, err
    }
    // ReadRune() will return io.EOF as an error when the EOF is reached
    // io.EOF is an error however, but we need a rune
    return []rune{EOF}, io.EOF
  }
  if ch == '\n' {
    s.lineno++
  }
  s.offset++  // move the read pointer
  return []rune{ch}, nil
}

// peek will return a rune slice for the next N runes
// we can't peek less than 0 runes, that will throw io.EOF
// if we peek past the EOF we'll return the runes upto EOF and return io.EOF
func (s *scanner) peek(count int) ([]rune, error) {
  offset := s.offset
  var runes []rune

  if count < 1 {
    return nil, io.EOF
  }

  for i := 1; i <= count; i++ {
    ch, _, err := s.reader.ReadRune()
    if err != nil {
      if err := s.seek(offset); err != nil {
        panic("can't unroll peek")
      }
      if err == io.EOF {
        // if we scan past the EOF return runes up to the EOF
        // this allows for less complex peek/peek logic elsewhere
        return runes, err
      } else {
        return nil, err
      }
    }
    runes = append(runes, ch)
  }
  // unroll - can use unreadrune because it requires having called readrune
  // as the previous call
  if err := s.seek(offset); err != nil {
    panic("can't unroll peek")
  }

  return runes, nil
}

// the 'offset' is in runes no bytes, this is because we want to
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
    //log.Println(string(runes))
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

// test for start of whitespace
func (s *scanner) isWhitespace(runes []rune) bool {
  if len(runes) >= 1 {
    return runes[0] == ' ' || runes[0] == '\t' || runes[0] == '\n' || runes[0] == '\r'
  }
  return false
}

// consumes whitespace at the current scanner offset
// attempting to consume whitespace from a non-whitespace rune errors
func (s *scanner) scanForWhitespace() (*token, error) {
  var rs []rune
  offset := s.offset
  lineno := s.lineno

  // back out if we are scanning starting from non whitespace
  runes, err := s.peek(1)
  if err != nil && err != io.EOF {
    s.offset = offset
    s.lineno = lineno
    return nil, err
  }

  for s.isWhitespace(runes) {
    // we can now start pulling the whitespace runes
    consumed, err := s.next()
    // if we get an error we reset the offset and back out
    if err != nil {
      if err == io.EOF {
        // we shouldn't get here, but if we do read EOF off of next() we'll
        // just break and return the rs
        break
      }
      s.offset = offset
      s.lineno = lineno
      return nil, err
    }

    rs = append(rs, consumed...)
    runes, err = s.peek(1)
    if err != nil && err != io.EOF {
      s.offset = offset
      s.lineno = lineno
      return nil, err
    }
  }
  return &token{runes:rs, ttype:WHITESPACE, offset:offset, lineno:lineno}, nil
}

// for our purposes, anything that isn't whitspace should be collapsed
// this allows us to reconstruct things with whitespace in the parser
func (s *scanner) isIdent(runes []rune) bool {
  if len(runes) >= 1 {
    return !s.isWhitespace(runes) && runes[0] != EOF
  }
  return false
}

// consumes ident runes at the current scanner offset
// attempting to consume an ident from a non-ident rune errors
func (s *scanner) scanForIdent() (*token, error) {
  var rs []rune
  offset := s.offset
  lineno := s.lineno

  // back out if we are scanning starting from non whitespace
  runes, err := s.peek(2)
  if err != nil && err != io.EOF {
    s.offset = offset
    s.lineno = lineno
    return nil, err
  }
  for s.isIdent(runes) && !s.isComment(runes) {
    // we can now start pulling the whitespace runes
    consumed, err := s.next()
    // if we get an error we reset the offset and back out
    if err != nil {
      if err == io.EOF {
        // we shouldn't get here, but if we do read EOF off of next() we'll
        // just break and return the rs
        break
      }
      s.offset = offset
      s.lineno = lineno
      return nil, err
    }

    rs = append(rs, consumed...)
    runes, err = s.peek(2)
    if err != nil && err != io.EOF {
      s.offset = offset
      s.lineno = lineno
      return nil, err
    }
  }
  return &token{runes:rs, ttype:IDENT, offset:offset, lineno:lineno}, nil
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
// attempting to consume a comment from a non-comment rune errors
func (s *scanner) scanForComment() (*token, error) {
  var rs []rune
  offset := s.offset
  lineno := s.lineno

  // peek 2 runes
  runes, err := s.peek(2)
  if err != nil && err != io.EOF {
    s.offset = offset
    s.lineno = lineno
    return nil, err
  }

  // consume a c++ style comment
  if (runes[0] == '/' && runes[1] == '/') {

    for len(runes) > 0 && runes[0] != '\n' && runes[0] != EOF {
      consumed, err := s.next()
      // if we get an error we reset the offset and back out
      if err != nil {
        if err == io.EOF {
          // we shouldn't get here, but if we do read EOF off of next() we'll
          // just break and return the rs
          break
        }
        s.offset = offset
        s.lineno = lineno
        return nil, err
      }

      rs = append(rs, consumed...)
      runes, err = s.peek(1)
      if err != nil && err != io.EOF {
        s.offset = offset
        s.lineno = lineno
        return nil, err
      }
    }
  // consume a sql style comment
  } else if (runes[0] == '-' && runes[1] == '-') {
    for len(runes) > 0 && runes[0] != '\n' && runes[0] != EOF {
      consumed, err := s.next()
      // if we get an error we reset the offset and back out
      if err != nil {
        if err == io.EOF {
          // we shouldn't get here, but if we do read EOF off of next() we'll
          // just break and return the rs
          break
        }
        s.offset = offset
        s.lineno = lineno
        return nil, err
      }

      rs = append(rs, consumed...)
      runes, err = s.peek(1)
      if err != nil && err != io.EOF {
        s.offset = offset
        s.lineno = lineno
        return nil, err
      }
    }
  //consume a java style comment
  } else if (runes[0] == '/' && runes[1] == '*') {
    // we exit if we peek EOF, err or the last 2 consumed runes are */
    for len(runes) > 0 && runes[0] != EOF {
      // exit this if we have consumed */ previously
      if len(rs) > 1 {
        if rs[len(rs)-2] == '*' && rs[len(rs)-1] == '/' {
          break
        }
      }
      consumed, err := s.next()
      // if we get an error we reset the offset and back out
      if err != nil {
        if err == io.EOF {
          // we shouldn't get here, but if we do read EOF off of next() we'll
          // just break and return the rs
          break
        }
        s.offset = offset
        s.lineno = lineno
        return nil, err
      }

      rs = append(rs, consumed...)
      runes, err = s.peek(1)
      if err != nil && err != io.EOF {
        s.offset = offset
        s.lineno = lineno
        return nil, err
      }
    }
  }
  return &token{runes:rs, ttype:COMMENT, offset:offset, lineno:lineno}, nil
}

// simple non-errorable test for has more tokens
// errors are interpreted as false - no token for you :(
func (s *scanner) HasMoreTokens() (bool) {
  runes, err := s.peek(1)
  if err != nil && len(runes) < 1 {
    return false
  }
  return true
}

func (s *scanner) NextToken() (*token, error) {
  var runes []rune
  runes, err := s.peek(2)
  // we can run into EOF if only 1 character is left
  if err != nil && len(runes) < 1 {
    return nil, err
  }
  // at this point runes can contain either 1 or 2 runes
  if s.isWhitespace(runes) {
    // skip whitespace
    _, err := s.scanForWhitespace()
    if err != nil {
      return nil, err
    }
    return s.NextToken()
  }
  if s.isComment(runes) {
    // skip comments
    _, err :=s.scanForComment()
    if err != nil {
      return nil, err
    }
    return s.NextToken()
  }
  if s.isIdent(runes) {
    ident, err := s.scanForIdent()
    if err != nil {
      return nil, err
    }
    return ident, nil
  }
  return nil, errors.New("Unknown token type")
}
