package b

import (
	"fmt"
	"strconv"
	"strings"
)

const RET_REG = 0x0F

type CodeGen struct {
	asmLines  []string
	varMap    map[string]byte
	nextAddr  byte
	funcNodes map[string]*ASTNode
	funcOrder []string
	labelCnt  int
	paramBase byte
}

func NewCodeGen() *CodeGen {
	return &CodeGen{
		asmLines:  []string{},
		varMap:    make(map[string]byte),
		nextAddr:  0x10,
		funcNodes: make(map[string]*ASTNode),
		funcOrder: []string{},
		labelCnt:  0,
		paramBase: 0x20, // 从 0x20 开始分配参数和变量，避开 0x0F
	}
}

func (g *CodeGen) Generate(node *ASTNode) string {
	g.collect(node)
	g.asmLines = []string{}
	g.varMap = make(map[string]byte)
	g.nextAddr = 0x10
	g.labelCnt = 0

	g.genProgram(node)
	return strings.Join(g.asmLines, "\n")
}

func (g *CodeGen) collect(node *ASTNode) {
	if node == nil {
		return
	}
	switch node.Type {
	case NODE_PROGRAM:
		for _, s := range node.Body {
			g.collect(s)
		}
	case NODE_FUNCTION:
		g.funcOrder = append(g.funcOrder, node.Value)
		g.funcNodes[node.Value] = node
		for _, s := range node.Body {
			g.collect(s)
		}
	}
}

func (g *CodeGen) genProgram(node *ASTNode) {
	if node == nil {
		return
	}
	switch node.Type {
	case NODE_PROGRAM:
		if mainNode, ok := g.funcNodes["main"]; ok {
			g.emitLabel("main")
			g.genFunc(mainNode)
		}
		for _, name := range g.funcOrder {
			if name == "main" {
				continue
			}
			if fn, ok := g.funcNodes[name]; ok {
				g.emitLabel(name)
				g.genFunc(fn)
			}
		}
	}
}

func (g *CodeGen) genFunc(node *ASTNode) {
	if node == nil {
		return
	}

	paramAddr := g.paramBase
	for _, p := range node.Params {
		g.varMap[p] = paramAddr
		paramAddr++
	}

	for _, s := range node.Body {
		g.genDecl(s)
	}

	for _, s := range node.Body {
		g.genNode(s)
	}

	hasReturn := false
	if len(node.Body) > 0 {
		last := node.Body[len(node.Body)-1]
		if last.Type == NODE_RETURN {
			hasReturn = true
		}
	}

	if node.Value == "main" {
		g.emit("HALT")
	} else if !hasReturn {
		g.emit("RET")
	}
}

func (g *CodeGen) genDecl(node *ASTNode) {
	if node == nil || node.Type != NODE_DECL {
		return
	}
	for _, name := range strings.Split(node.Value, ",") {
		name = strings.TrimSpace(name)
		if _, ok := g.varMap[name]; !ok {
			g.varMap[name] = g.nextAddr
			g.nextAddr++
		}
	}
}

func (g *CodeGen) genNode(node *ASTNode) {
	if node == nil {
		return
	}

	switch node.Type {
	case NODE_DECL:
		// 声明已在 genFunc 中处理

	case NODE_ASSIGN:
		addr := g.getAddr(node.Value)

		if node.Right != nil && node.Right.Type == NODE_CALL {
			g.genCallWithAssign(node.Right, addr)
		} else {
			g.genNode(node.Right)
			g.emit("STORES %d", addr)
		}

	case NODE_BINOP:
		op := node.Value
		switch op {
		case "==", "!=", "<", ">", "<=", ">=":
			g.genCompare(node)
		default:
			g.genNode(node.Left)
			g.genNode(node.Right)
			switch op {
			case "+": g.emit("ADD")
			case "-": g.emit("SUB")
			case "*": g.emit("MUL")
			case "/": g.emit("DIV")
			}
		}

	case NODE_NUMBER:
		val, _ := strconv.Atoi(node.Value)
		if val < 0 || val > 255 {
			val = 0
		}
		g.emit("PUSH %d", val)

	case NODE_ID:
		addr := g.getAddr(node.Value)
		g.emit("LOAD %d", addr)

	case NODE_CALL:
		g.genCall(node)

	case NODE_IF:
		g.genIf(node)

	case NODE_WHILE:
		g.genWhile(node)

	case NODE_BLOCK:
		for _, s := range node.Body {
			g.genNode(s)
		}

	case NODE_RETURN:
		if node.Right != nil {
			g.genNode(node.Right)
			g.emit("STORES %d", RET_REG)
		}
		if node.Value != "main" { // main 不需要 RET
			g.emit("RET")
		}

	default:
		panic(fmt.Sprintf("未知节点类型: %s", node.Type))
	}
}

