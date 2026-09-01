package pylit

import (
	"fmt"
	"strconv"
	"strings"
	"unicode"
)

type parser struct {
	src []rune
	pos int
}

func Parse(src []byte) (any, error) {
	p := &parser{src: []rune(string(src))}
	p.skipTrivia()
	if p.pos >= len(p.src) {
		return nil, fmt.Errorf("empty input")
	}
	v, err := p.parseValue()
	if err != nil {
		return nil, err
	}
	p.skipTrivia()
	if p.pos < len(p.src) {
		return nil, fmt.Errorf("unexpected trailing content at offset %d", p.pos)
	}
	return v, nil
}

func (p *parser) skipTrivia() {
	for p.pos < len(p.src) {
		c := p.src[p.pos]
		switch {
		case c == ' ' || c == '\t' || c == '\n' || c == '\r' || c == '\f' || c == '\v':
			p.pos++
		case c == '#':
			for p.pos < len(p.src) && p.src[p.pos] != '\n' {
				p.pos++
			}
		case c == '\\' && p.pos+1 < len(p.src) && p.src[p.pos+1] == '\n':
			p.pos += 2
		default:
			return
		}
	}
}

func (p *parser) parseValue() (any, error) {
	p.skipTrivia()
	if p.pos >= len(p.src) {
		return nil, fmt.Errorf("unexpected end of input")
	}
	c := p.src[p.pos]
	switch {
	case c == '{':
		return p.parseBrace()
	case c == '[':
		return p.parseSequence(']')
	case c == '(':
		return p.parseSequence(')')
	case c == '-' || (c >= '0' && c <= '9'):
		return p.parseNumber()
	default:
		if _, _, ok := p.atString(); ok {
			return p.parseStringConcat()
		}
		return p.parseIdent()
	}
}

func (p *parser) parseBrace() (any, error) {
	p.pos++
	p.skipTrivia()
	if p.pos < len(p.src) && p.src[p.pos] == '}' {
		p.pos++
		return map[string]any{}, nil
	}
	first, err := p.parseValue()
	if err != nil {
		return nil, err
	}
	p.skipTrivia()
	if p.pos >= len(p.src) {
		return nil, fmt.Errorf("unterminated brace literal")
	}
	if p.src[p.pos] == ':' {
		return p.parseDictRest(first)
	}
	return p.parseSetRest(first)
}

func (p *parser) parseDictRest(firstKey any) (any, error) {
	out := map[string]any{}
	key, ok := firstKey.(string)
	if !ok {
		return nil, fmt.Errorf("dict key must be a string at offset %d", p.pos)
	}
	for {
		p.pos++
		val, err := p.parseValue()
		if err != nil {
			return nil, err
		}
		out[key] = val
		p.skipTrivia()
		if p.pos >= len(p.src) {
			return nil, fmt.Errorf("unterminated dict")
		}
		switch p.src[p.pos] {
		case ',':
			p.pos++
			p.skipTrivia()
			if p.pos < len(p.src) && p.src[p.pos] == '}' {
				p.pos++
				return out, nil
			}
		case '}':
			p.pos++
			return out, nil
		default:
			return nil, fmt.Errorf("expected ',' or '}' at offset %d", p.pos)
		}
		if _, _, ok := p.atString(); !ok {
			return nil, fmt.Errorf("dict key must be a string at offset %d", p.pos)
		}
		k, err := p.parseString()
		if err != nil {
			return nil, err
		}
		key = k
		p.skipTrivia()
		if p.pos >= len(p.src) || p.src[p.pos] != ':' {
			return nil, fmt.Errorf("expected ':' after key %q", key)
		}
	}
}

func (p *parser) parseSetRest(first any) (any, error) {
	out := []any{first}
	for {
		p.skipTrivia()
		if p.pos >= len(p.src) {
			return nil, fmt.Errorf("unterminated set")
		}
		switch p.src[p.pos] {
		case ',':
			p.pos++
			p.skipTrivia()
			if p.pos < len(p.src) && p.src[p.pos] == '}' {
				p.pos++
				return out, nil
			}
			v, err := p.parseValue()
			if err != nil {
				return nil, err
			}
			out = append(out, v)
		case '}':
			p.pos++
			return out, nil
		default:
			return nil, fmt.Errorf("expected ',' or '}' at offset %d", p.pos)
		}
	}
}

func (p *parser) parseSequence(close rune) (any, error) {
	p.pos++
	out := []any{}
	trailingComma := false
	for {
		p.skipTrivia()
		if p.pos >= len(p.src) {
			return nil, fmt.Errorf("unterminated sequence")
		}
		if p.src[p.pos] == close {
			p.pos++
			if close == ')' && len(out) == 1 && !trailingComma {
				return out[0], nil
			}
			return out, nil
		}
		v, err := p.parseValue()
		if err != nil {
			return nil, err
		}
		out = append(out, v)
		p.skipTrivia()
		if p.pos >= len(p.src) {
			return nil, fmt.Errorf("unterminated sequence")
		}
		switch p.src[p.pos] {
		case ',':
			p.pos++
			trailingComma = true
		case close:
			p.pos++
			if close == ')' && len(out) == 1 {
				return out[0], nil
			}
			return out, nil
		default:
			return nil, fmt.Errorf("expected ',' or %q at offset %d", string(close), p.pos)
		}
	}
}

