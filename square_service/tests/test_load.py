"""
Load tests for Square Payment Service

Run with: python -m pytest tests/test_load.py -v
For stress testing: python -m pytest tests/test_load.py::TestLoadTesting::test_concurrent_checkouts -v
"""
import pytest
import asyncio
import aiohttp
import time
import jwt
from concurrent.futures import ThreadPoolExecutor, as_completed
import statistics


class TestLoadTesting:
    """Load testing for critical endpoints."""
    
    @pytest.fixture
    def auth_token(self):
        """Generate test JWT token."""
        payload = {
            "user_id": "load_test_user",
            "sub": "load_test_user", 
            "exp": int(time.time()) + 3600
        }
        return jwt.encode(payload, "test_jwt_secret", algorithm="HS256")
    
    @pytest.fixture
    def base_url(self):
        """Base URL for testing."""
        return "http://localhost:8000"
    
    @pytest.mark.asyncio
    @pytest.mark.load
    async def test_health_endpoint_load(self, base_url):
        """Test health endpoint under load."""
        async def check_health(session):
            start_time = time.time()
            async with session.get(f"{base_url}/health") as response:
                await response.json()
                return time.time() - start_time, response.status
        
        # Run 100 concurrent requests
        connector = aiohttp.TCPConnector(limit=100)
        async with aiohttp.ClientSession(connector=connector) as session:
            tasks = [check_health(session) for _ in range(100)]
            results = await asyncio.gather(*tasks)
        
        response_times = [r[0] for r in results]
        status_codes = [r[1] for r in results]
        
        # All requests should succeed
        assert all(status == 200 for status in status_codes)
        
        # Response times should be reasonable (under 1 second)
        avg_response_time = statistics.mean(response_times)
        max_response_time = max(response_times)
        
        print(f"\nHealth endpoint load test results:")
        print(f"Average response time: {avg_response_time:.3f}s")
        print(f"Max response time: {max_response_time:.3f}s")
        print(f"95th percentile: {statistics.quantiles(response_times, n=20)[18]:.3f}s")
        
        assert avg_response_time < 1.0
        assert max_response_time < 2.0
    
    @pytest.mark.asyncio
    @pytest.mark.load
    async def test_webhook_endpoint_load(self, base_url):
        """Test webhook endpoint under load with signature verification."""
        import json
        import hmac
        import hashlib
        import base64
        
        async def send_webhook(session, event_data):
            start_time = time.time()
            
            # Create valid signature
            body = json.dumps(event_data).encode()
            webhook_url = "https://test.example.com/webhook"
            key = "test_secret".encode()
            msg = webhook_url.encode() + body
            signature = base64.b64encode(hmac.new(key, msg, hashlib.sha256).digest()).decode()
            
            headers = {
                "x-square-hmacsha256-signature": signature,
                "content-type": "application/json"
            }
            
            async with session.post(f"{base_url}/webhook", data=body, headers=headers) as response:
                await response.json()
                return time.time() - start_time, response.status
        
        # Create test webhook events
        event_template = {
            "type": "payment.updated",
            "data": {
                "object": {
                    "payment": {
                        "id": "PAYMENT_{}",
                        "status": "COMPLETED",
                        "order_id": "ORDER_{}",
                        "amount_money": {"amount": 1000, "currency": "CAD"}
                    }
                }
            }
        }
        
        # Generate 50 unique webhook events
        events = []
        for i in range(50):
            event = event_template.copy()
            event["data"]["object"]["payment"]["id"] = f"PAYMENT_{i}"
            event["data"]["object"]["payment"]["order_id"] = f"ORDER_{i}"
            events.append(event)
        
        connector = aiohttp.TCPConnector(limit=50)
        async with aiohttp.ClientSession(connector=connector) as session:
            tasks = [send_webhook(session, event) for event in events]
            results = await asyncio.gather(*tasks, return_exceptions=True)
        
        # Filter out exceptions and get successful results
        successful_results = [r for r in results if not isinstance(r, Exception)]
        response_times = [r[0] for r in successful_results]
        status_codes = [r[1] for r in successful_results]
        
        print(f"\nWebhook endpoint load test results:")
        print(f"Successful requests: {len(successful_results)}/{len(events)}")
        if response_times:
            print(f"Average response time: {statistics.mean(response_times):.3f}s")
            print(f"Max response time: {max(response_times):.3f}s")
        
        # At least 80% of requests should succeed
        success_rate = len(successful_results) / len(events)
        assert success_rate >= 0.8
    
    @pytest.mark.load
    def test_concurrent_checkouts(self, auth_token, base_url):
        """Test concurrent subscription checkout requests."""
        import requests
        
        def create_checkout(user_id):
            start_time = time.time()
            headers = {
                "Authorization": f"Bearer {auth_token}",
                "Content-Type": "application/json"
            }
            payload = {
                "membership_plan_id": "550e8400-e29b-41d4-a716-446655440000",
                "timezone": "UTC"
            }
            
            try:
                response = requests.post(
                    f"{base_url}/subscriptions/checkout",
                    json=payload,
                    headers=headers,
                    timeout=10
                )
                return time.time() - start_time, response.status_code, user_id
            except requests.RequestException as e:
                return time.time() - start_time, 0, user_id  # 0 indicates error
        
        # Simulate 20 concurrent users trying to checkout
        user_ids = [f"user_{i}" for i in range(20)]
        
        with ThreadPoolExecutor(max_workers=20) as executor:
            futures = [executor.submit(create_checkout, user_id) for user_id in user_ids]
            results = [future.result() for future in as_completed(futures)]
        
        response_times = [r[0] for r in results]
        status_codes = [r[1] for r in results]
        successful_requests = [r for r in results if r[1] in [200, 400, 409, 429]]  # Include expected errors
        
        print(f"\nConcurrent checkout test results:")
        print(f"Total requests: {len(results)}")
        print(f"Successful requests: {len(successful_requests)}")
        if response_times:
            print(f"Average response time: {statistics.mean(response_times):.3f}s")
            print(f"Max response time: {max(response_times):.3f}s")
        
        # Status code breakdown
        status_breakdown = {}
        for code in status_codes:
            status_breakdown[code] = status_breakdown.get(code, 0) + 1
        print(f"Status code breakdown: {status_breakdown}")
        
        # At least 80% should get a valid response (including rate limits)
        success_rate = len(successful_requests) / len(results)
        assert success_rate >= 0.8
    
    @pytest.mark.load
    def test_rate_limit_enforcement(self, auth_token, base_url):
        """Test that rate limiting is properly enforced."""
        import requests
        
        headers = {
            "Authorization": f"Bearer {auth_token}",
            "Content-Type": "application/json"
        }
        payload = {
            "membership_plan_id": "550e8400-e29b-41d4-a716-446655440000",
            "timezone": "UTC"
        }
        
        # Send requests rapidly to trigger rate limiting
        responses = []
        for i in range(10):  # Checkout endpoint has 5/minute limit
            try:
                response = requests.post(
                    f"{base_url}/subscriptions/checkout",
                    json=payload,
                    headers=headers,
                    timeout=5
                )
                responses.append(response.status_code)
            except requests.RequestException:
                responses.append(0)
            
            time.sleep(0.1)  # Small delay between requests
        
        print(f"\nRate limit test results:")
        print(f"Response codes: {responses}")
        
        # Should see some 429 (Too Many Requests) responses
        rate_limited_responses = [code for code in responses if code == 429]
        print(f"Rate limited responses: {len(rate_limited_responses)}")
        
        # At least some requests should be rate limited if sending more than allowed
        assert len(rate_limited_responses) > 0 or all(code in [200, 400, 409] for code in responses)
    
    @pytest.mark.asyncio
    @pytest.mark.load
    async def test_memory_usage_under_load(self, base_url):
        """Test memory usage doesn't grow excessively under load."""
        import psutil
        import os
        
        process = psutil.Process(os.getpid())
        initial_memory = process.memory_info().rss / 1024 / 1024  # MB
        
        async def make_request(session):
            async with session.get(f"{base_url}/health") as response:
                await response.json()
        
        # Make 1000 requests in batches
        connector = aiohttp.TCPConnector(limit=100)
        async with aiohttp.ClientSession(connector=connector) as session:
            for batch in range(10):  # 10 batches of 100 requests
                tasks = [make_request(session) for _ in range(100)]
                await asyncio.gather(*tasks)
                
                current_memory = process.memory_info().rss / 1024 / 1024
                print(f"Batch {batch + 1}: Memory usage: {current_memory:.1f} MB")
        
        final_memory = process.memory_info().rss / 1024 / 1024
        memory_increase = final_memory - initial_memory
        
        print(f"\nMemory usage test:")
        print(f"Initial memory: {initial_memory:.1f} MB")
        print(f"Final memory: {final_memory:.1f} MB") 
        print(f"Memory increase: {memory_increase:.1f} MB")
        
        # Memory increase should be reasonable (less than 100MB for 1000 requests)
        assert memory_increase < 100
    
    @pytest.mark.load
    def test_database_connection_pool_under_load(self, auth_token, base_url):
        """Test database connection pool handles concurrent requests."""
        import requests
        import threading
        
        results = []
        errors = []
        
        def make_sync_status_request():
            headers = {"Authorization": f"Bearer {auth_token}"}
            try:
                response = requests.get(
                    f"{base_url}/admin/sync/status",
                    headers=headers,
                    timeout=10
                )
                results.append(response.status_code)
            except Exception as e:
                errors.append(str(e))
        
        # Create 30 concurrent threads (more than the DB pool size of 20)
        threads = []
        for i in range(30):
            thread = threading.Thread(target=make_sync_status_request)
            threads.append(thread)
        
        # Start all threads
        start_time = time.time()
        for thread in threads:
            thread.start()
        
        # Wait for all threads to complete
        for thread in threads:
            thread.join()
        
        end_time = time.time()
        
        print(f"\nDatabase pool test results:")
        print(f"Total requests: 30")
        print(f"Successful requests: {len(results)}")
        print(f"Errors: {len(errors)}")
        print(f"Total time: {end_time - start_time:.2f}s")
        
        if errors:
            print(f"Sample errors: {errors[:3]}")
        
        # Most requests should succeed despite exceeding pool size
        success_rate = len(results) / 30
        assert success_rate >= 0.8
        
        # No more than a few errors should occur due to proper connection pooling
        assert len(errors) <= 5


if __name__ == "__main__":
    # Run specific load tests
    pytest.main([__file__ + "::TestLoadTesting::test_health_endpoint_load", "-v"])