// 新增：专门处理带赋值的函数调用（关键修复）
func (g *CodeGen) genCallWithAssign(call *ASTNode, targetAddr byte) {
	name := call.Value

	if name == "getchar" {
		g.emit("LISTEN")
		g.emit("STORES %d", RET_REG)
		g.emit("LOAD %d", RET_REG)
		g.emit("STORES %d", targetAddr)
		return
	}

	// 普通函数调用
	for i, arg := range call.Args {
		g.genNode(arg)
		g.emit("STORES %d", g.paramBase+byte(i))
	}
	g.emit("CALL %s", name)
	g.emit("LOAD %d", RET_REG)
	g.emit("STORES %d", targetAddr)
}

// 新增：普通函数调用处理
func (g *CodeGen) genCall(call *ASTNode) {
	name := call.Value

	switch name {
	case "putchar":
		if len(call.Args) == 1 {
			g.genNode(call.Args[0])
			g.emit("PUTC")
			g.emit("POP") // 消耗掉栈上的值
		}
	case "putnumb":
		if len(call.Args) == 1 {
			g.genNode(call.Args[0])
			g.emit("PUT")
			g.emit("POP")
		}
	case "getchar":
		g.emit("LISTEN")
		g.emit("STORES %d", RET_REG)
		g.emit("LOAD %d", RET_REG) // 为后续可能的赋值准备
	default:
		for i, arg := range call.Args {
			g.genNode(arg)
			g.emit("STORES %d", g.paramBase+byte(i))
		}
		g.emit("CALL %s", name)
	}
}

func (g *CodeGen) genIf(node *ASTNode) {
	g.genNode(node.Cond)
	endLabel := g.newLabel()
	g.emit("JZ %s", endLabel)
	for _, s := range node.Body {
		g.genNode(s)
	}
	g.emitLabel(endLabel)
}

func (g *CodeGen) genWhile(node *ASTNode) {
	startLabel := g.newLabel()
	endLabel := g.newLabel()
	g.emitLabel(startLabel)
	g.genNode(node.Cond)
	g.emit("JZ %s", endLabel)
	for _, s := range node.Body {
		g.genNode(s)
	}
	g.emit("JMP %s", startLabel)
	g.emitLabel(endLabel)
}

func (g *CodeGen) genCompare(node *ASTNode) {
	g.genNode(node.Left)
	g.genNode(node.Right)

	trueLabel := g.newLabel()
	endLabel := g.newLabel()

	switch node.Value {
	case "==": g.emit("JEQ %s", trueLabel)
	case "!=": g.emit("JNEQ %s", trueLabel)
	case "<":  g.emit("JLT %s", trueLabel)
	case ">":  g.emit("JGT %s", trueLabel)
	case "<=": g.emit("JLE %s", trueLabel)
	case ">=": g.emit("JGE %s", trueLabel)
	}

	g.emit("PUSH 0")
	g.emit("JMP %s", endLabel)
	g.emitLabel(trueLabel)
	g.emit("PUSH 1")
	g.emitLabel(endLabel)
}

func (g *CodeGen) getAddr(name string) byte {
	if addr, ok := g.varMap[name]; ok {
		return addr
	}
	g.varMap[name] = g.nextAddr
	g.nextAddr++
	return g.varMap[name]
}

func (g *CodeGen) emit(format string, args ...interface{}) {
	line := fmt.Sprintf(format, args...)
	g.asmLines = append(g.asmLines, line)
}

func (g *CodeGen) emitLabel(name string) {
	g.asmLines = append(g.asmLines, name+":")
}

func (g *CodeGen) newLabel() string {
	g.labelCnt++
	return fmt.Sprintf("L%d", g.labelCnt)
}

