set shell := ["pwsh", "-NoProfile", "-c"]

BINARY_NAME := "audateci.exe"
BUILD_DIR := "build"
OUT_PATH := BUILD_DIR + "\\" + BINARY_NAME

default:
    @just --list

build:
    @echo "Compiling..."
    go build -o {{OUT_PATH}}
    @echo "Done"

run +args='': build
    @echo "Running..."
    .\{{OUT_PATH}} {{args}}

clean:
    go clean
    rm -Force {{OUT_PATH}}
