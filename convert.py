#!/usr/bin/env python3
"""
Simple converter - converts models/platdetection.pt to models/platdetection.onnx
Output will automatically be saved in models/ folder
"""

import sys
import os


def main():
    input_file = "models/platdetection.pt"
    output_file = "models/platdetection.onnx"

    if not os.path.exists(input_file):
        print(f"Error: {input_file} not found")
        sys.exit(1)

    print(f"Converting {input_file} to {output_file}")
    print("This may take a few minutes...")

    try:
        from ultralytics import YOLO

        # Load the YOLO model
        model = YOLO(input_file)

        # Export to ONNX format
        # Output will be saved as models/platdetection.onnx
        model.export(format="onnx", imgsz=640, opset=12, simplify=True)

        if os.path.exists(output_file):
            size_mb = os.path.getsize(output_file) / (1024 * 1024)
            print(f"\n✓ Conversion successful!")
            print(f"  Output: {output_file}")
            print(f"  Size: {size_mb:.2f} MB")
        else:
            print(f"\nError: Output file not created")
            sys.exit(1)

    except ImportError:
        print("\nError: ultralytics library not installed")
        print("Run: pip install ultralytics torch onnx")
        sys.exit(1)
    except Exception as e:
        print(f"\nError: {e}")
        import traceback

        traceback.print_exc()
        sys.exit(1)


if __name__ == "__main__":
    main()
