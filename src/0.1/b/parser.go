package b

import (
	"fmt"
)

type ASTNodeType string

const (
	NODE_PROGRAM   ASTNodeType = "PROGRAM"
	NODE_FUNCTION  ASTNodeType = "FUNCTION"
	NODE_DECL      ASTNodeType = "DECL"
	NODE_ASSIGN    ASTNodeType = "ASSIGN"
	NODE_BINOP     ASTNodeType = "BINOP"
	NODE_NUMBER    ASTNodeType = "NUMBER"
	NODE_ID        ASTNodeType = "ID"
	NODE_IF        ASTNodeType = "IF"
	NODE_WHILE     ASTNodeType = "WHILE"
	NODE_CALL      ASTNodeType = "CALL"
	NODE_BLOCK     ASTNodeType = "BLOCK"
	NODE_RETURN    ASTNodeType = "RETURN"
)

type ASTNode struct {
	Type   ASTNodeType
	Value  string
	Left   *ASTNode
	Right  *ASTNode
	Cond   *ASTNode
	Body   []*ASTNode
	Params []string
	Args   []*ASTNode
}

type Parser struct {
	tokens []Token
	pos    int
	tok    Token
}

func NewParser(tokens []Token) *Parser {
	p := &Parser{tokens: tokens, pos: 0}
	p.next()
	return p
}

func (p *Parser) next() {
	if p.pos < len(p.tokens) {
		p.tok = p.tokens[p.pos]
		p.pos++
	} else {
		p.tok = Token{TOKEN_EOF, "", 0, 0}
	}
}

func (p *Parser) peek() Token {
	if p.pos < len(p.tokens) {
		return p.tokens[p.pos]
	}
	return Token{TOKEN_EOF, "", 0, 0}
}

func (p *Parser) expect(typ TokenType, value string) {
	if p.tok.Type != typ || p.tok.Value != value {
		panic(fmt.Sprintf("期望 %s '%s'，但得到 %s '%s' at line %d", typ, value, p.tok.Type, p.tok.Value, p.tok.Line))
	}
	p.next()
}

func (p *Parser) Parse() *ASTNode {
	prog := &ASTNode{Type: NODE_PROGRAM, Body: []*ASTNode{}}
	for p.tok.Type != TOKEN_EOF {
		stmt := p.parseStatement()
		if stmt != nil {
			prog.Body = append(prog.Body, stmt)
		}
	}
	return prog
}

func (p *Parser) parseStatement() *ASTNode {
	switch p.tok.Type {
	case TOKEN_KEYWORD:
		switch p.tok.Value {
		case "auto":
			return p.parseDecl()
		case "if":
			return p.parseIf()
		case "while":
			return p.parseWhile()
		case "return":
			return p.parseReturn()
		default:
			panic(fmt.Sprintf("未知关键字 '%s' at line %d", p.tok.Value, p.tok.Line))
		}
	case TOKEN_ID:
		if p.peek().Type == TOKEN_SYMBOL && p.peek().Value == "(" {
			savePos, saveTok := p.pos, p.tok
			isFuncDef := false
			p.next()
			if p.tok.Type == TOKEN_SYMBOL && p.tok.Value == "(" {
				p.next()
				for p.tok.Type != TOKEN_EOF {
					if p.tok.Type == TOKEN_SYMBOL && p.tok.Value == ")" {
						p.next()
						if p.tok.Type == TOKEN_SYMBOL && p.tok.Value == "{" {
							isFuncDef = true
						}
						break
					}
					if p.tok.Type == TOKEN_ID {
						p.next()
						if p.tok.Type == TOKEN_SYMBOL && p.tok.Value == "," {
							p.next()
							continue
						}
						continue
					}
					break
				}
			}
			p.pos, p.tok = savePos, saveTok
			if isFuncDef {
				return p.parseFunctionDef()
			}
			node := p.parseCall()
			if p.tok.Type == TOKEN_SYMBOL && p.tok.Value == ";" {
				p.next()
			}
			return node
		}
		return p.parseAssign()
	case TOKEN_SYMBOL:
		if p.tok.Value == "{" {
			return p.parseBlock()
		}
		if p.tok.Value == ";" {
			p.next()
			return nil
		}
		panic(fmt.Sprintf("意外的符号 '%s' at line %d", p.tok.Value, p.tok.Line))
	default:
		if p.tok.Type == TOKEN_EOF {
			return nil
		}
		panic(fmt.Sprintf("意外的 Token: %s '%s' at line %d", p.tok.Type, p.tok.Value, p.tok.Line))
	}
}

func (p *Parser) parseFunctionDef() *ASTNode {
	name := p.tok.Value
	p.next()
	p.expect(TOKEN_SYMBOL, "(")
	params := []string{}
	if p.tok.Value != ")" {
		if p.tok.Type != TOKEN_ID {
			panic(fmt.Sprintf("期望参数名 at line %d", p.tok.Line))
		}
		params = append(params, p.tok.Value)
		p.next()
		for p.tok.Value == "," {
			p.next()
			if p.tok.Type != TOKEN_ID {
				panic(fmt.Sprintf("期望参数名 at line %d", p.tok.Line))
			}
			params = append(params, p.tok.Value)
			p.next()
		}
	}
	p.expect(TOKEN_SYMBOL, ")")
	p.expect(TOKEN_SYMBOL, "{")
	body := []*ASTNode{}
	for p.tok.Value != "}" && p.tok.Type != TOKEN_EOF {
		stmt := p.parseStatement()
		if stmt != nil {
			body = append(body, stmt)
		}
	}
	p.expect(TOKEN_SYMBOL, "}")
	return &ASTNode{Type: NODE_FUNCTION, Value: name, Params: params, Body: body}
}

