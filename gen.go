package main

import (
  "os"
  "time"
  "fmt"
  "strings"
  "github.com/gutengo/tagen/strings2"
  "github.com/gutengo/fil"
  "github.com/gutengo/shell"
  "github.com/gutengo/mustache"
  "github.com/BurntSushi/toml"
  "github.com/fatih/color"
)

var rc map[string]interface{}
var mustacheHelpers map[string]interface{}

func New(templateName, projectPath string) {
  template := initialize(templateName, projectPath)

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
    relDest := mustache.Render(relSrc, mustacheHelpers, rc)
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

func Add(templateName, name string) {
  template := initialize(templateName, name)
  if name == "" { name = templateName }
  shell.Say("      %s %s\n", color.CyanString("create"), name)
  if err := cpFile(template, name, rc); err!= nil {
    shell.ErrorExit(err)
  }
}

/*
func List() []string {
  return
}
*/

func loadRc(file string) (ret map[string]interface{}) {
  if ok, _ := fil.IsNotExist(file); ok {
    return map[string]interface{}{}
  }

  if _, err := toml.DecodeFile(file, &ret); err != nil {
    shell.ErrorExit("%s: %s\n", "Load "+file, err)
  }
  return ret
}

func cpFile(src, dest string, data interface{}) error {
  fi, err := fil.Lstat(src)
  if err != nil {
    return err
  }
  tmpl, err := mustache.ParseFile(src)
  if err != nil {
    return fmt.Errorf("%s: %s", src, err.Error())
  }
  ret := tmpl.Render(mustacheHelpers, data)
  if err := fil.WriteFile(dest, []byte(ret), fi.Mode().Perm()); err != nil {
    return err
  }
  return nil
}

func initialize(templateName, projectPath string) string {
  mustache.Otag, mustache.Ctag = "%%", "%%"

  mustacheHelpers = map[string]interface{}{
    "var": func(text string, render mustache.RenderFunc) string {
      lines := strings.Split(text, "\n")
      for _, line := range lines {
        // skip empty lines
        if line == "" {
          continue
        }
        if strings.Contains(line, "||=") {
          parts := strings.Split(line, "||=")
          name, value := strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1])
          if _, ok := mustacheHelpers[name]; !ok {
            mustacheHelpers[name] = render(value)
          }
        } else {
          parts := strings.Split(line, "=")
          name, value := strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1])
          mustacheHelpers[name] = render(value)
        }
      }
      return ""
    },

    "year": func() string {
      return fmt.Sprintf("%v", time.Now().Year())
    },
  }

  appDir := os.Getenv("HOME")+"/.gutgen"
  template := appDir+"/"+templateName
  if ok, _ := fil.IsNotExist(template); ok {
    shell.ErrorExit("template does not exists -- "+template)
  }

  rc = loadRc(os.Getenv("HOME") + "/.gutgenrc")
  project := fil.Base(projectPath)
  rc["project"] = project
  rc["Project"] = strings2.ClassCase(project)

  return template
}
