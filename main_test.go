package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRenderResumeWritesLatex(t *testing.T) {
	tempDir := t.TempDir()
	inputPath := filepath.Join(tempDir, "resume.json")
	outputDir := filepath.Join(tempDir, "output")

	input := []byte(`{
		"name": "Ada Lovelace",
		"title": "Software Engineer",
		"phone": "+1-555-0100",
		"email": "ada@example.com",
		"website": "adalovelace.dev",
		"github": "ada",
		"linkedin": "ada-lovelace",
		"summary": "Builds reliable software.",
		"skills": [],
		"experience": [],
		"projects": [],
		"education": [{"school":"Example University","degree":"BSc Computer Science","location":"London","start":"2018","end":"2022"}],
		"certifications": []
	}`)

	if err := os.WriteFile(inputPath, input, 0o644); err != nil {
		t.Fatalf("write input json: %v", err)
	}

	if err := renderResume(inputPath, "templates/resume.tex.tmpl", outputDir); err != nil {
		t.Fatalf("renderResume returned error: %v", err)
	}

	latexPath := filepath.Join(outputDir, "resume.tex")
	content, err := os.ReadFile(latexPath)
	if err != nil {
		t.Fatalf("read generated latex: %v", err)
	}

	if !bytes.Contains(content, []byte("Ada Lovelace")) {
		t.Fatalf("generated latex does not contain name: %s", string(content))
	}

	if !bytes.Contains(content, []byte("Education")) {
		t.Fatalf("generated latex does not contain education section: %s", string(content))
	}
}

func TestRenderResumeIncludesOpenSourceContributionsSection(t *testing.T) {
	tempDir := t.TempDir()
	inputPath := filepath.Join(tempDir, "resume.json")
	outputDir := filepath.Join(tempDir, "output")

	input := []byte(`{
		"name": "Ada Lovelace",
		"title": "Software Engineer",
		"phone": "+1-555-0100",
		"email": "ada@example.com",
		"website": "adalovelace.dev",
		"github": "ada",
		"linkedin": "ada-lovelace",
		"summary": "Builds reliable software.",
		"skills": [],
		"experience": [],
		"projects": [],
		"education": [],
		"openSourceContributions": [{
			"name": "Go Parser",
			"description": "Improved parser performance.",
			"contribution": "Added parser optimizations and benchmarked performance.",
			"link": "https://example.com/go-parser",
			"stars": "123",
			"keywords": ["golang", "parser"]
		}],
		"certifications": []
	}`)

	if err := os.WriteFile(inputPath, input, 0o644); err != nil {
		t.Fatalf("write input json: %v", err)
	}

	if err := renderResume(inputPath, "templates/resume.tex.tmpl", outputDir); err != nil {
		t.Fatalf("renderResume returned error: %v", err)
	}

	latexPath := filepath.Join(outputDir, "resume.tex")
	content, err := os.ReadFile(latexPath)
	if err != nil {
		t.Fatalf("read generated latex: %v", err)
	}

	if !bytes.Contains(content, []byte("Open Source Contributions")) {
		t.Fatalf("generated latex does not contain open source section: %s", string(content))
	}

	if !bytes.Contains(content, []byte("Go Parser")) {
		t.Fatalf("generated latex does not contain contribution name: %s", string(content))
	}

	if !bytes.Contains(content, []byte("Added parser optimizations and benchmarked performance.")) {
		t.Fatalf("generated latex does not contain contribution details: %s", string(content))
	}

	if !bytes.Contains(content, []byte("123")) {
		t.Fatalf("generated latex does not contain repository stars: %s", string(content))
	}
}

