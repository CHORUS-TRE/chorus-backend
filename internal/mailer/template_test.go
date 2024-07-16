package mailer

import (
	"github.com/stretchr/testify/assert"
	"html/template"
	"testing"
)

func TestTemplates(t *testing.T) {
	for name, mailTemplate := range mailTemplates {
		var tmpl, err = template.New(name.String()).Parse(mailTemplate)
		assert.Nil(t, err)
		assert.NotNil(t, tmpl)
	}
}
