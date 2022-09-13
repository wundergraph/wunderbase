package graphiql

import (
	_ "embed"
	"strings"
)

//go:embed "graphiql.html"
var s string

func GetGraphiqlPlaygroundHTML(listenAddr string) string {
	return strings.Replace(s, "{{apiURL}}", listenAddr, -1)
}
