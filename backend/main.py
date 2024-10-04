from fastapi import FastAPI
# import sqlite3
from backend.routers import user
import logging
app = FastAPI()
logging.info("Staring!")
# Include the routers
app.include_router(user.router)
# app.include_router(expense.router)