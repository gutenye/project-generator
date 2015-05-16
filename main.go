package main

import (
  "fmt"
  "os"
  "github.com/codegangsta/cli"
  "github.com/gutengo/shell"
)

func main() {
  cli.AppHelpTemplate = `{{.Name}} v{{.Version}} - {{.Usage}}

USAGE:
   {{.Name}} {{if .Flags}}[options] {{end}}<template ..>

COMMANDS:
   {{range .Commands}}{{.Name}}{{with .ShortName}}, {{.}}{{end}}{{ "\t" }}{{.Usage}}
   {{end}}{{if .Flags}}
GLOBAL OPTIONS:
   {{range .Flags}}{{.}}
   {{end}}{{end}}
`

  app := cli.NewApp()
  app.Name = "gutgenerator"
  app.Usage = "a generic project generator"
  app.Version = "0.0.1"

  app.Action = func(c *cli.Context) {
    cli.ShowAppHelp(c)
  }

  app.Commands = []cli.Command{
    {
      Name:      "new",
      Usage:     "create a new project",
      Action: func(c *cli.Context) {
        if len(c.Args()) < 2 {
          shell.ErrorExit("arguments length < 2\nUSAGE: gutgenerator new <template> <project>")
        }
        New(c.Args()[0], c.Args()[1])
      },
    }, {
      Name:      "add",
      Usage:     "add more templates in current directory",
      Action: func(c *cli.Context) {
        args := c.Args()
        if len(args) < 1 {
          shell.ErrorExit("arguments length < 1\nUSAGE: gutgenerator add <template> [name]")
        }
        Add(args.Get(0), args.Get(1))
      },
    },
    {
      Name:      "list",
      Usage:     "list all templates",
      Action: func(c *cli.Context) {
        fmt.Println("list", c.Args())
      },
    },
  }

  app.Run(os.Args)
}
