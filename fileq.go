package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

type Q struct {
	fileID   string
	fileName string
}

var q []Q

func main() {
	port := flag.String("port", "80", "bind port (default: 80)")

	handler := http.HandlerFunc(requestHandler)

	err := http.ListenAndServe(":"+*port, handler)
	log.Fatal(err)
}

func requestHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		// Queue POP request

		last := len(q) - 1
		if last < 0 {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		http.ServeFile(w, r, q[last].fileID+"/"+q[last].fileName)

		q = q[:last]

	} else if r.Method == http.MethodPost {
		// Queue PUT request

		file, fileHeader, err := r.FormFile("file")
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		fileId := r.FormValue("file_id")

		fileBuffer, err := ioutil.ReadAll(file)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		os.Mkdir(fileId, os.ModePerm)                                            // Create dir
		err = ioutil.WriteFile(fileId+"/"+fileHeader.Filename, fileBuffer, 0644) // Write file
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Print(err)
			return
		}

		q = append(q, Q{fileID: fileId, fileName: fileHeader.Filename}) // PUT to Queue

		w.WriteHeader(http.StatusOK)
	} else {
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}
