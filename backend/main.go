package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	texttemplate "text/template"
)

const currentResumeVersion = 1
const defaultResumeTemplate = "enhanced-faang-resume"

var resumeTemplatePaths = map[string]string{
	defaultResumeTemplate: "templates/enhanced-faang-resume.tex.tmpl",
	"resume":              "templates/enhanced-faang-resume.tex.tmpl",
}

var defaultSectionOrder = []string{
	"summary",
	"skills",
	"experience",
	"openSource",
	"projects",
	"certifications",
	"education",
}

type Resume struct {
	Version int    `json:"version"`
	Name    string `json:"name"`
	Title   string `json:"title"`
	Phone   string `json:"phone"`
	Email   string `json:"email"`

	Website  string `json:"website"`
	Linkedin string `json:"linkedin"`
	Github   string `json:"github"`
	Template string `json:"template,omitempty"`

	Summary      string   `json:"summary"`
	SectionOrder []string `json:"sectionOrder,omitempty"`

	Skills                  []SkillCategory          `json:"skills"`
	Experience              []Experience             `json:"experience"`
	Projects                []Project                `json:"projects"`
	Education               []Education              `json:"education"`
	OpenSourceContributions []OpenSourceContribution `json:"openSourceContributions"`
	Certifications          []Certification          `json:"certifications"`
}

type resumeInput struct {
	Version *int   `json:"version"`
	Name    string `json:"name"`
	Title   string `json:"title"`
	Phone   string `json:"phone"`
	Email   string `json:"email"`

	Website  string `json:"website"`
	Linkedin string `json:"linkedin"`
	Github   string `json:"github"`
	Template string `json:"template,omitempty"`

	Summary      string   `json:"summary"`
	SectionOrder []string `json:"sectionOrder,omitempty"`

	Skills                  []SkillCategory          `json:"skills"`
	Experience              []Experience             `json:"experience"`
	Projects                []Project                `json:"projects"`
	Education               []Education              `json:"education"`
	OpenSourceContributions []OpenSourceContribution `json:"openSourceContributions"`
	Certifications          []Certification          `json:"certifications"`
}

type Experience struct {
	Company  string   `json:"company"`
	Title    string   `json:"title"`
	Location string   `json:"location"`
	Start    string   `json:"start"`
	End      string   `json:"end"`
	Bullets  []string `json:"bullets"`
}

type Project struct {
	Name         string   `json:"name"`
	Link         string   `json:"link"`
	Technologies []string `json:"technologies"`
	Bullets      []string `json:"bullets"`
}

type Education struct {
	School   string `json:"school"`
	Degree   string `json:"degree"`
	Location string `json:"location"`
	Start    string `json:"start"`
	End      string `json:"end"`
}

type SkillCategory struct {
	Title string   `json:"title"`
	Items []string `json:"items"`
}

type OpenSourceContribution struct {
	Name         string   `json:"name"`
	Description  string   `json:"description"`
	Role         string   `json:"role"`
	Contribution string   `json:"contribution"`
	Link         string   `json:"link"`
	RepoLink     string   `json:"repoLink"`
	Stars        string   `json:"stars"`
	Keywords     []string `json:"keywords"`
}

type Certification struct {
	Name string `json:"name"`
}

func main() {
	inputPath := flag.String("input", "resume.json", "path to the resume JSON file")
	templatePath := flag.String("template", "templates/enhanced-faang-resume.tex.tmpl", "path to the LaTeX template")
	outputDir := flag.String("output", "output", "directory for generated files")
	outputFormat := flag.String("format", "both", "output format: latex, pdf, or both")
	dbPath := flag.String("db", defaultResumeDBPath, "sqlite database file path (server mode)")
	seedPath := flag.String("seed", defaultSeedResumePath, "seed resume json file path (server mode)")
	serverMode := flag.String("server", "", "start HTTP server on specified port (e.g., ':8080')")
	flag.Parse()

	if *serverMode != "" {
		if err := initResumeStore(*dbPath, *seedPath); err != nil {
			fmt.Fprintf(os.Stderr, "failed to initialize resume store: %v\n", err)
			os.Exit(1)
		}
		if err := startServer(*serverMode); err != nil {
			fmt.Fprintf(os.Stderr, "server failed: %v\n", err)
			os.Exit(1)
		}
		return
	}

	format := strings.ToLower(strings.TrimSpace(*outputFormat))
	if format != "latex" && format != "pdf" && format != "both" {
		fmt.Fprintf(os.Stderr, "invalid -format value %q; expected one of: latex, pdf, both\n", *outputFormat)
		os.Exit(2)
	}

	if err := renderResume(*inputPath, *templatePath, *outputDir); err != nil {
		fmt.Fprintf(os.Stderr, "resume generation failed: %v\n", err)
		os.Exit(1)
	}

	if format == "latex" {
		return
	}

	if err := renderPDF(*outputDir); err != nil {
		fmt.Fprintf(os.Stderr, "pdf generation failed: %v\n", err)
		os.Exit(1)
	}
}

