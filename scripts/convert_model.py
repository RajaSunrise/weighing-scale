#!/usr/bin/env python3
"""
Convert PyTorch model (.pt) to ONNX format for OpenCV DNN
Usage: python convert_to_onnx.py <input_pt_file> <output_onnx_file>
"""

import torch
import sys
import os

def main():
    if len(sys.argv) < 3:
        print("Usage: python convert_to_onnx.py <input_pt_file> <output_onnx_file>")
        sys.exit(1)

    input_file = sys.argv[1]
    output_file = sys.argv[2]

    if not os.path.exists(input_file):
        print(f"Error: Input file '{input_file}' not found")
        sys.exit(1)

    print(f"Loading model from {input_file}...")

    try:
        # Use simple default (latest opset 18)
        print("Attempting to load via torch.hub (yolov5)...")
        model = torch.hub.load('ultralytics/yolov5', 'custom', path=input_file, force_reload=True)
        model.eval()

        dummy_input = torch.randn(1, 3, 640, 640)
        print("Exporting to ONNX via Hub Model (Opset 18 - Default)...")

        # Omit opset_version to use default (latest)
        torch.onnx.export(model, dummy_input, output_file)
        print(f"Success! Saved to {output_file}")

    except Exception as e:
        print(f"Error: {e}")
        sys.exit(1)

if __name__ == "__main__":
    main()
