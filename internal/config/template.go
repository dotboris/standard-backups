package config

import (
	"bytes"
	"fmt"
	"text/template"
)

type configTemplate struct {
	Secrets map[string]string
}

func (t *configTemplate) Apply(path string, value any) (any, error) {
	switch value := value.(type) {
	case string:
		tpl := template.New(path)
		tpl, err := tpl.Parse(value)
		if err != nil {
			return nil, err
		}
		out := bytes.Buffer{}
		err = tpl.Execute(&out, t)
		if err != nil {
			return nil, err
		}
		return out.String(), nil
	case map[string]any:
		res := map[string]any{}
		for k, v := range value {
			templated, err := t.Apply(fmt.Sprintf("%s.%s", path, k), v)
			if err != nil {
				return nil, err
			}
			res[k] = templated
		}
		return res, nil
	case []any:
		res := make([]any, len(value))
		for i, v := range value {
			templated, err := t.Apply(fmt.Sprintf("%s.%d", path, i), v)
			if err != nil {
				return nil, err
			}
			res[i] = templated
		}
		return res, nil
	default:
		return value, nil
	}
}
