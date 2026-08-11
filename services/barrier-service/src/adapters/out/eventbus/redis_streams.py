import asyncio
import json
import redis.asyncio as redis
from typing import Callable, Awaitable, Optional

from src.config import settings


class RedisEventBus:
    def __init__(self):
        self.client: Optional[redis.Redis] = None

    async def connect(self) -> None:
        self.client = redis.from_url(
            settings.redis_url,
            encoding="utf-8",
            decode_responses=True,
        )

    async def disconnect(self) -> None:
        if self.client:
            await self.client.close()

    async def publish(self, stream: str, event: dict) -> None:
        if not self.client:
            await self.connect()
        await self.client.xadd(stream, {"data": json.dumps(event)})

    async def subscribe(
        self,
        stream: str,
        group: str,
        consumer: str,
        handler: Callable[[dict], Awaitable[None]],
    ) -> None:
        if not self.client:
            await self.connect()

        try:
            await self.client.xgroup_create(stream, group, id="0", mkstream=True)
        except redis.ResponseError as e:
            if "BUSYGROUP" not in str(e):
                raise

        while True:
            try:
                response = await self.client.xreadgroup(
                    group,
                    consumer,
                    {stream: ">"},
                    count=10,
                    block=5000,
                )
            except redis.ResponseError:
                await asyncio.sleep(1)
                continue

            if not response:
                continue

            for _, messages in response:
                for msg_id, fields in messages:
                    try:
                        data = json.loads(fields["data"])
                        await handler(data)
                        await self.client.xack(stream, group, msg_id)
                    except Exception as e:
                        print(f"Error processing message: {e}")