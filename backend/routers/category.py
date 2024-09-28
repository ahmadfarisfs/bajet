from fastapi import APIRouter, HTTPException, Header
from pydantic import BaseModel
from typing import List
import sqlite3

router = APIRouter(prefix="/category", tags=["category"])

class Category(BaseModel):
    name: str
    icon: str

class CategoryResponse(BaseModel):
    id: int
    name: str
    icon: str

# API to view all categories for a user
@router.get("/", response_model=List[CategoryResponse])
def get_user_categories(user_id: int = Header(...)):
    cursor = conn.cursor()
    cursor.execute('''
        SELECT c.id, c.name, c.icon FROM categories c
        JOIN user_categories uc ON c.id = uc.category_id
        WHERE uc.user_id = ?
    ''', (user_id,))
    rows = cursor.fetchall()
    return [CategoryResponse(id=row[0], name=row[1], icon=row[2]) for row in rows]

# API to create a new category for a user
@router.post("/", response_model=CategoryResponse)
def create_user_category(category: Category, user_id: int = Header(...)):
    cursor = conn.cursor()
    try:
        cursor.execute('''
            INSERT INTO categories (name, icon) VALUES (?, ?)
        ''', (category.name, category.icon))
        conn.commit()
        category_id = cursor.lastrowid

        cursor.execute('''
            INSERT INTO user_categories (user_id, category_id) VALUES (?, ?)
        ''', (user_id, category_id))
        conn.commit()
    except sqlite3.Error as e:
        conn.rollback()
        raise HTTPException(status_code=500, detail="Database error: " + str(e))

    return {**category.dict(), "id": category_id}

# API to delete a category for a user
@router.delete("/{category_id}", response_model=dict)
def delete_user_category(category_id: int, user_id: int = Header(...)):
    cursor = conn.cursor()
    try:
        cursor.execute('''
            DELETE FROM user_categories WHERE user_id = ? AND category_id = ?
        ''', (user_id, category_id))
        if cursor.rowcount == 0:
            raise HTTPException(status_code=404, detail="Category not found for user")

        cursor.execute('''
            DELETE FROM categories WHERE id = ? AND NOT EXISTS (
                SELECT 1 FROM user_categories WHERE category_id = ?
            )
        ''', (category_id, category_id))
        conn.commit()
    except sqlite3.Error as e:
        conn.rollback()
        raise HTTPException(status_code=500, detail="Database error: " + str(e))

    return {"message": "Category deleted successfully"}