$goos = $Env:GOOS; $goarch = $Env:GOARCH
$Env:GOOS = "js"; $Env:GOARCH = "wasm";
tinygo build -o .\static\main.wasm -tags tinygo -target wasm .\src
$env:GOOS = $goos; $env:GOARCH = $goarch;
# go build -o server.exe ./server.go
go run ./server.go
