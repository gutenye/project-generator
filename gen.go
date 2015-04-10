package main

/*
 * rc map[string]string
 *
 */

import (
  "os"
  "time"
  "fmt"
  "strings"
  "github.com/gutenye/mustache"
  "github.com/gutenye/fil"
  "github.com/BurntSushi/toml"
  "github.com/fatih/color"
  "github.com/gutenye/gutgen/shell"
)

var mustacheHelpers map[string]interface{}

func initialize() {
  mustacheHelpers = map[string]interface{}{
    "var": func(text string, render mustache.RenderFunc) string {
      lines := strings.Split(text, "\n")
      for _, line := range lines {
        // skip empty lines
        if line == "" {
          continue
        }
        parts := strings.Split(line, " = ")
        name, value := strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1])
        mustacheHelpers[name] = render(value)
      }
      return ""
    },

    "year": func() string {
      return fmt.Sprintf("%v", time.Now().Year())
    },
  }
}

func New(templateName, projectPath string) {
  initialize()
  appDir := os.Getenv("HOME")+"/.gutgen"
  template := appDir+"/"+templateName
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

/*
func List() []string {
  return
}
*/
