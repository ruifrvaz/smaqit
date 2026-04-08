#!/usr/bin/env python3
"""
E2E Test Runner using Copilot SDK

Invokes smaqit.ci-testing agent programmatically for CI/CD automation.
"""

import asyncio
import sys
import time
import logging
from pathlib import Path
from copilot import CopilotClient

# Configure logging
logging.basicConfig(
    level=logging.DEBUG,
    format='%(asctime)s [%(levelname)s] %(message)s',
    datefmt='%H:%M:%S'
)
logger = logging.getLogger(__name__)


async def show_progress(task, description="Working", check_interval=30):
    """Show progress updates while task runs with health checks"""
    start = time.time()
    dots = 0
    last_check = start
    
    while not task.done():
        elapsed = int(time.time() - start)
        mins, secs = divmod(elapsed, 60)
        dot_str = "." * (dots % 4)
        print(f"\r⏱  {description}{dot_str:<3} [{mins:02d}:{secs:02d}]", end="", flush=True)
        
        # Log periodic heartbeat
        if time.time() - last_check >= check_interval:
            logger.info(f"Heartbeat: Still waiting after {mins}m {secs}s")
            last_check = time.time()
        
        dots += 1
        await asyncio.sleep(5)  # Check more frequently for responsiveness
    
    print()  # New line after completion
    elapsed = int(time.time() - start)
    mins, secs = divmod(elapsed, 60)
    logger.info(f"Task completed after {mins}m {secs}s")


async def run_smoke_test() -> bool:
    """Quick smoke test to verify SDK connectivity"""
    logger.info("Running smoke test: 2+2 calculation")
    client = None
    try:
        client = CopilotClient()
        logger.debug("Starting Copilot client...")
        await client.start()
        logger.debug("Client started successfully")
        
        session = await client.create_session()
        logger.debug(f"Session created: {session}")
        
        logger.info("Sending test prompt: '2+2' (30s timeout)")
        response = await session.send_and_wait({"prompt": "2+2"}, timeout=30)
        logger.debug(f"Response type: {type(response)}")
        logger.debug(f"Response data: {response.data if hasattr(response, 'data') else 'No data attr'}")
        
        if hasattr(response.data, 'content'):
            logger.info(f"Smoke test response: {response.data.content[:100]}")
            return True
        else:
            logger.error("No content in smoke test response")
            return False
            
    except asyncio.TimeoutError as e:
        logger.error("❌ Smoke test timed out after 30s")
        logger.error("This usually means:")
        logger.error("  1. Copilot CLI is not authenticated properly")
        logger.error("  2. GITHUB_TOKEN doesn't have Copilot subscription access")
        logger.error("  3. CLI is not responding to requests")
        logger.error("")
        logger.error("In CI/CD environments, you may need:")
        logger.error("  - A Personal Access Token with Copilot access")
        logger.error("  - Store it as a repository secret (e.g., COPILOT_TOKEN)")
        logger.error("  - Pass it as GITHUB_TOKEN or COPILOT_GITHUB_TOKEN")
        logger.error("")
        logger.error("The default GitHub Actions GITHUB_TOKEN may not have Copilot access.")
        return False
        
    except Exception as e:
        logger.error(f"Smoke test failed: {type(e).__name__}: {e}")
        import traceback
        traceback.print_exc()
        return False
    finally:
        if client:
            await client.stop()


