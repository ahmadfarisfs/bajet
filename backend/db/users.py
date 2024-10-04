import logging
from backend.db.base import MongoManager  # Adjust the import path as necessary
import datetime
import bcrypt

class UserManager(MongoManager):
    _instance = None

    def __new__(cls, *args, **kwargs):
        if cls._instance is None:
            cls._instance = super(UserManager, cls).__new__(cls)
        return cls._instance

    def __init__(self, salt_password: str):
        if not hasattr(self, 'initialized'):  # Ensure __init__ is only called once
            self.salt_password = salt_password
            self.initialized = True

    async def get_user_by_email(self, email: str):
        user = await self.db.users.find_one({"email": email})
        if user:
            return user
        else:
            raise ValueError("User not found")

    async def update_user_password(self, email: str, new_password: str):
        result = await self.db.users.update_one(
            {"email": email},
            {"$set": {"password": new_password}}
        )
        if result.matched_count == 0:
            raise ValueError("User not found")
        logging.info(f"Updated password for user with email: {email}")
        
    async def authenticate_user(self, email: str, password: str):
        user = await self.get_user_by_email(email)
        if user and bcrypt.checkpw(password.encode('utf-8'), user['password'].encode('utf-8')):
            return user
        else:
            raise ValueError("Invalid email or password")

    async def add_user(self, first_name: str, last_name: str, email: str, password: str):
        hashed_password = bcrypt.hashpw(password.encode('utf-8'), bcrypt.gensalt())
        password = hashed_password.decode('utf-8')
        user = {
            "first_name": first_name,
            "last_name": last_name,
            "email": email,
            "password": password,
            "created_at": datetime.datetime.now(),
        }
        result = await self.db.users.insert_one(user)
        logging.info(f"User added with id: {result.inserted_id}")