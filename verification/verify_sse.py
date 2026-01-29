from playwright.sync_api import sync_playwright

def verify_sse(page):
    # Track requests to SSE endpoint
    sse_requests = []
    page.on("request", lambda request: sse_requests.append(request.url) if "/api/scales/stream" in request.url else None)

    print("Navigating to login...")
    page.goto("http://localhost:8080/login")

    # Login
    page.fill("input[name='username']", "indraaja")
    page.fill("input[name='password']", "indraaja")
    page.fill("input[name='captcha']", "123456") # Magic code
    page.click("button[type='submit']")

    # Wait for dashboard
    page.wait_for_url("http://localhost:8080/dashboard")
    print("Logged in, on dashboard.")

    # Wait 3 seconds to see if SSE connects
    page.wait_for_timeout(3000)

    # Check SSE requests on Dashboard
    dashboard_sse_count = len(sse_requests)
    print(f"SSE requests on Dashboard: {dashboard_sse_count}")

    if dashboard_sse_count == 0:
        print("PASS: No SSE requests on Dashboard.")
    else:
        print("FAIL: SSE requests detected on Dashboard.")

    page.screenshot(path="verification/dashboard.png")

    # Navigate to Weighing
    print("Navigating to Weighing page...")
    # Clear previous requests tracking (optional, but easier to count)
    sse_requests.clear()

    page.click("a[href='/weighing']")
    page.wait_for_url("http://localhost:8080/weighing")

    # Wait for SSE to connect
    # It might take a moment
    try:
        page.wait_for_function("() => window.weighingSSE && window.weighingSSE.readyState !== 2", timeout=5000)
        print("PASS: SSE object exists and is not closed.")
    except Exception as e:
        print(f"FAIL: SSE object not found or closed. {e}")

    # Wait a bit more to catch the request in our listener
    page.wait_for_timeout(2000)

    weighing_sse_count = len(sse_requests)
    print(f"SSE requests on Weighing: {weighing_sse_count}")

    if weighing_sse_count > 0:
        print("PASS: SSE request detected on Weighing page.")
    else:
        print("FAIL: No SSE request detected on Weighing page.")

    page.screenshot(path="verification/weighing.png")

if __name__ == "__main__":
    with sync_playwright() as p:
        browser = p.chromium.launch()
        page = browser.new_page()
        try:
            verify_sse(page)
        except Exception as e:
            print(f"Error: {e}")
        finally:
            browser.close()
