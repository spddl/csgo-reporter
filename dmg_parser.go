package main

import (
	"bytes"
	"html/template"
	"strconv"
	"strings"
)

type dmgParser struct {
	Enemys []player
	Enable bool
}

type player struct {
	Name string
	HP   int
	Hits string
}

func (dmg *dmgParser) Add(line string) {
	seperatorIndex := strings.Index(line, " - ")
	inIndex := strings.LastIndex(line, " in ")
	hp, _ := strconv.Atoi(line[seperatorIndex+3 : inIndex])
	if hp < 100 { // Nur wenn der Gegner noch nicht Tot ist
		firstMark := strings.Index(line, "\"")
		secondMark := strings.LastIndex(line, "\"")
		playerName := line[firstMark+1 : secondMark]

		hitsIndex := strings.Index(line, " hit")

		dmg.Enemys = append(dmg.Enemys, player{
			Name: playerName,
			HP:   hp,
			Hits: line[inIndex+4 : hitsIndex],
		})
	}
}

const reporttemplate = `{{ range $index, $element := .Enemys}}{{if $index}} - {{end}}{{.Name}} {{.HP}} in {{.Hits}} {{if eq .Hits "1"}}hit{{else}}hits{{end}}{{end}}`

func (dmg *dmgParser) Output(config *Config) {
	var templ string
	if config.ReportTemplate != "" {
		templ = config.ReportTemplate
	} else {
		templ = reporttemplate
	}
	tmpl := template.Must(template.New("").Parse(templ))

	var tpl bytes.Buffer
	if err := tmpl.Execute(&tpl, dmg); err != nil {
		panic(err)
	}

	WriteFile(config, tpl.Bytes())

	dmg.Enemys = []player{}
}
