from fastapi import FastAPI, HTTPException
from pydantic import BaseModel
from typing import List
import logging
import os

from sentence_transformers import SentenceTransformer


class EmbedRequest(BaseModel):
    text: str


class EmbedResponse(BaseModel):
    vector: List[float]


app = FastAPI()

logger = logging.getLogger("embedding-service")
logging.basicConfig(level=logging.INFO)


MODEL_NAME = os.environ.get(
    "MODEL_NAME",
    "BAAI/bge-m3"
)

MODEL_DIMENSION = 1024


# ==============================
# Load BGE-M3 model
# ==============================

try:
    logger.info(f"[BGE-M3] Loading model: {MODEL_NAME}")

    model = SentenceTransformer(
        MODEL_NAME,
        device="cpu"
    )

    logger.info(f"[BGE-M3] Model loaded: {MODEL_NAME}")
    logger.info(
        f"[BGE-M3] Expected Dimension: {MODEL_DIMENSION}"
    )

except Exception as exc:
    logger.exception(
        "[BGE-M3] Failed to load model"
    )
    raise exc


# ==============================
# Embedding API
# ==============================

@app.post(
    "/embed",
    response_model=EmbedResponse
)
async def embed(request: EmbedRequest):

    if not request.text or not request.text.strip():
        raise HTTPException(
            status_code=400,
            detail="Text is required"
        )

    try:

        vector = generate_embedding(
            request.text
        )

        logger.info(
            f"[BGE-M3] Generated embedding dimension={len(vector)}"
        )

        return {
            "vector": vector
        }

    except Exception as exc:

        logger.exception(
            "[BGE-M3] Embedding failed"
        )

        raise HTTPException(
            status_code=502,
            detail=str(exc)
        )


# ==============================
# Generate vector
# ==============================

def generate_embedding(
    text: str
) -> List[float]:

    embeddings = model.encode(
        [text],
        normalize_embeddings=True,
        convert_to_numpy=True,
    )

    if embeddings is None:
        raise RuntimeError(
            "Embedding model returned no vector"
        )


    vector = embeddings[0]


    if len(vector) != MODEL_DIMENSION:

        raise RuntimeError(
            f"Unexpected embedding dimension: "
            f"expected {MODEL_DIMENSION}, "
            f"got {len(vector)}"
        )


    return vector.tolist()