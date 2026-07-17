package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	texttemplate "text/template"
)

type APIResumeRequest struct {
	Resume Resume `json:"resume"`
}

type APIResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

// corsMiddleware adds CORS headers to responses
func corsMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next(w, r)
	}
}

// generateResumeLatexHandler generates LaTeX from JSON resume data
func generateResumeLatexHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("[API] POST /api/resume/latex")
	if r.Method != http.MethodPost {
		respondJSON(w, http.StatusMethodNotAllowed, APIResponse{
			Success: false,
			Message: "Method not allowed. Use POST.",
		})
		return
	}

	var req APIResumeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		fmt.Printf("[API] Error decoding request: %v\n", err)
		respondJSON(w, http.StatusBadRequest, APIResponse{
			Success: false,
			Message: fmt.Sprintf("Invalid request body: %v", err),
		})
		return
	}

	latex, err := generateLatex(req.Resume)
	if err != nil {
		fmt.Printf("[API] Error generating LaTeX: %v\n", err)
		respondJSON(w, http.StatusInternalServerError, APIResponse{
			Success: false,
			Message: fmt.Sprintf("Failed to generate LaTeX: %v", err),
		})
		return
	}

	fmt.Println("[API] LaTeX generated successfully")
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Header().Set("Content-Disposition", "attachment; filename=\"resume.tex\"")
	w.Write([]byte(latex))
}

// generateResumePDFHandler generates PDF from JSON resume data
func generateResumePDFHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("[API] POST /api/resume/pdf")
	if r.Method != http.MethodPost {
		respondJSON(w, http.StatusMethodNotAllowed, APIResponse{
			Success: false,
			Message: "Method not allowed. Use POST.",
		})
		return
	}

	var req APIResumeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		fmt.Printf("[API] Error decoding request: %v\n", err)
		respondJSON(w, http.StatusBadRequest, APIResponse{
			Success: false,
			Message: fmt.Sprintf("Invalid request body: %v", err),
		})
		return
	}

	// Generate LaTeX
	fmt.Println("[API] Generating LaTeX...")
	latex, err := generateLatex(req.Resume)
	if err != nil {
		fmt.Printf("[API] Error generating LaTeX: %v\n", err)
		respondJSON(w, http.StatusInternalServerError, APIResponse{
			Success: false,
			Message: fmt.Sprintf("Failed to generate LaTeX: %v", err),
		})
		return
	}

	// Compile to PDF (using temp directory)
	fmt.Println("[API] Creating temp directory...")
	tempDir, err := os.MkdirTemp("", "resume-*")
	if err != nil {
		fmt.Printf("[API] Error creating temp directory: %v\n", err)
		respondJSON(w, http.StatusInternalServerError, APIResponse{
			Success: false,
			Message: fmt.Sprintf("Failed to create temp directory: %v", err),
		})
		return
	}
	defer os.RemoveAll(tempDir)

	// Write LaTeX to temp file
	latexPath := filepath.Join(tempDir, "resume.tex")
	if err := os.WriteFile(latexPath, []byte(latex), 0644); err != nil {
		fmt.Printf("[API] Error writing LaTeX file: %v\n", err)
		respondJSON(w, http.StatusInternalServerError, APIResponse{
			Success: false,
			Message: fmt.Sprintf("Failed to write LaTeX file: %v", err),
		})
		return
	}

	// Compile to PDF
	fmt.Println("[API] Compiling PDF...")
	pdfPath := filepath.Join(tempDir, "resume.pdf")
	if err := compilePDF(latexPath, tempDir); err != nil {
		fmt.Printf("[API] Error compiling PDF: %v\n", err)
		respondJSON(w, http.StatusInternalServerError, APIResponse{
			Success: false,
			Message: fmt.Sprintf("Failed to compile PDF: %v", err),
		})
		return
	}

	// Read PDF file
	fmt.Println("[API] Reading PDF file...")
	pdfData, err := os.ReadFile(pdfPath)
	if err != nil {
		fmt.Printf("[API] Error reading PDF: %v\n", err)
		respondJSON(w, http.StatusInternalServerError, APIResponse{
			Success: false,
			Message: fmt.Sprintf("Failed to read PDF: %v", err),
		})
		return
	}

	// Send PDF
	fmt.Printf("[API] PDF generated successfully (%d bytes)\n", len(pdfData))
	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", "attachment; filename=\"resume.pdf\"")
	w.Header().Set("Content-Length", fmt.Sprintf("%d", len(pdfData)))
	w.Write(pdfData)
}

// healthHandler returns health status
func healthHandler(w http.ResponseWriter, r *http.Request) {
	respondJSON(w, http.StatusOK, APIResponse{
		Success: true,
		Message: "API is healthy",
		Data: map[string]string{
			"status": "ok",
		},
	})
}

// generateLatex renders resume data to LaTeX using template
func generateLatex(resume Resume) (string, error) {
	templatePath, err := resolveResumeTemplatePath(resume.Template, "templates/enhanced-faang-resume.tex.tmpl")
	if err != nil {
		return "", err
	}

	templateContent, err := os.ReadFile(templatePath)
	if err != nil {
		return "", fmt.Errorf("read template file: %w", err)
	}

	tpl, err := texttemplate.New("resume").Funcs(texttemplate.FuncMap{
		"join":        joinStrings,
		"latex":       escapeLatex,
		"md":          markdownToLatex,
		"profileLink": normalizeProfileLink,
		"phoneLink":   normalizePhoneLink,
	}).Parse(string(templateContent))
	if err != nil {
		return "", fmt.Errorf("parse template: %w", err)
	}

	var buf bytes.Buffer
	if err := tpl.Execute(&buf, resume); err != nil {
		return "", fmt.Errorf("render template: %w", err)
	}

	return buf.String(), nil
}

// compilePDF compiles LaTeX to PDF using available compiler
func compilePDF(latexPath, outputDir string) error {
	// Try compilers in order
	compilers := []string{"tectonic", "pdflatex"}

	for _, compiler := range compilers {
		if canCompile(compiler) {
			return executeCompiler(compiler, latexPath, outputDir)
		}
	}

	return fmt.Errorf("no LaTeX compiler found; install pdflatex or tectonic")
}

// joinStrings joins template slices and matches template call order: join .Items ", "
func joinStrings(items []string, sep string) string {
	result := ""
	for i, item := range items {
		if i > 0 {
			result += sep
		}
		result += item
	}
	return result
}

// respondJSON writes JSON response
func respondJSON(w http.ResponseWriter, statusCode int, response APIResponse) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(response)
}
