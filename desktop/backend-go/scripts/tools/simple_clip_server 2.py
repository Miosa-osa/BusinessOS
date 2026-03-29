#!/usr/bin/env python3
"""
Simple CLIP Server for BusinessOS
Provides image embedding generation via HTTP API
"""

import base64
import io
from fastapi import FastAPI, HTTPException
from fastapi.responses import JSONResponse
from pydantic import BaseModel
import uvicorn
from PIL import Image
import torch
from transformers import CLIPProcessor, CLIPModel

# Initialize FastAPI
app = FastAPI(title="Simple CLIP Server", version="1.0.0")

# Global model and processor
model = None
processor = None
device = "cuda" if torch.cuda.is_available() else "cpu"

class ImageEmbeddingRequest(BaseModel):
    image: str  # Base64 encoded image
    model: str = "openai/clip-vit-base-patch32"

class TextEmbeddingRequest(BaseModel):
    text: str
    model: str = "openai/clip-vit-base-patch32"

@app.on_event("startup")
async def load_model():
    """Load CLIP model on startup"""
    global model, processor
    print(f"Loading CLIP model on {device}...")
    model_name = "openai/clip-vit-base-patch32"

    try:
        processor = CLIPProcessor.from_pretrained(model_name)
        model = CLIPModel.from_pretrained(model_name).to(device)
        model.eval()  # Set to evaluation mode
        print(f"✅ CLIP model loaded successfully on {device}")
    except Exception as e:
        print(f"❌ Error loading model: {e}")
        raise

@app.get("/")
async def root():
    """Root endpoint"""
    return {
        "name": "Simple CLIP Server",
        "version": "1.0.0",
        "model": "openai/clip-vit-base-patch32",
        "device": device,
        "status": "ready" if model is not None else "loading"
    }

@app.get("/health")
async def health():
    """Health check endpoint"""
    return {
        "status": "healthy" if model is not None else "initializing",
        "model_loaded": model is not None,
        "device": device
    }

@app.post("/embed/image")
async def embed_image(request: ImageEmbeddingRequest):
    """Generate embedding for an image"""
    if model is None:
        raise HTTPException(status_code=503, detail="Model not loaded yet")

    try:
        # Decode base64 image
        image_data = base64.b64decode(request.image)
        image = Image.open(io.BytesIO(image_data)).convert("RGB")

        # Process image
        inputs = processor(images=image, return_tensors="pt").to(device)

        # Generate embedding
        with torch.no_grad():
            image_features = model.get_image_features(**inputs)
            # Normalize
            image_features = image_features / image_features.norm(dim=-1, keepdim=True)

        # Convert to list
        embedding = image_features.cpu().numpy()[0].tolist()

        return {
            "embedding": embedding,
            "model": request.model,
            "dimensions": len(embedding)
        }

    except Exception as e:
        raise HTTPException(status_code=500, detail=f"Error processing image: {str(e)}")

@app.post("/embed/text")
async def embed_text(request: TextEmbeddingRequest):
    """Generate embedding for text"""
    if model is None:
        raise HTTPException(status_code=503, detail="Model not loaded yet")

    try:
        # Process text
        inputs = processor(text=request.text, return_tensors="pt", padding=True).to(device)

        # Generate embedding
        with torch.no_grad():
            text_features = model.get_text_features(**inputs)
            # Normalize
            text_features = text_features / text_features.norm(dim=-1, keepdim=True)

        # Convert to list
        embedding = text_features.cpu().numpy()[0].tolist()

        return {
            "embedding": embedding,
            "model": request.model,
            "dimensions": len(embedding)
        }

    except Exception as e:
        raise HTTPException(status_code=500, detail=f"Error processing text: {str(e)}")

if __name__ == "__main__":
    print("=" * 60)
    print("Starting Simple CLIP Server for BusinessOS")
    print("=" * 60)
    print(f"Device: {device}")
    print("Endpoints:")
    print("  - GET  /           - Server info")
    print("  - GET  /health     - Health check")
    print("  - POST /embed/image - Generate image embedding")
    print("  - POST /embed/text  - Generate text embedding")
    print("=" * 60)

    uvicorn.run(app, host="localhost", port=8000, log_level="info")