func TestRenderResumePreservesFullProfileUrls(t *testing.T) {
	tempDir := t.TempDir()
	inputPath := filepath.Join(tempDir, "resume.json")
	outputDir := filepath.Join(tempDir, "output")

	input := []byte(`{
		"name": "Ada Lovelace",
		"title": "Software Engineer",
		"phone": "+1-555-0100",
		"email": "ada@example.com",
		"website": "adalovelace.dev",
		"github": "https://github.com/ada",
		"linkedin": "https://linkedin.com/in/ada-lovelace",
		"summary": "Builds reliable software.",
		"skills": [],
		"experience": [],
		"projects": [],
		"education": [],
		"certifications": []
	}`)

	if err := os.WriteFile(inputPath, input, 0o644); err != nil {
		t.Fatalf("write input json: %v", err)
	}

	if err := renderResume(inputPath, "templates/resume.tex.tmpl", outputDir); err != nil {
		t.Fatalf("renderResume returned error: %v", err)
	}

	latexPath := filepath.Join(outputDir, "resume.tex")
	content, err := os.ReadFile(latexPath)
	if err != nil {
		t.Fatalf("read generated latex: %v", err)
	}

	if bytes.Contains(content, []byte("https://github.com/https://github.com/ada")) {
		t.Fatalf("generated latex duplicated github url: %s", string(content))
	}

	if bytes.Contains(content, []byte("https://linkedin.com/in/https://linkedin.com/in/ada-lovelace")) {
		t.Fatalf("generated latex duplicated linkedin url: %s", string(content))
	}

	if !bytes.Contains(content, []byte("\\href{https://github.com/ada}")) {
		t.Fatalf("generated latex does not preserve github url: %s", string(content))
	}

	if !bytes.Contains(content, []byte("\\href{https://linkedin.com/in/ada-lovelace}")) {
		t.Fatalf("generated latex does not preserve linkedin url: %s", string(content))
	}
}

func TestResolveResumeTemplatePathUsesNamedTemplate(t *testing.T) {
	templatePath, err := resolveResumeTemplatePath("enhanced-faang-resume", "templates/resume.tex.tmpl")
	if err != nil {
		t.Fatalf("resolveResumeTemplatePath returned error: %v", err)
	}

	if templatePath != "templates/enhanced-faang-resume.tex.tmpl" {
		t.Fatalf("expected named template path, got %q", templatePath)
	}
}

func TestRenderResumeUsesSelectedTemplate(t *testing.T) {
	tempDir := t.TempDir()
	inputPath := filepath.Join(tempDir, "resume.json")
	outputDir := filepath.Join(tempDir, "output")

	input := []byte(`{
		"name": "Ada Lovelace",
		"title": "Software Engineer",
		"phone": "+1-555-0100",
		"email": "ada@example.com",
		"template": "enhanced-faang-resume",
		"website": "adalovelace.dev",
		"github": "ada",
		"linkedin": "ada-lovelace",
		"summary": "Builds reliable software.",
		"skills": [],
		"experience": [],
		"projects": [],
		"education": [],
		"certifications": []
	}`)

	if err := os.WriteFile(inputPath, input, 0o644); err != nil {
		t.Fatalf("write input json: %v", err)
	}

	if err := renderResume(inputPath, "templates/does-not-exist.tex.tmpl", outputDir); err != nil {
		t.Fatalf("renderResume returned error: %v", err)
	}

	latexPath := filepath.Join(outputDir, "resume.tex")
	content, err := os.ReadFile(latexPath)
	if err != nil {
		t.Fatalf("read generated latex: %v", err)
	}

	if !bytes.Contains(content, []byte("enhanced FAANG resume template")) {
		t.Fatalf("generated latex does not use the selected template: %s", string(content))
	}
}

func TestParseResumeRejectsUnsupportedVersion(t *testing.T) {
	_, err := parseResume([]byte(`{"version":2,"name":"Ada"}`))
	if err == nil {
		t.Fatal("expected unsupported version error")
	}
	if !strings.Contains(err.Error(), "unsupported resume version") {
		t.Fatalf("expected unsupported version error, got %v", err)
	}
}

func TestMarkdownFormattingRendersToLatex(t *testing.T) {
	rendered := markdownToLatex("Use **bold** and *italic* text")

	if !strings.Contains(rendered, "\\textbf{bold}") {
		t.Fatalf("expected bold markdown to render to latex, got %q", rendered)
	}

	if !strings.Contains(rendered, "\\textit{italic}") {
		t.Fatalf("expected italic markdown to render to latex, got %q", rendered)
	}
}
