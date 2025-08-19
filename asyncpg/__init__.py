class Pool:
    async def acquire(self):
        raise NotImplementedError
    async def release(self, conn):
        raise NotImplementedError
    async def close(self):
        pass

async def create_pool(*args, **kwargs):
    return Pool()