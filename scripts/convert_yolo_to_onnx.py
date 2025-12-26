#!/usr/bin/env python3
"""
Convert YOLO model to ONNX using ultralytics
Install: pip install ultralytics torch onnx
Usage: python convert_yolo_to_onnx.py <input_pt_file> <output_onnx_file>
"""

import sys
import os


def main():
    if len(sys.argv) < 2:
        print(
            "Usage: python convert_yolo_to_onnx.py <input_pt_file> [output_onnx_file]"
        )
        print("Example: python convert_yolo_to_onnx.py models/platdetection.pt")
        print("Output will be saved as models/platdetection.onnx by default")
        sys.exit(1)

    input_file = sys.argv[1]

    # Generate output filename
    if len(sys.argv) >= 3:
        output_file = sys.argv[2]
    else:
        output_file = input_file.replace(".pt", ".onnx")

    if not os.path.exists(input_file):
        print(f"Error: Input file '{input_file}' not found")
        sys.exit(1)

    try:
        from ultralytics import YOLO
    except ImportError:
        print("Error: ultralytics library not found")
        print("Install it with: pip install ultralytics")
        sys.exit(1)

    print(f"Loading YOLO model from {input_file}...")
    try:
        model = YOLO(input_file)
        print(f"Model loaded successfully!")
        print(f"Model info: {model.info()}")

        print(f"\nExporting to ONNX format...")
        model.export(format="onnx", imgsz=640, opset=12, simplify=True, dynamic=False)

        # ultralytics saves with .onnx extension in same directory
        output_onnx = input_file.replace(".pt", ".onnx")

        if os.path.exists(output_onnx):
            # Rename if custom output filename specified
            if output_file != output_onnx:
                os.rename(output_onnx, output_file)

            size_mb = os.path.getsize(output_file) / (1024 * 1024)
            print(f"\n✓ Successfully converted to ONNX: {output_file}")
            print(f"  File size: {size_mb:.2f} MB")
        else:
            print(f"\nError: ONNX file not generated at expected location")
            print(f"Expected: {output_onnx}")

    except Exception as e:
        print(f"\nError during conversion: {e}")
        import traceback

        traceback.print_exc()
        sys.exit(1)


if __name__ == "__main__":
    main()
