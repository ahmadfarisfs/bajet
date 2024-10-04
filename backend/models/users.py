from fastapi import APIRouter, HTTPException, Header, Form, Depends
from pydantic import BaseModel
from typing import List
import sqlite3
from typing import Annotated
from fastapi import FastAPI, Form
from backend.db.users import UserManager

router = APIRouter(prefix="/user", tags=["user"])

# Ensure the UserManager instance is created and connected in an async context
async def get_user_manager():
    user_manager = UserManager("this_isMysalt")
    await user_manager.connect()
    return user_manager

class User(BaseModel):
    first_name: str
    last_name: str
    email: str
    password: str

class UserResponse(BaseModel):
    id: int
    name: str
    email: str
    params: str

@router.post("/", response_model=UserResponse)
async def add_user(
    first_name: str = Form(...),
    last_name: str = Form(...),
    email: str = Form(...),
    password: str = Form(...),
    user_manager: UserManager = Depends(get_user_manager)
):
    try:
        new_user = await user_manager.add_user(
            first_name=first_name,
            last_name=last_name,
            email=email,
            password=password
        )
        return UserResponse(
            id=new_user.id,
            name=f"{new_user.first_name} {new_user.last_name}",
            email=new_user.email,
            params=new_user.params
        )
    except Exception as e:
        raise HTTPException(status_code=400, detail=str(e))