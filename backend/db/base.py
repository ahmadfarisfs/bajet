import logging
import os
import datetime
from typing import Optional
import motor.motor_asyncio
from motor.motor_asyncio import AsyncIOMotorClient, AsyncIOMotorDatabase
import asyncio
import sys


mongodb_url = os.getenv("MONGODB_URL")
if not mongodb_url:
    raise ValueError("MONGODB_URL environment variable not set")
db = AsyncIOMotorClient(mongodb_url)#.get_database()
userDB = db.users

# class MongoManager:
#     async def connect(self):
#         if MongoManager.client is not None:
#             logging.info(f"Using existing MongoDB client in {self.__class__.__name__}.")
#             self.db = MongoManager.client.get_database()
#             return

#         mongodb_url = os.getenv("MONGODB_URL")
#         if not mongodb_url:
#             raise ValueError("MONGODB_URL environment variable not set")
        
#         logging.info(f"Connecting to MongoDB in {self.__class__.__name__}.")
#         MongoManager.client = AsyncIOMotorClient(mongodb_url)
#         self.db = MongoManager.client.get_database()
#         logging.info(f"Connected to MongoDB in {self.__class__.__name__}.")

#     async def close(self):
#         logging.info("Closing connection with MongoDB.")
#         MongoManager.client.close()
#         logging.info("Closed connection with MongoDB.")
