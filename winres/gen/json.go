// Package main generates winres.json
package main

import (
	"fmt"
	"os"
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
          "name": "",
          "version": ""
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
            "FileDescription": "Project: https://github.com/FloatTech/ZeroBot-Plugin",
            "FileVersion": "%s",
            "InternalName": "",
            "LegalCopyright": "%s",
            "LegalTrademarks": "",
            "OriginalFilename": "ZBP.EXE",
            "PrivateBuild": "",
            "ProductName": "",
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
	_, err = fmt.Fprintf(f, js, banner.Version, banner.Version, time.Now().Format(timeformat), banner.Version, banner.Copyright, banner.Version)
	if err != nil {
		panic(err)
	}
}
