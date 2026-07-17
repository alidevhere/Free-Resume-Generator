package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	texttemplate "text/template"
	"time"
)

type APIResumeRequest struct {
	Resume Resume `json:"resume"`
}

type APIResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

type ResumeSaveRequest struct {
	Name   string `json:"name"`
	Resume Resume `json:"resume"`
}

type ResumeListResponse struct {
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	CreatedAt string `json:"createdAt"`
	UpdatedAt string `json:"updatedAt"`
}

type ResumeRecordResponse struct {
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	CreatedAt string `json:"createdAt"`
	UpdatedAt string `json:"updatedAt"`
	Resume    Resume `json:"resume"`
}

func resumesHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/api/resumes" {
		http.NotFound(w, r)
		return
	}

	switch r.Method {
	case http.MethodGet:
		if resumeStore == nil {
			respondJSON(w, http.StatusInternalServerError, APIResponse{Success: false, Message: "resume store is not initialized"})
			return
		}

		resumes, err := resumeStore.List()
		if err != nil {
			respondJSON(w, http.StatusInternalServerError, APIResponse{Success: false, Message: err.Error()})
			return
		}

		response := make([]ResumeListResponse, 0, len(resumes))
		for _, record := range resumes {
			response = append(response, ResumeListResponse{
				ID:        record.ID,
				Name:      record.Name,
				CreatedAt: record.CreatedAt.Format(timeRFC3339Display),
				UpdatedAt: record.UpdatedAt.Format(timeRFC3339Display),
			})
		}

		respondJSON(w, http.StatusOK, APIResponse{Success: true, Data: response})
	case http.MethodPost:
		if resumeStore == nil {
			respondJSON(w, http.StatusInternalServerError, APIResponse{Success: false, Message: "resume store is not initialized"})
			return
		}

		var req ResumeSaveRequest
		if r.Body != nil {
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil && err.Error() != "EOF" {
				respondJSON(w, http.StatusBadRequest, APIResponse{Success: false, Message: fmt.Sprintf("invalid request body: %v", err)})
				return
			}
		}

		record, err := resumeStore.Create(req.Name, req.Resume)
		if err != nil {
			respondJSON(w, http.StatusInternalServerError, APIResponse{Success: false, Message: err.Error()})
			return
		}

		respondJSON(w, http.StatusCreated, APIResponse{Success: true, Data: toResumeRecordResponse(record)})
	default:
		respondJSON(w, http.StatusMethodNotAllowed, APIResponse{Success: false, Message: "Method not allowed. Use GET or POST."})
	}
}

func resumeByIDHandler(w http.ResponseWriter, r *http.Request) {
	id, ok := parseResumeID(r.URL.Path)
	if !ok {
		http.NotFound(w, r)
		return
	}

	switch r.Method {
	case http.MethodGet:
		record, err := resumeStore.Get(id)
		if err != nil {
			respondResumeError(w, err)
			return
		}
		respondJSON(w, http.StatusOK, APIResponse{Success: true, Data: toResumeRecordResponse(record)})
	case http.MethodPut:
		var req ResumeSaveRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			respondJSON(w, http.StatusBadRequest, APIResponse{Success: false, Message: fmt.Sprintf("invalid request body: %v", err)})
			return
		}

		record, err := resumeStore.Update(id, req.Name, req.Resume)
		if err != nil {
			respondResumeError(w, err)
			return
		}

		respondJSON(w, http.StatusOK, APIResponse{Success: true, Data: toResumeRecordResponse(record)})
	case http.MethodDelete:
		if err := resumeStore.Delete(id); err != nil {
			respondResumeError(w, err)
			return
		}
		respondJSON(w, http.StatusOK, APIResponse{Success: true, Message: "resume deleted"})
	default:
		respondJSON(w, http.StatusMethodNotAllowed, APIResponse{Success: false, Message: "Method not allowed. Use GET, PUT, or DELETE."})
	}
}

const timeRFC3339Display = time.RFC3339

func parseResumeID(path string) (int64, bool) {
	trimmed := strings.TrimPrefix(path, "/api/resumes/")
	if trimmed == "" || strings.Contains(trimmed, "/") {
		return 0, false
	}

	id, err := strconv.ParseInt(trimmed, 10, 64)
	if err != nil {
		return 0, false
	}

	return id, true
}

func respondResumeError(w http.ResponseWriter, err error) {
	if strings.Contains(err.Error(), "not found") {
		respondJSON(w, http.StatusNotFound, APIResponse{Success: false, Message: err.Error()})
		return
	}
	respondJSON(w, http.StatusInternalServerError, APIResponse{Success: false, Message: err.Error()})
}

func toResumeRecordResponse(record ResumeRecord) ResumeRecordResponse {
	return ResumeRecordResponse{
		ID:        record.ID,
		Name:      record.Name,
		CreatedAt: record.CreatedAt.Format(timeRFC3339Display),
		UpdatedAt: record.UpdatedAt.Format(timeRFC3339Display),
		Resume:    record.Resume,
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
