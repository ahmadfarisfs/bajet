import logging
import os
import datetime
import motor.motor_asyncio
from motor.motor_asyncio import AsyncIOMotorClient, AsyncIOMotorDatabase
import asyncio
import sys

# Configure the root logger
logging.basicConfig(
    level=logging.INFO,  # Set the logging level
    format='%(asctime)s - %(levelname)s - %(message)s',  # Format of the log message
    handlers=[logging.StreamHandler(sys.stdout)]  # Stream handler to output to the console
)

class MongoManager:
    client: AsyncIOMotorClient = None

    async def connect(self):
        if MongoManager.client is not None:
            logging.info(f"Using existing MongoDB client in {self.__class__.__name__}.")
            self.db = MongoManager.client.get_database()
            return

        mongodb_url = os.getenv("MONGODB_URL")
        if not mongodb_url:
            raise ValueError("MONGODB_URL environment variable not set")
        
        logging.info(f"Connecting to MongoDB in {self.__class__.__name__}.")
        MongoManager.client = AsyncIOMotorClient(mongodb_url)
        self.db = MongoManager.client.get_database()
        logging.info(f"Connected to MongoDB in {self.__class__.__name__}.")

    async def close(self):
        logging.info("Closing connection with MongoDB.")
        MongoManager.client.close()
        MongoManager.client = None
        logging.info("Closed connection with MongoDB.")



# mon = MongoManager()
# asyncio.run(mon.connect_to_database())
# asyncio.run(mon.add_user("koko","dsad","ppp2d2"))