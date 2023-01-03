package main

import (
	"embed"
	"encoding/json"
	"flag"
	"net/http"
	"os"
	"os/exec"
)

//go:embed build/*
var PSBuildFolder embed.FS

//go:embed files/*
var FileFolder embed.FS

// Default suffix for the embedded FileFolder file system.
var FileFolderSuffix = ".jsext"

// Hold the current path of the jsext tool for later use.
var CurrentPath string

// Set up default files for the initialization process.
var (
	github_tag_url = "https://api.github.com/repos/Nigel2392/jsext/tags"
	ModFilename    = "go.mod"
	IndexFilename  = "index.html"
	ServerFilenme  = "server.go"

	SourceDir       = "src"
	StaticDir       = "static"
	BuildDir        = "build"
	FilenameMain    = "main.go"
	FilenamePlain   = "plain.go"
	FilenamesBuild  = []string{"build.ps1", "build-run.ps1", "build-tiny.ps1"}
	FileNamesSrc    = []string{"main.go", "urls.go", "views.go"}
	FileNamesStatic = []string{"initjsext.js", "wasm_exec_go.js", "wasm_exec_tiny.js"}
)

// Run the jsext tool.
func main() {
	// Get the current path
	var err error
	CurrentPath, err = os.Getwd()
	if err != nil {
		panic(err)
	}

	// Commandline flags
	var initFlag = flag.Bool("init", false, "Initialize a project")
	var plainFlag = flag.Bool("plain", false, "Create a plain project")
	var projectName = flag.String("n", "", "Name of the project to initialize.")
	var vsCodeConfig = flag.Bool("vscode", false, "Create a vscode config file.")
	flag.Parse()

	// Make sure we create all files in the current directory
	os.Chdir(CurrentPath)

	// Check if we have a project name
	if *projectName == "" {
		panic("Please provide a project name with -n <project name>.")
	}

	// Check if the project already exists
	if _, err := os.Stat(*projectName); !os.IsNotExist(err) {
		panic("Directory with project name already exists.")
	}

	// Run the flag instructions.
	if *initFlag && !*plainFlag {
		InitProject(*projectName)
	} else if *plainFlag {
		InitPlain(*projectName)
	}
	if *vsCodeConfig {
		CreateVsCodeConfig()
	}
}

// Create a plain project.
func InitPlain(projectName string) {

	initDefault(projectName)

	initBuildFiles(projectName)

	// Create the file in the source folder
	// Write the plain.go file to the source folder with the name main.go.
	var f = ReadFileFolder(SourceDir + "/" + FilenamePlain)
	os.WriteFile(projectName+"/"+SourceDir+"/"+FilenameMain, f, 0644)

	// Update go mod file
	initGoMod(projectName)
}

// Create a project with some files already in place to get familiar.
func InitProject(projectName string) {

	initDefault(projectName)

	initBuildFiles(projectName)

	initStaticFiles(projectName)

	// Initialize the source files from the slice.
	// This excludes the plain.go file.
	for _, file := range FileNamesSrc {
		var f = ReadFileFolder(SourceDir + "/" + file)
		os.WriteFile(projectName+"/"+SourceDir+"/"+file, f, 0644)
	}

	initGoMod(projectName)
}

// Create a default vscode config file with GOOS and GOARCH set to js and wasm respectively.
func CreateVsCodeConfig() {
	var confMap = map[string]interface{}{
		"go.toolsEnvVars": map[string]string{
			"GOOS":   "js",
			"GOARCH": "wasm",
		},
		"go.BuildFlags": []string{
			"-tags=tinygo",
		},
	}

	var conf, err = json.MarshalIndent(confMap, "", "  ")
	if err != nil {
		panic(err)
	}

	makeDir(".vscode")

	err = os.WriteFile(".vscode/settings.json", conf, 0644)
	if err != nil {
		panic(err)
	}
}

// Create the build directory, write the files to it.
func initBuildFiles(projectName string) {
	makeDir(projectName + "/" + BuildDir)
	for _, file := range FilenamesBuild {
		createBuildFile(projectName+"/", file)
	}
}

// Create the static directory, write the files to it.
func initStaticFiles(projectName string) {
	makeDir(projectName + "/" + StaticDir)
	for _, file := range FileNamesStatic {
		createStaticFile(projectName+"/"+StaticDir+"/", file)
	}
}

// Create files for in the build folder.
func createBuildFile(dir, filename string) {
	var err error
	var fileContent []byte

	fileContent, err = PSBuildFolder.ReadFile("build/" + filename)
	if err != nil {
		panic(err)
	}

	err = os.WriteFile(dir+"/"+BuildDir+"/"+filename, fileContent, 0644)
	if err != nil {
		panic(err)
	}
}

// Create files for in the static folder.
func createStaticFile(dir, filename string) {
	var err error
	var fileContent []byte

	fileContent, err = FileFolder.ReadFile("files/" + StaticDir + "/" + filename)
	if err != nil {
		panic(err)
	}

	err = os.WriteFile(dir+filename, fileContent, 0644)
	if err != nil {
		panic(err)
	}
}

// Read a file from the embedded FileFolder file system.
func ReadFileFolder(filename string) []byte {
	var f []byte
	var err error
	println("Reading file: " + filename + FileFolderSuffix)
	f, err = FileFolder.ReadFile("files/" + filename + FileFolderSuffix)
	if err != nil {
		panic(err)
	}
	return f
}

// Run a system command.
func runCmd(name string, args ...string) {
	var cmd = exec.Command(name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	var err = cmd.Run()
	if err != nil {
		panic(err)
	}
}

// Init default files and directories, such as go.mod, index.html, server.go and the source directory.
func initDefault(projectName string) {
	var err = os.Mkdir(projectName, 0755)
	if err != nil {
		panic(err)
	}

	var modFile = ReadFileFolder(ModFilename)
	os.WriteFile(projectName+"/"+ModFilename, modFile, 0644)
	var indexFile = ReadFileFolder(IndexFilename)
	os.WriteFile(projectName+"/"+IndexFilename, indexFile, 0644)
	var serverFile = ReadFileFolder(ServerFilenme)
	os.WriteFile(projectName+"/"+ServerFilenme, serverFile, 0644)

	err = os.Mkdir(projectName+"/"+SourceDir, 0755)
	if err != nil {
		panic(err)
	}
}

// Initialize go.mod to get the latest version of the project.
// This only works for github repositories with tags in the following format:
//
//	vDIGITS.DIGITS.DIGITS
func initGoMod(projectName string) {
	os.Chdir(projectName)

	var github_tag_url = github_tag_url
	var req, _ = http.NewRequest("GET", github_tag_url, nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	var tagList = Tags{}
	err = json.NewDecoder(resp.Body).Decode(&tagList)
	if err != nil {
		panic(err)
	}
	tagList.Descending()

	// Get the latest version of jsext
	var latestTag = tagList[0].Name

	runCmd("go", "get", "github.com/Nigel2392/jsext@"+latestTag)
	runCmd("go", "mod", "tidy")
}

// Create a directory if it does not exist.
func makeDir(path string) {
	// Check if build folder exists, if not create it
	if _, err := os.Stat(path); os.IsNotExist(err) {
		err = os.Mkdir(path, 0755)
		if err != nil {
			panic(err)
		}
	}
}
