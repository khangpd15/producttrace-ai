from fastapi import FastAPI, HTTPException
from pydantic import BaseModel
from typing import List
import hashlib

class EmbedRequest(BaseModel):
    text: str

class EmbedResponse(BaseModel):
    vector: List[float]

app = FastAPI()

@app.post('/embed', response_model=EmbedResponse)
async def embed(request: EmbedRequest):
    if not request.text:
        raise HTTPException(status_code=400, detail='Text is required')

    try:
        vector = _text_to_embedding(request.text)
        return {'vector': vector}
    except Exception as exc:
        raise HTTPException(status_code=502, detail=str(exc))


def _text_to_embedding(text: str, dim: int = 768) -> List[float]:
    digest = hashlib.sha256(text.encode('utf-8')).digest()
    values = [b / 255.0 for b in digest]
    if dim <= len(values):
        return values[:dim]
    extra = [0.0] * (dim - len(values))
    return values + extra
