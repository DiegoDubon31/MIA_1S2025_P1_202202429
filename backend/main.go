package main

import (
	"MIA_Proyecto1/backend/Analyzer"
	"encoding/json"
	"log"
	"net/http"
	"time"
)

type ScriptRequest struct {
	Script string `json:"script"`
}

func main() {
	//Analyzer.Analyze()
	mux := http.NewServeMux()
	fs := http.FileServer(http.Dir("frontend"))
	mux.Handle("/", fs)
	mux.HandleFunc("/execute", executeHandler)
	server := &http.Server{
		Addr:           ":3000",
		Handler:        mux,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	log.Println("Servidor corriendo en http://localhost:3000")
	log.Fatal(server.ListenAndServe())

}

func executeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		return
	}

	var req ScriptRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Error al decodificar el cuerpo de la solicitud", http.StatusBadRequest)
		return
	}

	output := Analyzer.AnalyzeScript(req.Script)
	w.Write([]byte(output))
}
