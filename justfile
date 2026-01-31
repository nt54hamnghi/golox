[private]
default:
    @just --list --unsorted

# run the glox interpreter
run *args:
    @go run . {{ args }}

alias r := run

# format the code
fmt:
    @go fmt .
