# kraken-trader
A Go app to aid trading on Kraken

+ ## Testing
+ 
+ The project includes both unit tests and integration tests. Integration tests require Kraken API demo credentials.
+ 
+ ### Prerequisites
+ 
+ Install the test runner:
+ ```bash
+ make install-tools
+ ```
+ 
+ ### Running Tests
+ 
+ ```bash
+ # Run unit tests only (fast, no credentials needed)
+ make test
+ 
+ # Run integration tests only (requires KRAKEN_DEMO_KEY and KRAKEN_DEMO_SECRET)
+ make test-integration
+ 
+ # Run all tests
+ make test-all
+ 
+ # Generate test coverage report
+ make test-coverage
+ ```
+ 
+ ### Setting Up Integration Tests
+ 
+ 1. Get demo API credentials from [Kraken's Testing Environment](https://support.kraken.com/hc/en-us/articles/360024809011-API-Testing-Environment)
+ 2. Set environment variables:
+    ```bash
+    export KRAKEN_DEMO_KEY="your_demo_key"
+    export KRAKEN_DEMO_SECRET="your_demo_secret"
+    ```
+ 
+ Integration tests will be skipped if credentials are not provided.
+ 
+ ### Test Timeouts
+ 
+ - Unit tests: 10 seconds
+ - Integration tests: 30 seconds
+ - Full test suite: Should complete within 1 minute
