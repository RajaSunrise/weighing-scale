#!/usr/bin/env python3
"""
Create a minimal valid ONNX file that won't cause OpenCV DNN errors
This serves as a placeholder when the real model is not yet converted
"""

import onnx
from onnx import helper, numpy_helper
from onnx import TensorProto
import numpy as np


def create_valid_placeholder_onnx(output_path):
    """Create a minimal valid ONNX model with proper nodes"""

    # Create inputs
    X = helper.make_tensor_value_info("images", TensorProto.FLOAT, [1, 3, 640, 640])

    # Create outputs
    Y = helper.make_tensor_value_info("output", TensorProto.FLOAT, [1, 25200, 85])

    # Create a simple Identity node to make the graph valid
    identity_node = helper.make_node("Identity", inputs=["images"], outputs=["output"])

    # Create graph with the node
    graph = helper.make_graph(
        [identity_node],  # nodes
        "placeholder_yolo_model",
        [X],  # inputs
        [Y],  # outputs
    )

    # Create model
    model = helper.make_model(graph, producer_name="placeholder")

    # Set opset version
    model.opset_import[0].version = 12

    # Save
    onnx.save(model, output_path)
    print(f"Created valid placeholder ONNX: {output_path}")

    # Verify
    try:
        onnx.checker.check_model(model)
        print("ONNX model validation passed!")
    except Exception as e:
        print(f"Warning: ONNX validation failed: {e}")


if __name__ == "__main__":
    create_valid_placeholder_onnx("models/platdetection.onnx")
