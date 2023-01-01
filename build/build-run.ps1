$goos = $Env:GOOS; $goarch = $Env:GOARCH
$Env:GOOS = "js"; $Env:GOARCH = "wasm"
go build -ldflags="-s -w" -trimpath -o .\static\main.wasm .\src
# wasm-opt .\static\main.wasm -o=".\static\opt.wasm" -Oz --shrink-level=3 --optimize-level=3
$env:GOOS = $goos; $env:GOARCH = $goarch;
go run ./server.go
