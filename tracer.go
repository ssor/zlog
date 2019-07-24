package zlog

import (
	"bytes"
	"container/list"
	"encoding/json"
	"fmt"
)

var (
	chains map[int]*TraceChain

	colors map[int]bool
)

func init() {
	chains = make(map[int]*TraceChain)
	colors = make(map[int]bool)

	colors[red] = false
	colors[green] = false
	colors[yellow] = false
	colors[blue] = false
}

func TraceWithStructDefault(obj interface{}) {
	Trace(0, nil, obj, nil)
}

func TraceFiledsDefault(fields Fields) {
	Trace(0, fields, nil, nil)
}

func TraceArgsDefault(args ...interface{}) {
	Trace(0, nil, nil, args)
}

func TraceDefault(fields Fields, obj interface{}, args ...interface{}) {
	index := 0
	chain, ok := chains[index]
	if ok == false {
		chain = generateNewChain()
		chains[index] = chain
	}

	block := newTraceBlock(chain.color, args, obj, fields)
	chain.addBlock(block)
}

func TraceWithStruct(index int, obj interface{}) {
	Trace(index, nil, obj, nil)
}

func TraceFileds(index int, fields Fields) {
	Trace(index, fields, nil, nil)
}

func TraceArgs(index int, args ...interface{}) {
	Trace(index, nil, nil, args)
}

func Trace(index int, fields Fields, obj interface{}, args ...interface{}) {
	chain, ok := chains[index]
	if ok == false {
		chain = generateNewChain()
		chains[index] = chain
	}

	block := newTraceBlock(chain.color, args, obj, fields)
	chain.addBlock(block)
}

func generateNewChain() *TraceChain {
	chain := newTraceChain(nocolor)
	for color, used := range colors {
		if used == false {
			colors[color] = true
			chain.color = color
		}
	}
	return chain
}

func TraceEndDefault() {
	chain, ok := chains[0]
	if ok == false {
		return
	}
	for block := chain.blocks.Front(); block != nil; block = block.Next() {
		printBlock(block.Value.(*TraceBlock))
	}
}

func TraceEnd(index int) {
	chain, ok := chains[index]
	if ok == false {
		return
	}
	for block := chain.blocks.Front(); block != nil; block = block.Next() {
		printBlock(block.Value.(*TraceBlock))
	}
}

func newTraceChain(color int) *TraceChain {
	return &TraceChain{
		blocks: list.New(),
		color:  color,
	}
}

func (tc *TraceChain) addBlock(block *TraceBlock) {
	tc.blocks.PushBack(block)
}

type TraceChain struct {
	blocks *list.List
	color  int
}

func newTraceBlock(color int, args []interface{}, obj interface{}, fields Fields) *TraceBlock {
	tb := TraceBlock{
		Fields: fields,
		Obj:    obj,
		color:  color,
		args:   args,
	}
	return &tb
}

type TraceBlock struct {
	Fields Fields
	Obj    interface{}
	args   []interface{}
	color  int
}

func printBlock(block *TraceBlock) {
	b := bytes.NewBuffer([]byte{})
	message := fmt.Sprint(block.args...)
	fmt.Fprintf(b, "\x1b[%dm msg: %-44s \x1b[0m", block.color, message)

	for k, v := range block.Fields {
		value := fmt.Sprintf("%+v", v)
		if len(value) > 128 {
			value = value[:128] + "..."
		}
		fmt.Fprintf(b, "\n     \x1b[%dm- %-8s = %+v \x1b[0m", block.color, k, value)
	}

	jsonRaw, err := json.Marshal(block.Obj)
	if err != nil {
		panic(err)
	}
	if jsonRaw != nil {
		fmt.Fprintf(b, "\x1b[%dm \n%s \x1b[0m", block.color, prettyJSON(jsonRaw))
	}
	fmt.Fprintf(b, "\n---------------------------------------------------\n")

	logger := StandardLogger()

	logger.Out.Write(b.Bytes())
}
