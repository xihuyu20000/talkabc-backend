#!/bin/bash

echo "================================================"
echo "   TalkABC - Swagger Export Script"
echo "================================================"
echo ""

CHECK_ONLY=0
VERBOSE=0
OUTPUT_DIR="./swagger"

while [[ $# -gt 0 ]]; do
    case "$1" in
        --check)
            CHECK_ONLY=1
            shift
            ;;
        --verbose)
            VERBOSE=1
            shift
            ;;
        --output)
            OUTPUT_DIR="$2"
            shift 2
            ;;
        *)
            shift
            ;;
    esac
done

echo "Checking swag tool..."
if ! command -v swag &> /dev/null; then
    echo "swag not found, installing..."
    go install github.com/swaggo/swag/cmd/swag@latest
    if [[ $? -ne 0 ]]; then
        echo ""
        echo "================================================"
        echo "   ERROR: Failed to install swag!"
        echo "================================================"
        exit 1
    fi
    echo "swag installed successfully."
else
    echo "swag is already installed."
fi

if [[ $CHECK_ONLY -eq 1 ]]; then
    echo ""
    echo "================================================"
    echo "   Check completed: swag is available"
    echo "================================================"
    exit 0
fi

echo ""
echo "Generating Swagger documentation..."
echo "Output directory: $OUTPUT_DIR"
echo ""

swag init --dir ./cmd/server,./internal/handler --output "$OUTPUT_DIR"
if [[ $? -ne 0 ]]; then
    echo ""
    echo "================================================"
    echo "   ERROR: Failed to generate Swagger docs!"
    echo "================================================"
    exit 1
fi

echo "Swagger documentation generated successfully!"
echo ""

if [[ $VERBOSE -eq 1 ]]; then
    echo "Generated files:"
    echo "   - $OUTPUT_DIR/swagger.json"
    echo "   - $OUTPUT_DIR/swagger.yaml"
    echo "   - $OUTPUT_DIR/docs.go"
fi

echo ""
echo "================================================"
echo "   Export completed!"
echo ""
echo "   Swagger UI: http://localhost:8080/swagger/index.html"
echo "   Swagger JSON: $OUTPUT_DIR/swagger.json"
echo "   Apifox Import: $OUTPUT_DIR/swagger.json"
echo "================================================"