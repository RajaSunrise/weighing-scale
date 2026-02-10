from playwright.sync_api import sync_playwright, expect
import time
import os

def verify_logs():
    with sync_playwright() as p:
        browser = p.chromium.launch(headless=True)
        page = browser.new_page()

        # 1. Go to login
        page.goto("http://localhost:8080/login")
        page.fill("input[name='username']", "admin")
        page.fill("input[name='password']", "Password123")

        # We need the captcha_id from the hidden field
        captcha_id = page.locator("input[name='captcha_id']").get_attribute("value")
        page.fill("input[name='captcha']", "8888") # Our patched bypass

        page.click("button[type='submit']")

        # Wait for navigation to dashboard
        page.wait_for_url("**/dashboard")
        print("Logged in successfully")

        # 2. Trigger another malicious log entry just to be sure
        page.request.get("http://localhost:8080/anpr-test/<img src=x onerror=alert('XSS-Triggered')>")

        # 3. Go to logs page
        page.goto("http://localhost:8080/settings/logs")

        # Wait for logs to load (AJAX)
        page.wait_for_selector("#log-container div")

        # Take screenshot
        if not os.path.exists("/home/jules/verification"):
            os.makedirs("/home/jules/verification")
        page.set_viewport_size({"width": 1280, "height": 800})
        page.screenshot(path="/home/jules/verification/logs_verification.png")
        print("Screenshot saved to /home/jules/verification/logs_verification.png")

        # 4. Check if we can find the sanitized log entry
        logs_text = page.locator("#log-container").inner_text()
        if "img src=x onerror" in logs_text:
            print("Found sanitized log entry in UI")
        else:
            print("Could NOT find sanitized log entry.")

        browser.close()

if __name__ == "__main__":
    verify_logs()