func renderResume(inputPath, templatePath, outputDir string) error {
	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		return fmt.Errorf("create output directory: %w", err)
	}

	data, err := os.ReadFile(inputPath)
	if err != nil {
		return fmt.Errorf("read input file: %w", err)
	}

	resume, err := parseResume(data)
	if err != nil {
		return err
	}

	resolvedTemplatePath, err := resolveResumeTemplatePath(resume.Template, templatePath)
	if err != nil {
		return err
	}

	templateContent, err := os.ReadFile(resolvedTemplatePath)
	if err != nil {
		return fmt.Errorf("read template file: %w", err)
	}

	var tpl *texttemplate.Template
	tpl, err = texttemplate.New("resume").Funcs(texttemplate.FuncMap{
		"join":        strings.Join,
		"latex":       escapeLatex,
		"md":          markdownToLatex,
		"profileLink": normalizeProfileLink,
		"phoneLink":   normalizePhoneLink,
		"hasSkills":   hasSkillItems,
		"sectionBlock": func(name string, r Resume) (string, error) {
			var buf bytes.Buffer
			if execErr := tpl.ExecuteTemplate(&buf, "section_"+name, r); execErr != nil {
				return "", fmt.Errorf("render section %q: %w", name, execErr)
			}
			return buf.String(), nil
		},
	}).Parse(string(templateContent))
	if err != nil {
		return fmt.Errorf("parse template: %w", err)
	}

	latexPath := filepath.Join(outputDir, "resume.tex")
	out, err := os.Create(latexPath)
	if err != nil {
		return fmt.Errorf("create latex output: %w", err)
	}
	defer out.Close()

	if err := tpl.Execute(out, resume); err != nil {
		return fmt.Errorf("render latex template: %w", err)
	}

	fmt.Printf("Generated %s\n", latexPath)
	return nil
}

func parseResume(data []byte) (Resume, error) {
	var raw resumeInput
	if err := json.Unmarshal(data, &raw); err != nil {
		return Resume{}, fmt.Errorf("parse resume json: %w", err)
	}

	version := currentResumeVersion
	if raw.Version != nil {
		version = *raw.Version
	}

	if version < 1 {
		return Resume{}, fmt.Errorf("unsupported resume version %d", version)
	}
	if version > currentResumeVersion {
		return Resume{}, fmt.Errorf("unsupported resume version %d (current supported version is %d)", version, currentResumeVersion)
	}

	sectionOrder := raw.SectionOrder
	if len(sectionOrder) == 0 {
		sectionOrder = defaultSectionOrder
	}

	return Resume{
		Version:                 version,
		Name:                    raw.Name,
		Title:                   raw.Title,
		Phone:                   raw.Phone,
		Email:                   raw.Email,
		Website:                 raw.Website,
		Linkedin:                raw.Linkedin,
		Github:                  raw.Github,
		Template:                raw.Template,
		Summary:                 raw.Summary,
		SectionOrder:            sectionOrder,
		Skills:                  raw.Skills,
		Experience:              raw.Experience,
		Projects:                raw.Projects,
		Education:               raw.Education,
		OpenSourceContributions: raw.OpenSourceContributions,
		Certifications:          raw.Certifications,
	}, nil
}

func escapeLatex(value string) string {
	replacements := strings.NewReplacer(
		"#", "\\#",
		"%", "\\%",
		"&", "\\&",
		"_", "\\_",
		"{", "\\{",
		"}", "\\}",
		"~", "\\textasciitilde{}",
		"^", "\\textasciicircum{}",
	)
	return replacements.Replace(value)
}

func markdownToLatex(value string) string {
	if value == "" {
		return ""
	}

	var builder strings.Builder
	for i := 0; i < len(value); {
		if strings.HasPrefix(value[i:], "**") {
			end := strings.Index(value[i+2:], "**")
			if end >= 0 {
				builder.WriteString("\\textbf{")
				builder.WriteString(escapeLatex(value[i+2 : i+2+end]))
				builder.WriteString("}")
				i += 2 + end + 2
				continue
			}
		}

		if strings.HasPrefix(value[i:], "*") {
			end := strings.Index(value[i+1:], "*")
			if end >= 0 {
				builder.WriteString("\\textit{")
				builder.WriteString(escapeLatex(value[i+1 : i+1+end]))
				builder.WriteString("}")
				i += 1 + end + 1
				continue
			}
		}

		builder.WriteString(escapeLatex(string(value[i])))
		i++
	}

	return builder.String()
}

