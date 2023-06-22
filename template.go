package otium

import (
	"io"
	"text/template"
)

func renderTemplate(wr io.Writer, text string, bag Bag) error {
	tmpl, err := template.New("description").Parse(text)
	if err != nil {
		return err
	}
	err = tmpl.Execute(wr, bag.bag)
	if err != nil {
		return err
	}

	return nil
}
