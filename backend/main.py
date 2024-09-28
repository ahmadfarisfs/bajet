from fastapi import FastAPI
import sqlite3
from routers import user, expense

app = FastAPI()

# Global connection object
conn = None

# Database setup
def init_db():
    global conn
    conn = sqlite3.connect('expenses.db')
    cursor = conn.cursor()
    cursor.execute('''
        CREATE TABLE IF NOT EXISTS users (
            id INTEGER PRIMARY KEY,
            name TEXT,
            email TEXT UNIQUE,
            password TEXT,
            params TEXT
        )
    ''')
    cursor.execute('''
        CREATE TABLE IF NOT EXISTS expenses (
            id INTEGER PRIMARY KEY,
            amount REAL,
            date TEXT,
            name TEXT,
            user_id INTEGER,
            FOREIGN KEY (user_id) REFERENCES users (id)
        )
    ''')
    cursor.execute('''
        CREATE TABLE IF NOT EXISTS categories (
            id INTEGER PRIMARY KEY,
            name TEXT,
            icon TEXT
        )
    ''')
    cursor.execute('''
        CREATE TABLE IF NOT EXISTS user_categories (
            user_id INTEGER,
            category_id INTEGER,
            FOREIGN KEY (user_id) REFERENCES users (id),
            FOREIGN KEY (category_id) REFERENCES categories (id)
        )
    ''')
    # Prepopulate the categories table
    categories = [
        ("Groceries", "🛒"),
        ("Rent/Mortgage", "🏡"),
        ("Transportation", "🚕"),
        ("Utilities", "🔌"),
        ("Dining Out", "🍽️"),
        ("Entertainment", "🎬"),
        ("Healthcare", "💊"),
        ("Shopping", "🛍️"),
        ("Travel", "🚄"),
        ("Savings/Investments", "💰")
    ]
    cursor.executemany('''
        INSERT INTO categories (name, icon) VALUES (?, ?)
    ''', categories)
    conn.commit()

# Initialize the database when the application starts
@app.on_event("startup")
def startup_event():
    init_db()

# Include the routers
app.include_router(user.router)
app.include_router(expense.router)