func resolveResumeTemplatePath(selectedTemplate, fallbackTemplatePath string) (string, error) {
	selectedTemplate = strings.TrimSpace(selectedTemplate)
	if selectedTemplate == "" {
		if templateFileExists(fallbackTemplatePath) {
			return fallbackTemplatePath, nil
		}
		if fallbackTemplatePath == "templates/resume.tex.tmpl" {
			return "templates/enhanced-faang-resume.tex.tmpl", nil
		}
		return fallbackTemplatePath, nil
	}

	if templatePath, ok := resumeTemplatePaths[selectedTemplate]; ok {
		if templateFileExists(templatePath) {
			return templatePath, nil
		}
		if templatePath == "templates/resume.tex.tmpl" {
			return "templates/enhanced-faang-resume.tex.tmpl", nil
		}
		return templatePath, nil
	}

	if strings.Contains(selectedTemplate, "/") {
		return selectedTemplate, nil
	}

	return "", fmt.Errorf("unsupported resume template %q", selectedTemplate)
}

func templateFileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func normalizeProfileLink(prefix, value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}

	if strings.HasPrefix(value, "http://") || strings.HasPrefix(value, "https://") {
		return value
	}

	return prefix + value
}

func normalizePhoneLink(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}

	replacer := strings.NewReplacer(" ", "", "-", "", "(", "", ")", "")
	return "tel:" + replacer.Replace(value)
}

// hasSkillItems reports whether any skill category has at least one item.
func hasSkillItems(categories []SkillCategory) bool {
	for _, category := range categories {
		if len(category.Items) > 0 {
			return true
		}
	}
	return false
}

func renderPDF(outputDir string) error {
	latexPath := filepath.Join(outputDir, "resume.tex")
	if _, err := os.Stat(latexPath); err != nil {
		return fmt.Errorf("latex file not found: %w", err)
	}

	for _, compiler := range []string{"pdflatex", "tectonic"} {
		if _, err := exec.LookPath(compiler); err == nil {
			var cmd *exec.Cmd
			if compiler == "tectonic" {
				cmd = exec.Command(compiler, "--outdir", outputDir, latexPath)
			} else {
				cmd = exec.Command(compiler, "-interaction=nonstopmode", "-halt-on-error", "-output-directory", outputDir, latexPath)
			}
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			if err := cmd.Run(); err != nil {
				return fmt.Errorf("compile pdf with %s: %w", compiler, err)
			}
			fmt.Printf("Generated %s.pdf\n", filepath.Join(outputDir, "resume"))
			return nil
		}
	}

	return fmt.Errorf("no LaTeX compiler found; install pdflatex or tectonic to generate a PDF")
}

// canCompile checks if a compiler is available in PATH
func canCompile(compiler string) bool {
	_, err := exec.LookPath(compiler)
	return err == nil
}

// executeCompiler runs the specified LaTeX compiler
func executeCompiler(compiler, latexPath, outputDir string) error {
	var cmd *exec.Cmd
	if compiler == "tectonic" {
		cmd = exec.Command(compiler, "--outdir", outputDir, latexPath)
	} else {
		cmd = exec.Command(compiler, "-interaction=nonstopmode", "-halt-on-error", "-output-directory", outputDir, latexPath)
	}
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("compile pdf with %s: %w", compiler, err)
	}
	return nil
}

// startServer starts the HTTP API server
func startServer(port string) error {
	if resumeStore == nil {
		return fmt.Errorf("resume store not initialized")
	}

	http.HandleFunc("/api/resumes", resumesHandler)
	http.HandleFunc("/api/resumes/", resumeByIDHandler)
	http.HandleFunc("/health", healthHandler)
	http.HandleFunc("/api/resume/latex", generateResumeLatexHandler)
	http.HandleFunc("/api/resume/pdf", generateResumePDFHandler)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
	})

	fmt.Printf("Starting Resume API server on http://localhost%s\n", port)
	fmt.Println("Available endpoints:")
	fmt.Println("  GET  /api/resumes      - List resumes")
	fmt.Println("  POST /api/resumes      - Create resume")
	fmt.Println("  GET  /api/resumes/{id} - Get resume")
	fmt.Println("  PUT  /api/resumes/{id} - Update resume")
	fmt.Println("  POST /api/resume/latex - Generate LaTeX")
	fmt.Println("  POST /api/resume/pdf   - Generate PDF")
	fmt.Println("  GET  /health           - Health check")
	return http.ListenAndServe(port, nil)
}
