from pydantic import BaseModel
from typing import Optional

class TransactionClassifyRequest(BaseModel):
	merchant_name: Optional[str] = ""
	description: str
	amount: float
	direction: str # debit, credit

class CategoryResponse(BaseModel):
	category_id: str
	category_name: str
	confidence: float
