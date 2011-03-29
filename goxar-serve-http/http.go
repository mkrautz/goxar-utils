// Copyright (c) 2011 Mikkel Krautz <mikkel@krautz.dk>
// The use of this source code is goverened by a BSD-style
// license that can be found in the LICENSE-file.

package main

import (
	"flag"
	"fmt"
	"github.com/mkrautz/goxar"
	"http"
	"io"
	"log"
	"strings"
)

var file *string = flag.String("f", "", "The XAR file to serve")
var addr *string = flag.String("http", ":8580", "Address to serve HTTP at")
var xarFile *xar.Reader

func main() {
	flag.Parse()

	if len(*file) == 0 {
		flag.Usage()
		return
	}

	xf, err := xar.OpenReader(*file)
	if err != nil {
		log.Fatalf(err.String())
	}

	xarFile = xf
	http.HandleFunc("/", xarServe)
	http.ListenAndServe(*addr, nil)
}

func xarServe(rw http.ResponseWriter, r *http.Request) {
	log.Printf("%s %s", r.Method, r.RawURL)

	split := strings.Split(r.RawURL, "/", 2)
	fn := split[1]

	for _, file := range xarFile.File {
		if file.Name == fn || r.RawURL == "/" {
			if file.Type == xar.FileTypeDirectory || r.RawURL == "/" {
				body := fmt.Sprintf("<h1>Index of %s</h1>", r.RawURL)
				body += "<ul>"
				body += "<li><a href=\"..\">..</a></li>\n"
				for _, subFile := range xarFile.File {
					typeStr := ""
					switch subFile.Type {
					case xar.FileTypeFile:
						typeStr = "[file]"
					case xar.FileTypeDirectory:
						typeStr = "[dir]"
					default:
						typeStr = "[unknown]"
					}
					if r.RawURL == "/" || strings.Contains(subFile.Name, file.Name) {
						if r.RawURL != "/" {
							dirSplit := strings.Split(subFile.Name, file.Name+"/", -1)
							if len(dirSplit) == 1 {
								continue
							}
							if len(dirSplit[0]) != 0 {
								continue
							}
							if !strings.Contains(dirSplit[1], "/") {
								body += fmt.Sprintf("<li>%s <a href=\"%s\">%s</a></li>\n", typeStr, "/"+subFile.Name, dirSplit[1])
							}
						} else {
							if !strings.Contains(subFile.Name, "/") {
								body += fmt.Sprintf("<li>%s <a href=\"%s\">%s</a></li>", typeStr, "/"+subFile.Name, subFile.Name)
							}
						}
					}
				}
				body += "</ul>"
				body += "<hr/><i>goxar</i>"

				rw.Header().Set("Content-Type", "text/html")
				rw.Write([]byte(body))
				return
			} else if file.Type == xar.FileTypeFile {
				fr, err := file.Open()
				if err != nil {
					http.Error(rw, err.String(), 500)
					return
				}

				rw.Header().Set("Content-Type", "application/octet-stream")
				io.Copy(rw, fr)
				return
			}
			break
		}
	}
}