async def run_e2e_test(test_case: str = "mario-hello.automated", smoke_test_first: bool = True) -> bool:
    """
    Run end-to-end test using smaqit.ci-testing agent.
    
    Args:
        test_case: Test case name from docs/test-cases/ (e.g., "mario-hello.automated")
        smoke_test_first: Run smoke test before actual test
        
    Returns:
        True if test passed, False otherwise
    """
    print(f"🚀 Starting E2E test: {test_case}")
    print("=" * 70)
    
    # Optional smoke test
    if smoke_test_first:
        logger.info("Running pre-flight smoke test...")
        if not await run_smoke_test():
            logger.error("❌ Smoke test failed - SDK not working properly")
            return False
        logger.info("✅ Smoke test passed\n")
    
    client = None
    try:
        # Initialize Copilot client
        logger.info("Initializing Copilot SDK...")
        client = CopilotClient()
        logger.debug(f"Client instance: {client}")
        
        logger.debug("Starting client connection...")
        await client.start()
        logger.info("✓ Connected to Copilot CLI\n")
        
        # Create session
        logger.info("Creating session...")
        session = await client.create_session()
        logger.debug(f"Session instance: {session}")
        logger.info("✓ Session created\n")
        
        # Invoke CI testing agent
        logger.info(f"Invoking /smaqit.ci-testing with test case: {test_case}")
        print("-" * 70)
        
        prompt = f"/smaqit.ci-testing run test case {test_case} from docs/test-cases/"
        logger.debug(f"Prompt: {prompt}")
        
        # Run with progress indicator
        logger.info("Sending prompt to agent (25min timeout)...")
        task = asyncio.create_task(
            session.send_and_wait(
                {"prompt": prompt},
                timeout=1500  # 25 minutes for full E2E test (CI environment is slower than local)
            )
        )
        await show_progress(task, "Agent executing test workflow", check_interval=60)
        
        logger.info("Task completed, retrieving response...")
        response = await task
        logger.debug(f"Response type: {type(response)}")
        
        if not response:
            logger.error("❌ No response from testing agent (None)")
            return False
        
        logger.debug(f"Response attributes: {dir(response)}")
        
        if not hasattr(response, 'data'):
            logger.error("❌ Response has no 'data' attribute")
            return False
            
        logger.debug(f"Response.data type: {type(response.data)}")
        logger.debug(f"Response.data attributes: {dir(response.data)}")
        
        if not hasattr(response.data, 'content'):
            logger.error("❌ Response.data has no 'content' attribute")
            return False
        
        content = response.data.content
        logger.info(f"Response content length: {len(content)} chars")
        logger.debug(f"Response content preview: {content[:200]}...")
        
        print(f"\n{content}\n")
        print("-" * 70)
        
        # Check for success indicators in response
        passed = any(indicator in content.lower() for indicator in [
            "✓ pass",
            "test passed",
            "all tests passed",
            "success"
        ])
        
        if passed:
            logger.info("✅ E2E test PASSED")
            print("✅ E2E test PASSED")
            return True
        else:
            logger.warning("❌ E2E test FAILED or UNCLEAR")
            print("❌ E2E test FAILED or UNCLEAR")
            return False
            
    except asyncio.TimeoutError as e:
        logger.error(f"\n❌ Test timed out after {1500}s (25 minutes)")
        logger.error("This suggests the agent is not responding or is stuck")
        print(f"\n❌ TIMEOUT: Test exceeded 25 minute limit")
        print("   Possible causes:")
        print("   - Agent not responding")
        print("   - CLI process hung")
        print("   - Network connectivity issues")
        return False
        
    except Exception as e:
        logger.error(f"\n❌ Error during test execution:")
        logger.error(f"   {type(e).__name__}: {e}")
        print(f"\n❌ Error during test execution:")
        print(f"   {type(e).__name__}: {e}")
        import traceback
        traceback.print_exc()
        return False
        
    finally:
        if client:
            logger.info("Stopping Copilot client...")
            try:
                await client.stop()
                logger.info("✓ Cleanup complete")
            except Exception as e:
                logger.warning(f"Error during cleanup: {e}")


async def main():
    """Main entry point"""
    # Parse command line args
    test_case = "mario-hello.automated"
    smoke_test = True
    
    if len(sys.argv) > 1:
        if sys.argv[1] == "--no-smoke":
            smoke_test = False
            test_case = sys.argv[2] if len(sys.argv) > 2 else test_case
        else:
            test_case = sys.argv[1]
    
    logger.info("=" * 70)
    logger.info("smaqit E2E Test Runner (Copilot SDK)")
    logger.info(f"Test case: {test_case}")
    logger.info(f"Smoke test: {'enabled' if smoke_test else 'disabled'}")
    logger.info("=" * 70)
    
    print("\n" + "=" * 70)
    print("smaqit E2E Test Runner (Copilot SDK)")
    print("=" * 70 + "\n")
    
    success = await run_e2e_test(test_case, smoke_test_first=smoke_test)
    
    print("\n" + "=" * 70)
    if success:
        logger.info("RESULT: ✅ PASSED")
        print("RESULT: ✅ PASSED")
        print("=" * 70)
        sys.exit(0)
    else:
        logger.info("RESULT: ❌ FAILED")
        print("RESULT: ❌ FAILED")
        print("=" * 70)
        sys.exit(1)


if __name__ == "__main__":
    asyncio.run(main())
