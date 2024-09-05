package main

import (
	"github.com/gptscript-ai/cmd"
	"github.com/gptscript-ai/otto/pkg/cli"
)

func main() {
	cmd.Main(cli.New())
}
