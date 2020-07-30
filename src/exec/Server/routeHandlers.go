package main

import (
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"tech/app/comms"
	"tech/app/logger"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
)

const (
	// For testing purposes only. Will be changed to firmware update directory later
	maxUploadSize     = (500 * 1048576) // 500 MB
	webPagesServePath = "./"
	uploadPath        = "/home/root/"
	uploadFileName    = "updatefile"
)

func configureRoutes(router *chi.Mux, logHTTP bool) {

	if logHTTP {
		router.Use(middleware.Logger)
	}
	router.Route("/command", func(r chi.Router) {
		r.Post("/", handleCommand)
	})
	router.Route("/upload", func(r chi.Router) {
		r.Post("/", uploadFileHandler)
	})
	router.Get("/*", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.FileServer(http.Dir(webPagesServePath)).ServeHTTP(w, r)
	}))
}

func defaultHandler(w http.ResponseWriter, r *http.Request) {

	http.ServeFile(w, r, "index.html")
	logger.Log("Served unknown route")
}

func handleCommand(w http.ResponseWriter, r *http.Request) {

	if env.client == nil {
		logger.Log("Command client not available")
		http.Error(w, http.StatusText(500), 500)
		return
	}

	target := r.Header.Get("Target")
	action := r.Header.Get("Action")

	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		logger.Log("Failed to read request body, %v", err)
		http.Error(w, http.StatusText(500), 500)
		return
	}

	resp, err := env.client.Send(comms.BuildPacket(target, action, data), 100)
	if err == nil {
		w.WriteHeader(http.StatusOK)
		w.Write(resp.Data)
	} else {
		logger.Log("Failed to execute command, %v", err)
		http.Error(w, http.StatusText(500), 500)
	}
}

func uploadFileHandler(w http.ResponseWriter, r *http.Request) {

	logger.LogDebug("File received, please wait...")
	if env.client == nil {
		logger.Log("Command client not available")
		http.Error(w, http.StatusText(500), 500)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, maxUploadSize)
	if err := r.ParseMultipartForm(maxUploadSize); err != nil {
		logger.Log("File upload failed, %v", err)
		http.Error(w, http.StatusText(400), 400)
		return
	}

	file, _, err := r.FormFile("fileKey")
	if err != nil {
		logger.Log("File upload failed, %v", err)
		http.Error(w, http.StatusText(400), 400)
		return
	}
	defer file.Close()

	fileBytes, err := ioutil.ReadAll(file)
	if err != nil {
		logger.Log("File upload failed, %v", err)
		http.Error(w, http.StatusText(400), 400)
		return
	}

	logger.LogDebug("Decoding File Type")
	fileType := http.DetectContentType(fileBytes)
	var fileEndings string
	switch fileType {
	case "application/octet-stream":
		fileEndings = ".jpeg"
	default:
		logger.Log("Invalid file type, %v", err)
		http.Error(w, http.StatusText(400), 400)
		return
	}

	logger.LogDebug("File Type Valid")
	fileName := uploadFileName
	filePath := filepath.Join(uploadPath, fileName+fileEndings)

	newFile, err := os.Create(filePath)
	if err != nil {
		logger.Log("Cannot write file, %v", err)
		http.Error(w, http.StatusText(500), 500)
		return
	}
	defer newFile.Close()

	if _, err := newFile.Write(fileBytes); err != nil {
		logger.Log("Cannot write file, %v", err)
		http.Error(w, http.StatusText(500), 500)
		return
	}
	w.Write([]byte("SUCCESS"))
	logger.LogDebug("File uploaded as %s", filePath)

	return
}
