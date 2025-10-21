package syntax

import (
	"path/filepath"
	"strings"
)

// Language represents a programming language
type Language struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Extensions  []string `json:"extensions"`
	Aliases     []string `json:"aliases"`
	MimeTypes   []string `json:"mime_types"`
	Color       string   `json:"color"`
	AceMode     string   `json:"ace_mode"`
	CodeMirror  string   `json:"codemirror_mode"`
	Highlighter string   `json:"highlighter"`
}

// LanguageDetector detects programming language from file extension or content
type LanguageDetector struct {
	languages      map[string]*Language
	extensionMap   map[string]*Language
	filenameMap    map[string]*Language
	firstLineMap   map[string]*Language
	mimeTypeMap    map[string]*Language
}

// NewLanguageDetector creates a new language detector
func NewLanguageDetector() *LanguageDetector {
	ld := &LanguageDetector{
		languages:    make(map[string]*Language),
		extensionMap: make(map[string]*Language),
		filenameMap:  make(map[string]*Language),
		firstLineMap: make(map[string]*Language),
		mimeTypeMap:  make(map[string]*Language),
	}
	
	ld.loadLanguages()
	return ld
}

// loadLanguages loads all supported languages
func (ld *LanguageDetector) loadLanguages() {
	languages := []Language{
		// Web languages
		{
			ID:          "javascript",
			Name:        "JavaScript",
			Extensions:  []string{".js", ".mjs", ".jsx"},
			Aliases:     []string{"js", "node"},
			MimeTypes:   []string{"application/javascript", "text/javascript"},
			Color:       "#f1e05a",
			AceMode:     "javascript",
			CodeMirror:  "javascript",
			Highlighter: "javascript",
		},
		{
			ID:          "typescript",
			Name:        "TypeScript",
			Extensions:  []string{".ts", ".tsx", ".mts", ".cts"},
			Aliases:     []string{"ts"},
			MimeTypes:   []string{"application/typescript", "text/typescript"},
			Color:       "#2b7489",
			AceMode:     "typescript",
			CodeMirror:  "javascript",
			Highlighter: "typescript",
		},
		{
			ID:          "html",
			Name:        "HTML",
			Extensions:  []string{".html", ".htm", ".xhtml"},
			Aliases:     []string{"xhtml"},
			MimeTypes:   []string{"text/html", "application/xhtml+xml"},
			Color:       "#e34c26",
			AceMode:     "html",
			CodeMirror:  "xml",
			Highlighter: "html",
		},
		{
			ID:          "css",
			Name:        "CSS",
			Extensions:  []string{".css"},
			MimeTypes:   []string{"text/css"},
			Color:       "#563d7c",
			AceMode:     "css",
			CodeMirror:  "css",
			Highlighter: "css",
		},
		{
			ID:          "scss",
			Name:        "SCSS",
			Extensions:  []string{".scss", ".sass"},
			Aliases:     []string{"sass"},
			MimeTypes:   []string{"text/x-scss", "text/x-sass"},
			Color:       "#c6538c",
			AceMode:     "scss",
			CodeMirror:  "css",
			Highlighter: "scss",
		},
		
		// Backend languages
		{
			ID:          "go",
			Name:        "Go",
			Extensions:  []string{".go"},
			Aliases:     []string{"golang"},
			MimeTypes:   []string{"text/x-go", "application/x-go"},
			Color:       "#00ADD8",
			AceMode:     "golang",
			CodeMirror:  "go",
			Highlighter: "go",
		},
		{
			ID:          "python",
			Name:        "Python",
			Extensions:  []string{".py", ".pyw", ".pyc", ".pyo", ".pyi"},
			Aliases:     []string{"py"},
			MimeTypes:   []string{"text/x-python", "application/x-python"},
			Color:       "#3572A5",
			AceMode:     "python",
			CodeMirror:  "python",
			Highlighter: "python",
		},
		{
			ID:          "java",
			Name:        "Java",
			Extensions:  []string{".java", ".class", ".jar"},
			MimeTypes:   []string{"text/x-java", "application/x-java"},
			Color:       "#b07219",
			AceMode:     "java",
			CodeMirror:  "text/x-java",
			Highlighter: "java",
		},
		{
			ID:          "csharp",
			Name:        "C#",
			Extensions:  []string{".cs", ".csx"},
			Aliases:     []string{"c#", "cs"},
			MimeTypes:   []string{"text/x-csharp"},
			Color:       "#178600",
			AceMode:     "csharp",
			CodeMirror:  "text/x-csharp",
			Highlighter: "csharp",
		},
		{
			ID:          "php",
			Name:        "PHP",
			Extensions:  []string{".php", ".phtml", ".php3", ".php4", ".php5", ".phps"},
			MimeTypes:   []string{"text/x-php", "application/x-php"},
			Color:       "#4F5D95",
			AceMode:     "php",
			CodeMirror:  "php",
			Highlighter: "php",
		},
		{
			ID:          "ruby",
			Name:        "Ruby",
			Extensions:  []string{".rb", ".rbw", ".rake", ".gemspec"},
			Aliases:     []string{"rb"},
			MimeTypes:   []string{"text/x-ruby", "application/x-ruby"},
			Color:       "#701516",
			AceMode:     "ruby",
			CodeMirror:  "ruby",
			Highlighter: "ruby",
		},
		{
			ID:          "rust",
			Name:        "Rust",
			Extensions:  []string{".rs", ".rlib"},
			Aliases:     []string{"rs"},
			MimeTypes:   []string{"text/x-rust"},
			Color:       "#dea584",
			AceMode:     "rust",
			CodeMirror:  "rust",
			Highlighter: "rust",
		},
		
		// C/C++ family
		{
			ID:          "c",
			Name:        "C",
			Extensions:  []string{".c", ".h"},
			MimeTypes:   []string{"text/x-c", "text/x-csrc"},
			Color:       "#555555",
			AceMode:     "c_cpp",
			CodeMirror:  "text/x-csrc",
			Highlighter: "c",
		},
		{
			ID:          "cpp",
			Name:        "C++",
			Extensions:  []string{".cpp", ".cc", ".cxx", ".c++", ".hpp", ".hh", ".hxx", ".h++"},
			Aliases:     []string{"c++"},
			MimeTypes:   []string{"text/x-c++", "text/x-c++src"},
			Color:       "#f34b7d",
			AceMode:     "c_cpp",
			CodeMirror:  "text/x-c++src",
			Highlighter: "cpp",
		},
		{
			ID:          "objectivec",
			Name:        "Objective-C",
			Extensions:  []string{".m", ".mm"},
			Aliases:     []string{"objc"},
			MimeTypes:   []string{"text/x-objectivec"},
			Color:       "#438eff",
			AceMode:     "objectivec",
			CodeMirror:  "text/x-objectivec",
			Highlighter: "objectivec",
		},
		
		// Shell/Script languages
		{
			ID:          "shell",
			Name:        "Shell",
			Extensions:  []string{".sh", ".bash", ".zsh", ".fish", ".ksh", ".csh"},
			Aliases:     []string{"bash", "sh", "zsh"},
			MimeTypes:   []string{"text/x-sh", "application/x-sh"},
			Color:       "#89e051",
			AceMode:     "sh",
			CodeMirror:  "shell",
			Highlighter: "bash",
		},
		{
			ID:          "powershell",
			Name:        "PowerShell",
			Extensions:  []string{".ps1", ".psm1", ".psd1"},
			Aliases:     []string{"ps", "ps1"},
			MimeTypes:   []string{"text/x-powershell", "application/x-powershell"},
			Color:       "#012456",
			AceMode:     "powershell",
			CodeMirror:  "powershell",
			Highlighter: "powershell",
		},
		
		// Data formats
		{
			ID:          "json",
			Name:        "JSON",
			Extensions:  []string{".json", ".jsonc", ".json5"},
			MimeTypes:   []string{"application/json", "application/ld+json"},
			Color:       "#292929",
			AceMode:     "json",
			CodeMirror:  "javascript",
			Highlighter: "json",
		},
		{
			ID:          "yaml",
			Name:        "YAML",
			Extensions:  []string{".yml", ".yaml"},
			Aliases:     []string{"yml"},
			MimeTypes:   []string{"text/x-yaml", "application/x-yaml"},
			Color:       "#cb171e",
			AceMode:     "yaml",
			CodeMirror:  "yaml",
			Highlighter: "yaml",
		},
		{
			ID:          "xml",
			Name:        "XML",
			Extensions:  []string{".xml", ".xsd", ".xsl", ".xslt", ".svg"},
			MimeTypes:   []string{"text/xml", "application/xml"},
			Color:       "#0060ac",
			AceMode:     "xml",
			CodeMirror:  "xml",
			Highlighter: "xml",
		},
		{
			ID:          "toml",
			Name:        "TOML",
			Extensions:  []string{".toml"},
			MimeTypes:   []string{"text/x-toml", "application/toml"},
			Color:       "#9c4221",
			AceMode:     "toml",
			CodeMirror:  "toml",
			Highlighter: "toml",
		},
		{
			ID:          "ini",
			Name:        "INI",
			Extensions:  []string{".ini", ".cfg", ".conf", ".config"},
			Aliases:     []string{"cfg", "conf"},
			MimeTypes:   []string{"text/x-ini"},
			Color:       "#d1dbe0",
			AceMode:     "ini",
			CodeMirror:  "properties",
			Highlighter: "ini",
		},
		
		// Documentation
		{
			ID:          "markdown",
			Name:        "Markdown",
			Extensions:  []string{".md", ".markdown", ".mdown", ".mdx"},
			Aliases:     []string{"md"},
			MimeTypes:   []string{"text/markdown", "text/x-markdown"},
			Color:       "#083fa1",
			AceMode:     "markdown",
			CodeMirror:  "markdown",
			Highlighter: "markdown",
		},
		{
			ID:          "asciidoc",
			Name:        "AsciiDoc",
			Extensions:  []string{".adoc", ".asciidoc", ".asc"},
			Aliases:     []string{"adoc"},
			MimeTypes:   []string{"text/x-asciidoc"},
			Color:       "#73a0c5",
			AceMode:     "asciidoc",
			CodeMirror:  "asciidoc",
			Highlighter: "asciidoc",
		},
		{
			ID:          "latex",
			Name:        "LaTeX",
			Extensions:  []string{".tex", ".latex", ".ltx"},
			Aliases:     []string{"tex"},
			MimeTypes:   []string{"text/x-latex", "application/x-latex"},
			Color:       "#3D6117",
			AceMode:     "latex",
			CodeMirror:  "stex",
			Highlighter: "latex",
		},
		
		// Other popular languages
		{
			ID:          "swift",
			Name:        "Swift",
			Extensions:  []string{".swift"},
			MimeTypes:   []string{"text/x-swift"},
			Color:       "#ffac45",
			AceMode:     "swift",
			CodeMirror:  "swift",
			Highlighter: "swift",
		},
		{
			ID:          "kotlin",
			Name:        "Kotlin",
			Extensions:  []string{".kt", ".kts", ".ktm"},
			MimeTypes:   []string{"text/x-kotlin"},
			Color:       "#F18E33",
			AceMode:     "kotlin",
			CodeMirror:  "text/x-kotlin",
			Highlighter: "kotlin",
		},
		{
			ID:          "scala",
			Name:        "Scala",
			Extensions:  []string{".scala", ".sc"},
			MimeTypes:   []string{"text/x-scala"},
			Color:       "#c22d40",
			AceMode:     "scala",
			CodeMirror:  "text/x-scala",
			Highlighter: "scala",
		},
		{
			ID:          "r",
			Name:        "R",
			Extensions:  []string{".r", ".R", ".rmd", ".Rmd"},
			MimeTypes:   []string{"text/x-rsrc"},
			Color:       "#198CE7",
			AceMode:     "r",
			CodeMirror:  "r",
			Highlighter: "r",
		},
		{
			ID:          "julia",
			Name:        "Julia",
			Extensions:  []string{".jl"},
			MimeTypes:   []string{"text/x-julia"},
			Color:       "#a270ba",
			AceMode:     "julia",
			CodeMirror:  "julia",
			Highlighter: "julia",
		},
		{
			ID:          "perl",
			Name:        "Perl",
			Extensions:  []string{".pl", ".pm", ".pod", ".t"},
			MimeTypes:   []string{"text/x-perl", "application/x-perl"},
			Color:       "#0298c3",
			AceMode:     "perl",
			CodeMirror:  "perl",
			Highlighter: "perl",
		},
		{
			ID:          "lua",
			Name:        "Lua",
			Extensions:  []string{".lua"},
			MimeTypes:   []string{"text/x-lua", "application/x-lua"},
			Color:       "#000080",
			AceMode:     "lua",
			CodeMirror:  "lua",
			Highlighter: "lua",
		},
		{
			ID:          "dart",
			Name:        "Dart",
			Extensions:  []string{".dart"},
			MimeTypes:   []string{"text/x-dart", "application/dart"},
			Color:       "#00B4AB",
			AceMode:     "dart",
			CodeMirror:  "dart",
			Highlighter: "dart",
		},
		{
			ID:          "elixir",
			Name:        "Elixir",
			Extensions:  []string{".ex", ".exs"},
			MimeTypes:   []string{"text/x-elixir"},
			Color:       "#6e4a7e",
			AceMode:     "elixir",
			CodeMirror:  "elixir",
			Highlighter: "elixir",
		},
		{
			ID:          "clojure",
			Name:        "Clojure",
			Extensions:  []string{".clj", ".cljs", ".cljc", ".edn"},
			MimeTypes:   []string{"text/x-clojure"},
			Color:       "#db5855",
			AceMode:     "clojure",
			CodeMirror:  "clojure",
			Highlighter: "clojure",
		},
		{
			ID:          "haskell",
			Name:        "Haskell",
			Extensions:  []string{".hs", ".lhs"},
			MimeTypes:   []string{"text/x-haskell"},
			Color:       "#5e5086",
			AceMode:     "haskell",
			CodeMirror:  "haskell",
			Highlighter: "haskell",
		},
		
		// Database
		{
			ID:          "sql",
			Name:        "SQL",
			Extensions:  []string{".sql", ".mysql", ".pgsql", ".sqlite"},
			MimeTypes:   []string{"text/x-sql", "application/sql"},
			Color:       "#e38c00",
			AceMode:     "sql",
			CodeMirror:  "sql",
			Highlighter: "sql",
		},
		
		// Config files
		{
			ID:          "dockerfile",
			Name:        "Dockerfile",
			Extensions:  []string{".dockerfile"},
			Aliases:     []string{"docker"},
			MimeTypes:   []string{"text/x-dockerfile"},
			Color:       "#384d54",
			AceMode:     "dockerfile",
			CodeMirror:  "dockerfile",
			Highlighter: "dockerfile",
		},
		{
			ID:          "makefile",
			Name:        "Makefile",
			Extensions:  []string{".makefile", ".mk"},
			Aliases:     []string{"make"},
			MimeTypes:   []string{"text/x-makefile"},
			Color:       "#427819",
			AceMode:     "makefile",
			CodeMirror:  "cmake",
			Highlighter: "makefile",
		},
		{
			ID:          "nginx",
			Name:        "Nginx",
			Extensions:  []string{".nginx", ".nginxconf"},
			MimeTypes:   []string{"text/x-nginx-conf"},
			Color:       "#009639",
			AceMode:     "nginx",
			CodeMirror:  "nginx",
			Highlighter: "nginx",
		},
		
		// Default/Plain text
		{
			ID:          "text",
			Name:        "Plain Text",
			Extensions:  []string{".txt", ".text", ".log"},
			Aliases:     []string{"plaintext", "plain"},
			MimeTypes:   []string{"text/plain"},
			Color:       "#000000",
			AceMode:     "text",
			CodeMirror:  "text/plain",
			Highlighter: "plaintext",
		},
	}
	
	// Special filename mappings
	filenameLanguages := map[string]string{
		"Dockerfile":     "dockerfile",
		"Makefile":       "makefile",
		"makefile":       "makefile",
		"GNUmakefile":    "makefile",
		"Rakefile":       "ruby",
		"Gemfile":        "ruby",
		"Podfile":        "ruby",
		"Vagrantfile":    "ruby",
		"Berksfile":      "ruby",
		"Thorfile":       "ruby",
		"Guardfile":      "ruby",
		"package.json":   "json",
		"composer.json":  "json",
		"tsconfig.json":  "json",
		"jsconfig.json":  "json",
		".eslintrc.json": "json",
		".babelrc":       "json",
		".prettierrc":    "json",
		"nginx.conf":     "nginx",
		".gitignore":     "ini",
		".gitconfig":     "ini",
		".npmrc":         "ini",
		".env":           "ini",
		"requirements.txt": "text",
		"CMakeLists.txt": "cmake",
		"go.mod":         "go",
		"go.sum":         "go",
		"Cargo.toml":     "toml",
		"Cargo.lock":     "toml",
		"pyproject.toml": "toml",
	}
	
	// Build maps
	for i := range languages {
		lang := &languages[i]
		ld.languages[lang.ID] = lang
		
		// Map extensions
		for _, ext := range lang.Extensions {
			ld.extensionMap[strings.ToLower(ext)] = lang
		}
		
		// Map mime types
		for _, mime := range lang.MimeTypes {
			ld.mimeTypeMap[mime] = lang
		}
	}
	
	// Map special filenames
	for filename, langID := range filenameLanguages {
		if lang, ok := ld.languages[langID]; ok {
			ld.filenameMap[filename] = lang
		}
	}
}

