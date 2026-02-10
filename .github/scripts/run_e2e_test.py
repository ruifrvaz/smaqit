#!/usr/bin/env python3
"""
E2E Test Runner using Copilot SDK

Invokes smaqit.ci-testing agent programmatically for CI/CD automation.
"""

import asyncio
import sys
import time
from pathlib import Path
from copilot import CopilotClient


async def show_progress(task, description="Working"):
    """Show progress updates while task runs"""
    start = time.time()
    dots = 0
    while not task.done():
        elapsed = int(time.time() - start)
        mins, secs = divmod(elapsed, 60)
        dot_str = "." * (dots % 4)
        print(f"\r⏱  {description}{dot_str:<3} [{mins:02d}:{secs:02d}]", end="", flush=True)
        dots += 1
        await asyncio.sleep(30)
    print()  # New line after completion


async def run_e2e_test(test_case: str = "mario-hello.automated") -> bool:
    """
    Run end-to-end test using smaqit.ci-testing agent.
    
    Args:
        test_case: Test case name from docs/test-cases/ (e.g., "mario-hello.automated")
        
    Returns:
        True if test passed, False otherwise
    """
    print(f"🚀 Starting E2E test: {test_case}")
    print("=" * 70)
    
    client = None
    try:
        # Initialize Copilot client
        print("Initializing Copilot SDK...")
        client = CopilotClient()
        await client.start()
        print("✓ Connected to Copilot CLI\n")
        
        # Create session
        print("Creating session...")
        session = await client.create_session()
        print("✓ Session created\n")
        
        # Invoke CI testing agent
        print(f"Invoking /smaqit.ci-testing with test case: {test_case}")
        print("-" * 70)
        
        prompt = f"/smaqit.ci-testing run test case {test_case} from docs/test-cases/"
        
        # Run with progress indicator
        task = asyncio.create_task(
            session.send_and_wait(
                {"prompt": prompt},
                timeout=1500  # 25 minutes for full E2E test (CI environment is slower than local)
            )
        )
        await show_progress(task, "Agent executing test workflow")
        response = await task
        
        if not response or not hasattr(response.data, 'content'):
            print("❌ No response from testing agent")
            return False
        
        content = response.data.content
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
            print("✅ E2E test PASSED")
            return True
        else:
            print("❌ E2E test FAILED or UNCLEAR")
            return False
            
    except Exception as e:
        print(f"\n❌ Error during test execution:")
        print(f"   {type(e).__name__}: {e}")
        import traceback
        traceback.print_exc()
        return False
        
    finally:
        if client:
            print("\nStopping Copilot client...")
            await client.stop()
            print("✓ Cleanup complete")


async def main():
    """Main entry point"""
    # Get test case from command line args (default: mario-hello.automated)
    test_case = sys.argv[1] if len(sys.argv) > 1 else "mario-hello.automated"
    
    print("\n" + "=" * 70)
    print("smaqit E2E Test Runner (Copilot SDK)")
    print("=" * 70 + "\n")
    
    success = await run_e2e_test(test_case)
    
    print("\n" + "=" * 70)
    if success:
        print("RESULT: ✅ PASSED")
        print("=" * 70)
        sys.exit(0)
    else:
        print("RESULT: ❌ FAILED")
        print("=" * 70)
        sys.exit(1)


if __name__ == "__main__":
    asyncio.run(main())
