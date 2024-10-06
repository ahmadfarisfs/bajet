from fastapi import FastAPI
import uvicorn
from routers.user import router as UserRouter
import logging
import sys

# Configure the root logger
logging.basicConfig(
    level=logging.INFO,  # Set the logging level
    format='%(asctime)s - %(levelname)s - %(message)s',  
    handlers=[logging.StreamHandler(sys.stdout)]  
)

app = FastAPI()
app.include_router(UserRouter)

if __name__ == "__main__":
    uvicorn.run(app, host="0.0.0.0", port=8000)