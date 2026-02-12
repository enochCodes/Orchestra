package buildpack

type Framework struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	Description  string `json:"description"`
	DefaultBuild string `json:"default_build"`
	DefaultStart string `json:"default_start"`
}

type AppType struct {
	ID         string      `json:"id"`
	Name       string      `json:"name"`
	Frameworks []Framework `json:"frameworks"`
}

func GetMetadata() []AppType {
	return []AppType{
		{
			ID:   "web_service",
			Name: "Web Service",
			Frameworks: []Framework{
				{ID: "node", Name: "Node.js", Description: "Javascript Runtime", DefaultBuild: "npm install", DefaultStart: "npm start"},
				{ID: "go", Name: "Go", Description: "High performance compiled language", DefaultBuild: "go build -o main", DefaultStart: "./main"},
				{ID: "python", Name: "Python", Description: "Versatile scripting language", DefaultBuild: "pip install -r requirements.txt", DefaultStart: "python app.py"},
				{ID: "rust", Name: "Rust", Description: "Memory safe systems language", DefaultBuild: "cargo build --release", DefaultStart: "./target/release/app"},
			},
		},
		{
			ID:   "static_site",
			Name: "Static Site",
			Frameworks: []Framework{
				{ID: "nextjs-static", Name: "Next.js (Static)", Description: "React Framework (Static Export)", DefaultBuild: "npm run build", DefaultStart: "nginx"},
				{ID: "gatsby", Name: "Gatsby", Description: "Static Site Generator", DefaultBuild: "npm run build", DefaultStart: "nginx"},
			},
		},
	}
}
