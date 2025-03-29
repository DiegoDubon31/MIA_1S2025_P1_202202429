package main

import (
	"MIA_Proyecto1/backend/Analyzer"
	"encoding/json"
	"fmt"
	"net/http"
)

type ScriptRequest struct {
	Script string `json:"script"`
}

func main() {
	http.HandleFunc("/execute", executeHandler)

	fmt.Println("Servidor corriendo en http://localhost:3000")
	err := http.ListenAndServe(":3000", nil)
	if err != nil {
		fmt.Println("Error al iniciar el servidor:", err)
	}
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
