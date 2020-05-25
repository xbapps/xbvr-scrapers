package helpers

import (
	"github.com/d5/tengo/v2"
	"github.com/thoas/go-funk"
)

var funkModule = map[string]tengo.Object{
	"contains_string": &tengo.UserFunction{
		Name:  "contains_string",
		Value: FuncASsSRS(funk.ContainsString),
	}, // contains_string([]s, substr) => bool
	"uniq_string": &tengo.UserFunction{
		Name:  "uniq_string",
		Value: FuncASsRSs(funk.UniqString),
	}, // uniq_string([]s) => []s
}
