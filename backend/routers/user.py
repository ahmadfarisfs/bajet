from fastapi import APIRouter, HTTPException, Header
from pydantic import BaseModel
from typing import List
import sqlite3

router = APIRouter(prefix="/user", tags=["user"])

class User(BaseModel):
    name: str
    email: str
    password: str
    params: str

class UserResponse(BaseModel):
    id: int
    name: str
    email: str
    params: str

class UserLogin(BaseModel):
    email: str
    password: str

class UserLoginResponse(BaseModel):
    id: int
    name: str
    email: str

# API to add a user
@router.post("/add", response_model=UserResponse)
def add_user(user: User):
    cursor = conn.cursor()
    try:
        cursor.execute('''
            INSERT INTO users (name, email, password, params) VALUES (?, ?, ?, ?)
        ''', (user.name, user.email, user.password, user.params))
        conn.commit()
        user_id = cursor.lastrowid

        # Assign default categories to the new user
        cursor.execute('SELECT id FROM categories')
        category_ids = cursor.fetchall()
        user_categories = [(user_id, category_id[0]) for category_id in category_ids]
        cursor.executemany('''
            INSERT INTO user_categories (user_id, category_id) VALUES (?, ?)
        ''', user_categories)
        conn.commit()
    except sqlite3.Error as e:
        conn.rollback()
        raise HTTPException(status_code=500, detail="Database error: " + str(e))

    return {**user.dict(), "id": user_id}

# API to get users with cursor pagination
@router.get("/get", response_model=List[UserResponse])
def get_users(limit: int = 10, offset: int = 0):
    cursor = conn.cursor()
    cursor.execute('''
        SELECT id, name, email, params FROM users LIMIT ? OFFSET ?
    ''', (limit, offset))
    rows = cursor.fetchall()
    return [UserResponse(id=row[0], name=row[1], email=row[2], params=row[3]) for row in rows]

# API to login a user
@router.post("/login", response_model=UserLoginResponse)
def login_user(user: UserLogin):
    cursor = conn.cursor()
    cursor.execute('''
        SELECT id, name, email FROM users WHERE email = ? AND password = ?
    ''', (user.email, user.password))
    row = cursor.fetchone()
    if row is None:
        raise HTTPException(status_code=401, detail="Invalid email or password")
    return UserLoginResponse(id=row[0], name=row[1], email=row[2])