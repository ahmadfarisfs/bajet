from fastapi import APIRouter, HTTPException, Header
from pydantic import BaseModel
from typing import List
from datetime import date
import sqlite3

router = APIRouter(prefix="/expense", tags=["expense"])

class Expense(BaseModel):
    amount: float
    date: date
    name: str

class ExpenseResponse(BaseModel):
    id: int
    amount: float
    date: date
    name: str
    user_id: int

# API to add an expense
@router.post("/add", response_model=ExpenseResponse)
def add_expense(expense: Expense, user_id: int = Header(...)):
    cursor = conn.cursor()
    cursor.execute('''
        INSERT INTO expenses (amount, date, name, user_id) VALUES (?, ?, ?, ?)
    ''', (expense.amount, expense.date.isoformat(), expense.name, user_id))
    conn.commit()
    expense_id = cursor.lastrowid
    return {**expense.dict(), "id": expense_id, "user_id": user_id}

# API to get expenses for a specific day
@router.get("/get/{day}", response_model=List[ExpenseResponse])
def get_expenses(day: date, user_id: int = Header(...)):
    cursor = conn.cursor()
    cursor.execute('''
        SELECT id, amount, date, name, user_id FROM expenses WHERE date = ? AND user_id = ?
    ''', (day.isoformat(), user_id))
    rows = cursor.fetchall()
    return [ExpenseResponse(id=row[0], amount=row[1], date=row[2], name=row[3], user_id=row[4]) for row in rows]

# API to delete an expense by ID
@router.delete("/delete/{expense_id}", response_model=dict)
def delete_expense(expense_id: int, user_id: int = Header(...)):
    cursor = conn.cursor()
    cursor.execute('''
        DELETE FROM expenses WHERE id = ? AND user_id = ?
    ''', (expense_id, user_id))
    if cursor.rowcount == 0:
        raise HTTPException(status_code=404, detail="Expense not found")
    conn.commit()
    return {"message": "Expense deleted successfully"}