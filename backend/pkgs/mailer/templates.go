package mailer

import (
	"bytes"
	_ "embed"
	"html/template"
)

//go:embed templates/welcome.html
var templatesWelcome string

type TemplateDefaults struct {
	CompanyName        string
	CompanyAddress     string
	CompanyURL         string
	ActivateAccountURL string
	UnsubscribeURL     string
}

type TemplateProps struct {
	Defaults TemplateDefaults
	Data     map[string]string
}

func (tp *TemplateProps) Set(key, value string) {
	tp.Data[key] = value
}

func DefaultTemplateData() TemplateProps {
	return TemplateProps{
		Defaults: TemplateDefaults{
			CompanyName:        "sysadminsmedia.com",
			CompanyAddress:     "123 Main St, Anytown, CA 12345",
			CompanyURL:         "https://sysadminsmedia.com",
			ActivateAccountURL: "https://google.com",
			UnsubscribeURL:     "https://google.com",
		},
		Data: make(map[string]string),
	}
}

func render(tpl string, data TemplateProps) (string, error) {
	tmpl, err := template.New("name").Parse(tpl)
	if err != nil {
		return "", err
	}

	var tplBuffer bytes.Buffer

	err = tmpl.Execute(&tplBuffer, data)

	if err != nil {
		return "", err
	}

	return tplBuffer.String(), nil
}

func RenderWelcome() (string, error) {
	return render(templatesWelcome, DefaultTemplateData())
}
