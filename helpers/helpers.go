package helpers

import (
	"fmt"

	"github.com/d5/tengo/v2"
)

// HelperModules are scraper helper library modules.
var HelperModules = map[string]map[string]tengo.Object{
	"funk": funkModule,
}

// AllModuleNames returns a list of all helper module names.
func AllModuleNames() []string {
	var names []string
	for name := range HelperModules {
		names = append(names, name)
	}
	return names
}

// GetModuleMap returns the module map that includes all modules
// for the given module names.
func GetModuleMap(names ...string) *tengo.ModuleMap {
	modules := tengo.NewModuleMap()
	for _, name := range names {
		if mod := HelperModules[name]; mod != nil {
			modules.AddBuiltinModule(name, mod)
		}
	}
	return modules
}

// FuncASsSRS transform a function of 'func([]string, string) bool' signature
// into CallableFunc type.
// TODO(jrebey): Move this into tengo
func FuncASsSRS(fn func([]string, string) bool) tengo.CallableFunc {
	return func(args ...tengo.Object) (tengo.Object, error) {
		if len(args) != 2 {
			return nil, tengo.ErrWrongNumArguments
		}
		var ss1 []string
		switch arg0 := args[0].(type) {
		case *tengo.Array:
			for idx, a := range arg0.Value {
				as, ok := tengo.ToString(a)
				if !ok {
					return nil, tengo.ErrInvalidArgumentType{
						Name:     fmt.Sprintf("first[%d]", idx),
						Expected: "string(compatible)",
						Found:    a.TypeName(),
					}
				}
				ss1 = append(ss1, as)
			}
		case *tengo.ImmutableArray:
			for idx, a := range arg0.Value {
				as, ok := tengo.ToString(a)
				if !ok {
					return nil, tengo.ErrInvalidArgumentType{
						Name:     fmt.Sprintf("first[%d]", idx),
						Expected: "string(compatible)",
						Found:    a.TypeName(),
					}
				}
				ss1 = append(ss1, as)
			}
		default:
			return nil, tengo.ErrInvalidArgumentType{
				Name:     "first",
				Expected: "array",
				Found:    args[0].TypeName(),
			}
		}
		s2, ok := tengo.ToString(args[1])
		if !ok {
			return nil, tengo.ErrInvalidArgumentType{
				Name:     "second",
				Expected: "string(compatible)",
				Found:    args[1].TypeName(),
			}
		}

		if fn(ss1, s2) {
			return tengo.TrueValue, nil
		}
		return tengo.FalseValue, nil
	}
}

// FuncASsRSs transform a function of 'func([]string) []string' signature into
// CallableFunc type.
func FuncASsRSs(fn func([]string) []string) tengo.CallableFunc {
	return func(args ...tengo.Object) (ret tengo.Object, err error) {
		if len(args) != 1 {
			return nil, tengo.ErrWrongNumArguments
		}
		var ss1 []string
		switch arg0 := args[0].(type) {
		case *tengo.Array:
			for idx, a := range arg0.Value {
				as, ok := tengo.ToString(a)
				if !ok {
					return nil, tengo.ErrInvalidArgumentType{
						Name:     fmt.Sprintf("first[%d]", idx),
						Expected: "string(compatible)",
						Found:    a.TypeName(),
					}
				}
				ss1 = append(ss1, as)
			}
		case *tengo.ImmutableArray:
			for idx, a := range arg0.Value {
				as, ok := tengo.ToString(a)
				if !ok {
					return nil, tengo.ErrInvalidArgumentType{
						Name:     fmt.Sprintf("first[%d]", idx),
						Expected: "string(compatible)",
						Found:    a.TypeName(),
					}
				}
				ss1 = append(ss1, as)
			}
		default:
			return nil, tengo.ErrInvalidArgumentType{
				Name:     "first",
				Expected: "array",
				Found:    args[0].TypeName(),
			}
		}
		arr := &tengo.Array{}
		for _, elem := range fn(ss1) {
			if len(elem) > tengo.MaxStringLen {
				return nil, tengo.ErrStringLimit
			}
			arr.Value = append(arr.Value, &tengo.String{Value: elem})
		}
		return arr, nil
	}
}