func (p *parser) parseStringConcat() (any, error) {
	s, err := p.parseString()
	if err != nil {
		return nil, err
	}
	var b strings.Builder
	b.WriteString(s)
	for {
		save := p.pos
		p.skipTrivia()
		if p.pos < len(p.src) && p.src[p.pos] == '+' {
			p.pos++
			p.skipTrivia()
			if _, _, ok := p.atString(); !ok {
				return nil, fmt.Errorf("expected string after '+' at offset %d", p.pos)
			}
			s2, err := p.parseString()
			if err != nil {
				return nil, err
			}
			b.WriteString(s2)
			continue
		}
		if _, _, ok := p.atString(); ok {
			s2, err := p.parseString()
			if err != nil {
				return nil, err
			}
			b.WriteString(s2)
			continue
		}
		p.pos = save
		return b.String(), nil
	}
}

func (p *parser) atString() (prefixLen int, raw bool, ok bool) {
	i := p.pos
	n := 0
	rawFlag := false
	for n < 2 && i < len(p.src) {
		switch p.src[i] {
		case 'r', 'R':
			rawFlag = true
		case 'u', 'U', 'b', 'B', 'f', 'F':
		default:
			if i < len(p.src) && (p.src[i] == '\'' || p.src[i] == '"') {
				return n, rawFlag, true
			}
			return 0, false, false
		}
		i++
		n++
	}
	if i < len(p.src) && (p.src[i] == '\'' || p.src[i] == '"') {
		return n, rawFlag, true
	}
	return 0, false, false
}

func (p *parser) parseString() (string, error) {
	prefixLen, raw, ok := p.atString()
	if !ok {
		return "", fmt.Errorf("expected string at offset %d", p.pos)
	}
	p.pos += prefixLen
	quote := p.src[p.pos]
	triple := false
	if p.pos+2 < len(p.src) && p.src[p.pos+1] == quote && p.src[p.pos+2] == quote {
		triple = true
		p.pos += 3
	} else {
		p.pos++
	}
	var b strings.Builder
	for p.pos < len(p.src) {
		c := p.src[p.pos]
		if c == '\\' {
			if raw {
				b.WriteRune('\\')
				p.pos++
				continue
			}
			if p.pos+1 >= len(p.src) {
				return "", fmt.Errorf("dangling escape in string")
			}
			esc := p.src[p.pos+1]
			p.pos += 2
			switch esc {
			case 'n':
				b.WriteRune('\n')
			case 't':
				b.WriteRune('\t')
			case 'r':
				b.WriteRune('\r')
			case '\\':
				b.WriteRune('\\')
			case '\'':
				b.WriteRune('\'')
			case '"':
				b.WriteRune('"')
			case '0':
				b.WriteRune('\x00')
			case '\n':
			case 'x':
				r, err := p.readHex(2)
				if err != nil {
					return "", err
				}
				b.WriteRune(r)
			case 'u':
				r, err := p.readHex(4)
				if err != nil {
					return "", err
				}
				b.WriteRune(r)
			default:
				b.WriteRune('\\')
				b.WriteRune(esc)
			}
			continue
		}
		if triple {
			if c == quote && p.pos+2 < len(p.src) && p.src[p.pos+1] == quote && p.src[p.pos+2] == quote {
				p.pos += 3
				return b.String(), nil
			}
		} else {
			if c == quote {
				p.pos++
				return b.String(), nil
			}
			if c == '\n' {
				return "", fmt.Errorf("newline in single-line string at offset %d", p.pos)
			}
		}
		b.WriteRune(c)
		p.pos++
	}
	return "", fmt.Errorf("unterminated string")
}

func (p *parser) readHex(n int) (rune, error) {
	if p.pos+n > len(p.src) {
		return 0, fmt.Errorf("truncated hex escape at offset %d", p.pos)
	}
	v, err := strconv.ParseInt(string(p.src[p.pos:p.pos+n]), 16, 32)
	if err != nil {
		return 0, fmt.Errorf("invalid hex escape at offset %d: %w", p.pos, err)
	}
	p.pos += n
	return rune(v), nil
}

func (p *parser) parseNumber() (any, error) {
	start := p.pos
	if p.src[p.pos] == '-' {
		p.pos++
	}
	isFloat := false
	for p.pos < len(p.src) {
		c := p.src[p.pos]
		switch {
		case c >= '0' && c <= '9':
			p.pos++
		case c == '_':
			p.pos++
		case c == '.' || c == 'e' || c == 'E':
			isFloat = true
			p.pos++
		case (c == '+' || c == '-') && p.pos > start && (p.src[p.pos-1] == 'e' || p.src[p.pos-1] == 'E'):
			p.pos++
		default:
			goto done
		}
	}
done:
	text := strings.ReplaceAll(string(p.src[start:p.pos]), "_", "")
	if isFloat {
		f, err := strconv.ParseFloat(text, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid float %q: %w", text, err)
		}
		return f, nil
	}
	n, err := strconv.ParseInt(text, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid int %q: %w", text, err)
	}
	return n, nil
}

func (p *parser) parseIdent() (any, error) {
	start := p.pos
	for p.pos < len(p.src) {
		c := p.src[p.pos]
		if unicode.IsLetter(c) || unicode.IsDigit(c) || c == '_' {
			p.pos++
			continue
		}
		break
	}
	word := string(p.src[start:p.pos])
	switch word {
	case "True":
		return true, nil
	case "False":
		return false, nil
	case "None":
		return nil, nil
	case "":
		return nil, fmt.Errorf("unexpected character %q at offset %d", string(p.src[p.pos]), p.pos)
	default:
		return nil, fmt.Errorf("unsupported expression %q at offset %d", word, start)
	}
}
