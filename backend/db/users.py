import logging
from db.base import userDB  # Adjust the import path as necessary
import datetime
import bcrypt
import pytz
import os
from tzlocal import get_localzone

async def get_by_email(email: str):
    user = await userDB.find_one({"email": email})
    if user:
        return user
    else:
        raise ValueError("User not found")
    
async def update_user_password(email: str, new_password: str):
    result = await userDB.update_one(
        {"email": email},
        {"$set": {"password": new_password}}
    )
    if result.matched_count == 0:
        raise ValueError("User not found")
    logging.info(f"Updated password for user with email: {email}")
    
async def authenticate_user( email: str, password: str):
    user = await get_by_email(email)
    if user and bcrypt.checkpw(password.encode('utf-8'), user['password'].encode('utf-8')):
        return user
    else:
        raise ValueError("Invalid email or password")
    
async def add_user( first_name: str, last_name: str, email: str, password: str) :
    hashed_password = bcrypt.hashpw(password.encode('utf-8'), bcrypt.gensalt())
    password = hashed_password.decode('utf-8')
    timezone = os.getenv('TIMEZONE', 'Asia/Jakarta')
    local_tz = pytz.timezone(timezone)
    user = {
        "first_name": first_name,
        "last_name": last_name,
        "email": email,
        "password": password,
        "created_at": datetime.datetime.now(local_tz).isoformat(),
    }
    result = await userDB.insert_one(user)
    new_user = await userDB.find_one({"_id": result.inserted_id})
    logging.info(f"User added with id: {result.inserted_id}")
    logging.info(f"user added: {new_user}")
    return new_user