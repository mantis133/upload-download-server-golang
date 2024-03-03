package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
)

func hewwo(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/hello" {
		http.Error(w, "404 not found", http.StatusNotFound)
	}

	if r.Method != "GET" {
		http.Error(w, "Method is not supported.", http.StatusNotFound)
		return
	}

	fmt.Fprintf(w, "Hello!")

}

func formHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method is not supported.", http.StatusNotFound)
		return
	}

	if err := r.ParseForm(); err != nil {
		fmt.Fprintf(w, "ParseForm() err: %v", err)
		return
	}
	fmt.Fprintf(w, "POST request successful\n")
	name := r.FormValue("name")
	address := r.FormValue("address")

	fmt.Printf("\nName = %v\n", name)
	fmt.Printf("address = %v\n", address)
}

func index(w http.ResponseWriter, r *http.Request) {

}

func uploadFile(w http.ResponseWriter, r *http.Request) {
	file, handler, err := r.FormFile("file")
	if err != nil {
		fmt.Fprintf(w, "Error with the file %v", err, http.StatusForbidden)
		return
	}
	defer file.Close()

	f, err := os.OpenFile("tmp/"+handler.Filename, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		fmt.Println("Error creating file")
		return
	}
	defer f.Close()

	io.Copy(f, file)
	fmt.Fprintf(w, "File %s uploaded successfully", handler.Filename)
}

func downloadFile(w http.ResponseWriter, r *http.Request) {
	fileID := r.FormValue("fileID")
	if fileID != "001" {
		http.Error(w, "file not found", http.StatusNotFound)
		return
	}
	filePath := "static/index.html" // can be any file 
	file, err := os.OpenFile(filePath, os.O_RDONLY, 0666)
	if err != nil {
		http.Error(w, "Internal server error collecting the file", http.StatusInternalServerError)
		return
	}
	b := make([]byte, 1)
	file.Read(b)
	file.Close()
	w.Header().Set("Content-Disposition", "attachment; filename="+fileID)
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Length", strconv.Itoa(len(b)))
	//w.Write(b)

	http.ServeFile(w, r, filePath)

}

func testAuth(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("Bearer")
	println(token)
}

func GetOutboundIP() net.IP {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)

	return localAddr.IP
}

func HandleSizeCap(maxSize int64, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.ContentLength > maxSize {
			http.Error(w, "Request body is too large", http.StatusRequestEntityTooLarge)
			return
		}
		next(w, r)
	}
}

func main() {
	port := 8080

	fileServer := http.FileServer(http.Dir("./static"))

	http.HandleFunc("/hello", hewwo)
	http.HandleFunc("/form", formHandler)
	http.HandleFunc("/upload", HandleSizeCap(int64(10*1024*1024), setMethod("POST", uploadFile)))
	http.HandleFunc(
		"/download",
		setMethod(
			"POST",
			downloadFile,
		),
	)
	http.HandleFunc("/auth", setMethod("POST", testAuth))
	http.Handle("/", fileServer)

	fmt.Printf("Starting server at: %v:%v\n", GetOutboundIP(), port)
	//fmt.Printf("Open through port: %v\n", port)
	if err := http.ListenAndServe(fmt.Sprintf(":%v", port), nil); err != nil {
		log.Fatal(err)
	}
}
