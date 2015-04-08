package main

/*
 * rc map[string]string
 *
 */

import (
  "os"
  "bytes"
  "text/template"
  "github.com/gutenye/fil"
  "github.com/BurntSushi/toml"
  "github.com/fatih/color"
  "github.com/gutenye/gutgen/shell"
)

func New(templateName, projectPath string) {
  template := os.Getenv("HOME")+"/.gutgen/"+templateName
  if ok, _ := fil.IsNotExist(template); ok {
    shell.ErrorExit("template does not exists -- "+template)
  }

  rc := loadRc(os.Getenv("HOME") + "/.gutgenrc")
  rc["project"] = fil.Base(projectPath)

  shell.Say("      %s %s\n", color.CyanString("create"), projectPath)
  if err := fil.CpDirOnly(template, projectPath); err != nil {
    shell.ErrorExit(err)
  }

  err := fil.Walk(template, func(src string, fi os.FileInfo, e error) error {
		if e != nil {
      shell.ErrorExit(e)
    }
    // skip template itself
    if src == template {
      return nil
    }

    relSrc, _ := fil.Rel(template, src)
    relDest, err := executeTemplate(relSrc, rc)
    if err != nil {
      shell.ErrorExit(err)
    }
    dest := projectPath+"/"+relDest

    shell.Say("      %s %s\n", color.CyanString("create"), dest)
    switch t := fil.TypeFileInfo(fi); t {
    case "dir":
      if err := fil.CpDirOnly(src, dest); err != nil{
        shell.ErrorExit(err)
      }
    case "symlink":
      if err := fil.Cp(src, dest); err != nil{
        shell.ErrorExit(err)
      }
    case "regular":
      if err := cpFile(src, dest, rc); err!= nil {
        shell.ErrorExit(err)
      }
    default:
      shell.ErrorExit("Unknown file type -- "+t)
    }
    return nil
  })
  if err != nil {
    shell.ErrorExit(err)
  }
}

func loadRc(file string) map[string]string {
  if ok, _ := fil.IsNotExist(file); ok {
    return map[string]string{}
  }

  var rc map[string]string
  if _, err := toml.DecodeFile(file, &rc); err != nil {
    shell.ErrorExit(err)
  }
  return rc
}

func executeTemplate(path string, data interface{}) (string, error) {
  var buf bytes.Buffer
  tmpl, err := template.New("default").Parse(path)
  if err != nil {
    return "", err
  }
  tmpl.Execute(&buf, data)
  return buf.String(), nil
}

func cpFile(src, dest string, data interface{}) error {
  tmpl, err := template.ParseFiles(src)
  if err != nil {
    return err
  }
  file, err := os.OpenFile(dest, os.O_WRONLY | os.O_CREATE | os.O_TRUNC, 0644)
  if err != nil {
    return err
  }
  tmpl.Execute(file, data)
  return nil
}

/*
func List() []string {
  return
}
*/
