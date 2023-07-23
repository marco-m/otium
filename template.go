package otium

import (
	"io"
	"text/template"
)

func renderTemplate(wr io.Writer, text string, bag map[string]Variable) error {
	tmpl, err := template.New("description").Parse(text)
	if err != nil {
		return err
	}

	m := make(map[string]string, len(bag))
	for k, v := range bag {
		m[k] = v.val
	}

	err = tmpl.Execute(wr, m)
	if err != nil {
		return err
	}

	return nil
}
