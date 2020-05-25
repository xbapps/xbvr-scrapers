package helpers

import (
	"net/url"

	"github.com/d5/tengo/v2"
)

var urlModule = map[string]tengo.Object{
	"get_query_param": &tengo.UserFunction{
		Name:  "get_query_param",
		Value: urlGetQueryParam,
	}, // get_query_param(url, param) => string/error
}

func urlGetQueryParam(args ...tengo.Object) (ret tengo.Object, err error) {
	if len(args) != 2 {
		err = tengo.ErrWrongNumArguments
		return
	}

	s1, ok := tengo.ToString(args[0])
	if !ok {
		err = tengo.ErrInvalidArgumentType{
			Name:     "first",
			Expected: "string(compatible)",
			Found:    args[0].TypeName(),
		}
		return
	}

	u, err := url.Parse(s1)
	if err != nil {
		ret = wrapError(err)
		return
	}

	s2, ok := tengo.ToString(args[1])
	if !ok {
		err = tengo.ErrInvalidArgumentType{
			Name:     "second",
			Expected: "string(compatible)",
			Found:    args[1].TypeName(),
		}
		return
	}

	q := u.Query()
	ret = &tengo.String{Value: q.Get(s2)}

	return
}
