#!/bin/bash
# Quick script to convert model and update the application

echo "Converting PyTorch model to ONNX..."
python3 scripts/convert_yolo_to_onnx.py models/platdetection.pt

if [ $? -eq 0 ]; then
    echo ""
    echo "Conversion successful! Now updating the application..."
    
    # Check if onnx file exists
    if [ -f "models/platdetection.onnx" ]; then
        # Update main.go to use .onnx file
        if grep -q "platdetection.pt" cmd/server/main.go; then
            echo "Updating cmd/server/main.go to use .onnx file..."
            sed -i 's/platdetection\.pt/platdetection.onnx/g' cmd/server/main.go
            echo "✓ Updated main.go"
        fi
        
        echo ""
        echo "Done! Rebuild the application with:"
        echo "  docker-compose build"
        echo "  docker-compose up"
    else
        echo "Error: platdetection.onnx not found"
        exit 1
    fi
else
    echo "Error: Conversion failed. Creating placeholder ONNX file..."
    python3 scripts/create_placeholder_onnx.py
    
    if [ -f "models/platdetection.onnx" ]; then
        echo "✓ Placeholder created. Application will run with ANPR disabled."
        echo ""
        echo "Note: To enable ANPR, convert your PyTorch model:"
        echo "  pip3 install torch ultralytics onnx"
        echo "  python3 scripts/convert_yolo_to_onnx.py models/platdetection.pt"
    else
        echo "Error: Failed to create placeholder"
        exit 1
    fi
fi
