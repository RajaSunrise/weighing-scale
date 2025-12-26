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
        print(
            "Example: python convert_to_onnx.py models/platdetection.pt models/platdetection.onnx"
        )
        sys.exit(1)

    input_file = sys.argv[1]
    output_file = sys.argv[2]

    if not os.path.exists(input_file):
        print(f"Error: Input file '{input_file}' not found")
        sys.exit(1)

    print(f"Loading model from {input_file}...")

    try:
        # Load the model
        model = torch.load(input_file, map_location="cpu")

        # Handle different model formats
        if isinstance(model, dict):
            # Check if it's a YOLO model with 'model' key
            if "model" in model:
                model = model["model"]
            # Check if it has 'state_dict' key
            elif "state_dict" in model:
                model = model["state_dict"]
                # You might need to rebuild the model architecture here
                print(
                    "Warning: Model contains state_dict. You may need to specify the model architecture."
                )

        # Set to evaluation mode
        model.eval()

        # Create dummy input (adjust size according to your model's expected input)
        # YOLO models typically use 640x640
        dummy_input = torch.randn(1, 3, 640, 640)

        print("Exporting to ONNX...")
        torch.onnx.export(
            model,
            dummy_input,
            output_file,
            export_params=True,
            opset_version=12,
            do_constant_folding=True,
            input_names=["images"],
            output_names=["output"],
            dynamic_axes={"images": {0: "batch_size"}, "output": {0: "batch_size"}},
        )

        print(f"Successfully converted to {output_file}")
        print(f"File size: {os.path.getsize(output_file) / (1024 * 1024):.2f} MB")

    except Exception as e:
        print(f"Error during conversion: {e}")
        print("\nIf you're using YOLO models, try this instead:")
        print("  from ultralytics import YOLO")
        print("  model = YOLO('" + input_file + "')")
        print("  model.export(format='onnx')")
        sys.exit(1)


if __name__ == "__main__":
    main()