func (p *Parser) parseDecl() *ASTNode {
	p.next()
	names := []string{}
	for {
		if p.tok.Type != TOKEN_ID {
			panic(fmt.Sprintf("期望标识符 at line %d", p.tok.Line))
		}
		names = append(names, p.tok.Value)
		p.next()
		if p.tok.Value == ";" {
			p.next()
			break
		}
		if p.tok.Value == "," {
			p.next()
			continue
		}
		panic(fmt.Sprintf("期望 ',' 或 ';' at line %d", p.tok.Line))
	}
	declStr := ""
	for i, n := range names {
		if i > 0 {
			declStr += ","
		}
		declStr += n
	}
	return &ASTNode{Type: NODE_DECL, Value: declStr}
}

func (p *Parser) parseAssign() *ASTNode {
	name := p.tok.Value
	p.next()
	p.expect(TOKEN_OP, "=")
	expr := p.parseExpr()
	if p.tok.Type == TOKEN_SYMBOL && p.tok.Value == ";" {
		p.next()
	}
	return &ASTNode{Type: NODE_ASSIGN, Value: name, Right: expr}
}

func (p *Parser) parseIf() *ASTNode {
	p.next()
	p.expect(TOKEN_SYMBOL, "(")
	cond := p.parseExpr()
	p.expect(TOKEN_SYMBOL, ")")
	body := p.parseBody()
	return &ASTNode{Type: NODE_IF, Cond: cond, Body: body}
}

func (p *Parser) parseWhile() *ASTNode {
	p.next()
	p.expect(TOKEN_SYMBOL, "(")
	cond := p.parseExpr()
	p.expect(TOKEN_SYMBOL, ")")
	body := p.parseBody()
	return &ASTNode{Type: NODE_WHILE, Cond: cond, Body: body}
}

func (p *Parser) parseReturn() *ASTNode {
	p.next()
	expr := p.parseExpr()
	if p.tok.Type == TOKEN_SYMBOL && p.tok.Value == ";" {
		p.next()
	}
	return &ASTNode{Type: NODE_RETURN, Right: expr}
}

func (p *Parser) parseCall() *ASTNode {
	name := p.tok.Value
	p.next()
	p.expect(TOKEN_SYMBOL, "(")
	args := []*ASTNode{}
	if p.tok.Value != ")" {
		args = append(args, p.parseExpr())
		for p.tok.Value == "," {
			p.next()
			args = append(args, p.parseExpr())
		}
	}
	p.expect(TOKEN_SYMBOL, ")")
	return &ASTNode{Type: NODE_CALL, Value: name, Args: args}
}

func (p *Parser) parseBlock() *ASTNode {
	p.next()
	body := []*ASTNode{}
	for p.tok.Value != "}" && p.tok.Type != TOKEN_EOF {
		stmt := p.parseStatement()
		if stmt != nil {
			body = append(body, stmt)
		}
	}
	p.expect(TOKEN_SYMBOL, "}")
	return &ASTNode{Type: NODE_BLOCK, Body: body}
}

func (p *Parser) parseBody() []*ASTNode {
	if p.tok.Value == "{" {
		block := p.parseBlock()
		return block.Body
	}
	stmt := p.parseStatement()
	if stmt == nil {
		return []*ASTNode{}
	}
	return []*ASTNode{stmt}
}

func (p *Parser) parseExpr() *ASTNode {
	left := p.parseTerm()
	for {
		op := p.tok.Value
		if op != "==" && op != "!=" && op != "<" && op != ">" && op != "<=" && op != ">=" {
			break
		}
		p.next()
		right := p.parseTerm()
		left = &ASTNode{Type: NODE_BINOP, Value: op, Left: left, Right: right}
	}
	return left
}

func (p *Parser) parseTerm() *ASTNode {
	left := p.parseFactor()
	for {
		op := p.tok.Value
		if op != "+" && op != "-" {
			break
		}
		p.next()
		right := p.parseFactor()
		left = &ASTNode{Type: NODE_BINOP, Value: op, Left: left, Right: right}
	}
	return left
}

func (p *Parser) parseFactor() *ASTNode {
	switch p.tok.Type {
	case TOKEN_NUMBER:
		val := p.tok.Value
		p.next()
		return &ASTNode{Type: NODE_NUMBER, Value: val}
	case TOKEN_ID:
		if p.peek().Type == TOKEN_SYMBOL && p.peek().Value == "(" {
			return p.parseCall()
		}
		val := p.tok.Value
		p.next()
		return &ASTNode{Type: NODE_ID, Value: val}
	case TOKEN_SYMBOL:
		if p.tok.Value == "(" {
			p.next()
			expr := p.parseExpr()
			p.expect(TOKEN_SYMBOL, ")")
			return expr
		}
		panic(fmt.Sprintf("意外的符号 '%s' at line %d", p.tok.Value, p.tok.Line))
	default:
		panic(fmt.Sprintf("期望数字、变量或 '(' at line %d", p.tok.Line))
	}
}
