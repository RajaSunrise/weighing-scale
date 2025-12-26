#!/usr/bin/env python3
"""
Convert PyTorch model to ONNX format for OpenCV DNN

Usage:
    python convert_to_onnx.py models/platdetection.pt models/platdetection.onnx
"""

import torch
import sys
import argparse


def convert_to_onnx(pt_path, onnx_path, input_size=640):
    """
    Convert PyTorch model to ONNX format
    """
    print(f"Loading model from: {pt_path}")

    # Load model - try multiple approaches
    try:
        # Try loading as YOLOv5/YOLOv8 model
        from ultralytics import YOLO

        model = YOLO(pt_path)
        model_type = "YOLO (ultralytics)"
        export_model = model.model
    except ImportError:
        try:
            # Try loading as torchscript/regular PyTorch model
            model = torch.hub.load("ultralytics/yolov5", "custom", path=pt_path)
            model_type = "YOLOv5 (torchhub)"
            export_model = model
        except Exception as e:
            try:
                # Try loading directly as torch.load
                checkpoint = torch.load(pt_path, map_location="cpu")
                if isinstance(checkpoint, dict) and "model" in checkpoint:
                    model = checkpoint["model"]
                else:
                    model = checkpoint
                model_type = "Custom PyTorch model"
                export_model = model
                # Set to eval mode
                export_model.eval()
            except Exception as e2:
                print(f"Error loading model: {e2}")
                sys.exit(1)

    print(f"Model type detected: {model_type}")

    # Create dummy input
    dummy_input = torch.randn(1, 3, input_size, input_size)

    # Export to ONNX
    print(f"Converting to ONNX: {onnx_path}")
    print(f"Input size: {dummy_input.shape}")

    torch.onnx.export(
        export_model,
        dummy_input,
        onnx_path,
        export_params=True,
        opset_version=12,
        do_constant_folding=True,
        input_names=["images"],
        output_names=["output"],
        dynamic_axes={"images": {0: "batch_size"}, "output": {0: "batch_size"}},
    )

    print(f"Successfully converted to ONNX: {onnx_path}")

    # Verify ONNX model
    import onnx

    onnx_model = onnx.load(onnx_path)
    onnx.checker.check_model(onnx_model)
    print("ONNX model validation passed!")


def main():
    parser = argparse.ArgumentParser(description="Convert PyTorch model to ONNX format")
    parser.add_argument("input_model", help="Path to input .pt file")
    parser.add_argument("output_onnx", help="Path to output .onnx file")
    parser.add_argument(
        "--input-size", type=int, default=640, help="Input image size (default: 640)"
    )

    args = parser.parse_args()

    convert_to_onnx(args.input_model, args.output_onnx, args.input_size)


if __name__ == "__main__":
    main()
