[private]
default:
    @just --list --unsorted

# run the glox interpreter
run *args:
    @go run . {{ args }}

alias r := run

# run the ast generator
gen *args:
    @go run ./tool/generateAst.go {{ args }}

alias g := gen

# format the code
fmt:
    @go fmt .

alias f := fmt
