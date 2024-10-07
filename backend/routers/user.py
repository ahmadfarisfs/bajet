from fastapi import APIRouter, HTTPException, Header, Form, Depends
from pydantic import BaseModel
from datetime import datetime
from typing import List
import sqlite3
import logging
from typing import Annotated
from fastapi import FastAPI, Form
from db.users import add_user

router = APIRouter(prefix="/user", tags=["user"])

class CreateUserRequest(BaseModel):
    first_name: str
    last_name: str
    email: str
    password: str

class UserResponse(BaseModel):
    id: str
    first_name: str
    last_name: str
    email: str
    created_at: datetime

@router.post("/", response_model=UserResponse)
async def register_user(
    first_name: str = Form(...),
    last_name: str = Form(...),
    email: str = Form(...),
    password: str = Form(...)
):
    try:
        new_user = await add_user(
            first_name=first_name,
            last_name=last_name,
            email=email,
            password=password
        )
        if new_user is None:
            raise HTTPException(status_code=400, detail="Failed to create user")
        
        return UserResponse(
            id=str(new_user["_id"]),
            email=new_user["email"],
            first_name=new_user["first_name"],
            last_name=new_user["last_name"],
            created_at=new_user["created_at"]
        )
    except Exception as e:
        logging.error(e)
        raise HTTPException(status_code=400, detail=str(e))