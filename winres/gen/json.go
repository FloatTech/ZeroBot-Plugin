// Package main generates winres.json
package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/FloatTech/ZeroBot-Plugin/kanban/banner"
)

const js = `{
  "RT_GROUP_ICON": {
    "APP": {
      "0000": [
        "icon.png",
        "icon16.png"
      ]
    }
  },
  "RT_MANIFEST": {
    "#1": {
      "0409": {
        "identity": {
          "name": "ZeroBot-Plugin",
          "version": "%s"
        },
        "description": "",
        "minimum-os": "vista",
        "execution-level": "as invoker",
        "ui-access": false,
        "auto-elevate": false,
        "dpi-awareness": "system",
        "disable-theming": false,
        "disable-window-filtering": false,
        "high-resolution-scrolling-aware": false,
        "ultra-high-resolution-scrolling-aware": false,
        "long-path-aware": false,
        "printer-driver-isolation": false,
        "gdi-scaling": false,
        "segment-heap": false,
        "use-common-controls-v6": false
      }
    }
  },
  "RT_VERSION": {
    "#1": {
      "0000": {
        "fixed": {
          "file_version": "%s",
          "product_version": "%s",
          "timestamp": "%s"
        },
        "info": {
          "0409": {
            "Comments": "OneBot plugins based on ZeroBot",
            "CompanyName": "FloatTech",
            "FileDescription": "https://github.com/FloatTech/ZeroBot-Plugin",
            "FileVersion": "%s",
            "InternalName": "",
            "LegalCopyright": "%s",
            "LegalTrademarks": "",
            "OriginalFilename": "ZBP.EXE",
            "PrivateBuild": "",
            "ProductName": "ZeroBot-Plugin",
            "ProductVersion": "%s",
            "SpecialBuild": ""
          }
        }
      }
    }
  }
}`

const timeformat = `2006-01-02T15:04:05+08:00`

func main() {
	f, err := os.Create("winres.json")
	if err != nil {
		panic(err)
	}
	defer f.Close()
	i := strings.LastIndex(banner.Version, "-")
	if i <= 0 {
		i = len(banner.Version)
	}
	commitcnt := strings.Builder{}
	commitcnt.WriteString(banner.Version[1:i])
	commitcnt.WriteByte('.')
	commitcntcmd := exec.Command("git", "rev-list", "--count", "HEAD")
	commitcntcmd.Stdout = &commitcnt
	err = commitcntcmd.Run()
	if err != nil {
		panic(err)
	}
	fv := commitcnt.String()[:commitcnt.Len()-1]
	_, err = fmt.Fprintf(f, js, fv, fv, banner.Version, time.Now().Format(timeformat), fv, banner.Copyright+". All Rights Reserved.", banner.Version)
	if err != nil {
		panic(err)
	}
}
