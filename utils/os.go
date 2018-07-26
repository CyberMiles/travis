package utils

import (
	"bytes"
	"log"
	"os/exec"
	"strings"
)

var GOOSDIST string
var GOOSVERS string

var SUPPORT_OS = map[string][]string {
	"ubuntu": {"16.04"},
	"centos": {"^7.0"},
}

func init() {
	cmd := exec.Command("cat", "/etc/os-release")
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		log.Fatal("Can't get the os info, for now we only support: \n", SUPPORT_OS)
	}
	rs := strings.Split(out.String(), "\n")
	for _, r := range rs {
		m := strings.Split(r, "=")
		if len(m) == 2 {
			switch m[0] {
			case "ID":
				GOOSDIST = strings.Replace(m[1], "\"", "", -1)
			case "VERSION_ID":
				GOOSVERS = strings.Replace(m[1], "\"", "", -1)
			}
		}
	}

	if GOOSDIST == "" || GOOSVERS == "" {
		log.Fatal("Can't get the os info, for now we only support: \n", SUPPORT_OS)
	}

	if versions := SUPPORT_OS[GOOSDIST]; versions != nil {
		for _, vs := range versions {
			if strings.Index(vs, "^") == 0 {
				vMj := vs[1:strings.Index(vs, ".")]
				idx := strings.Index(GOOSVERS, ".")
				dvMj := GOOSVERS
				if idx != -1 {
					dvMj = GOOSVERS[0:idx]
				}
				if vMj == dvMj {
					return
				}
			} else if vs == GOOSVERS {
				return
			}
		}
	}
	log.Fatal("Your os is not supported, for now we only support: \n", SUPPORT_OS)
}
