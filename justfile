[private]
default:
    @just --list --unsorted

# run the glox interpreter
run *args:
    @go run . {{ args }}

alias r := run

# build the project
build:
    @go build -o ./tmp/golox .

alias b := build

# run the ast generator
gen *args:
    @go run ./tool/generateAst.go {{ args }}
    @go fmt ./...

alias g := gen

# format the code
fmt:
    @go fmt .

alias f := fmt