// DetectLanguage detects the programming language from filename and content
func (ld *LanguageDetector) DetectLanguage(filename, content string) *Language {
	// 1. Check exact filename match
	if lang, ok := ld.filenameMap[filename]; ok {
		return lang
	}
	
	// 2. Check file extension
	ext := strings.ToLower(filepath.Ext(filename))
	if lang, ok := ld.extensionMap[ext]; ok {
		return lang
	}
	
	// 3. Check shebang line
	if strings.HasPrefix(content, "#!") {
		firstLine := strings.SplitN(content, "\n", 2)[0]
		lang := ld.detectFromShebang(firstLine)
		if lang != nil {
			return lang
		}
	}
	
	// 4. Default to text
	return ld.languages["text"]
}

// detectFromShebang detects language from shebang line
func (ld *LanguageDetector) detectFromShebang(shebang string) *Language {
	shebang = strings.ToLower(shebang)
	
	// Common shebang patterns
	patterns := map[string]string{
		"python": "python",
		"ruby":   "ruby",
		"perl":   "perl",
		"bash":   "shell",
		"sh":     "shell",
		"zsh":    "shell",
		"fish":   "shell",
		"node":   "javascript",
		"php":    "php",
		"lua":    "lua",
	}
	
	for pattern, langID := range patterns {
		if strings.Contains(shebang, pattern) {
			if lang, ok := ld.languages[langID]; ok {
				return lang
			}
		}
	}
	
	return nil
}

// GetLanguageByID returns a language by its ID
func (ld *LanguageDetector) GetLanguageByID(id string) *Language {
	return ld.languages[id]
}

// GetLanguageByMimeType returns a language by MIME type
func (ld *LanguageDetector) GetLanguageByMimeType(mimeType string) *Language {
	return ld.mimeTypeMap[mimeType]
}

// GetAllLanguages returns all supported languages
func (ld *LanguageDetector) GetAllLanguages() []*Language {
	languages := make([]*Language, 0, len(ld.languages))
	for _, lang := range ld.languages {
		languages = append(languages, lang)
	}
	return languages
}

// GetPopularLanguages returns commonly used languages
func (ld *LanguageDetector) GetPopularLanguages() []*Language {
	popularIDs := []string{
		"javascript", "typescript", "python", "go", "java",
		"csharp", "cpp", "php", "ruby", "swift", "rust",
		"html", "css", "json", "yaml", "markdown", "shell",
		"sql", "dockerfile", "text",
	}
	
	languages := make([]*Language, 0, len(popularIDs))
	for _, id := range popularIDs {
		if lang, ok := ld.languages[id]; ok {
			languages = append(languages, lang)
		}
	}
	return languages
}