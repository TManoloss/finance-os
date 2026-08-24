from pydantic import BaseModel, Field
from typing import Any, Dict, List, Optional


class ChatMessage(BaseModel):
    role: str  # user, assistant
    content: str


class ChatRequest(BaseModel):
    user_id: str
    message: str
    history: List[ChatMessage] = Field(default_factory=list)
    context: Optional[Dict[str, Any]] = None


class ChatResponse(BaseModel):
    response: str
