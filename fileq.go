package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/urfave/negroni"
)

type Q struct {
	fileID   string
	fileName string
}

var q []Q

// FILE_DIR_NAME string
// root dir for files
const FILE_DIR_NAME = "file"

func main() {
	port := flag.String("port", "80", "bind port (default: 80)")

	handler := http.HandlerFunc(requestHandler)

	n := negroni.New(negroni.NewLogger(), negroni.NewRecovery())
	n.UseHandler(handler)

	err := http.ListenAndServe(":"+*port, n)
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

		http.ServeFile(w, r, makeSavePath(q[last].fileID, q[last].fileName))

		os.RemoveAll("file/" + q[last].fileID)

		q = q[:last]

	} else if r.Method == http.MethodPost {
		// Queue PUT request

		file, fileHeader, err := r.FormFile("file")
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		fileBuffer, err := ioutil.ReadAll(file)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		fileId := strconv.Itoa(len(q))

		path := makeSavePath(fileId, fileHeader.Filename)
		if _, err := os.Stat(path); !os.IsNotExist(err) {
			w.WriteHeader(http.StatusConflict)
			return
		}

		os.MkdirAll("file/"+fileId, os.ModePerm)       // Create dir
		err = ioutil.WriteFile(path, fileBuffer, 0644) // Write file
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

// makeSavePath func
// private function
// returns save dir
func makeSavePath(fileId string, fileName string) string {
	return FILE_DIR_NAME + "/" + fileId + "/" + fileName
}
