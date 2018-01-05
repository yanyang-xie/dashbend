Benchmark Test Tool Design:
--------------------------

#### Key Point

* Benchmark Test Executor
    >Benchmark executor is to send dash request to VEX accoring to configred rate.

* Benchmark Test Validator
    >Benchmark test validator is to check VEX response.

* Benchmark Test Reporter
    >Collect statistics information of request/response and validation from test executor and validator, then provide a API to get the report with json format.

* Distributed Test
    > Using fabric to start/stop multiple benchmark test

* Test Result Collection
    > Fetch test result from all the test executor and store summarized test result into DB.

* Benchmark Test UI
    > Support to change warmup/benchmark number and then start/stop benchmark test from UI.
    > Support to check test result from UI.
    
    
---------
- Benchmark Test Executor
   - Read config file path from command line.
   - Read configured value from INI.
   - Generate request URLs by configured rule.
   - Do warm up process according to configured warmup rate.
   - Do benchmark test according to configured benchmark rate.
   - Provide basic request/response info to data collector
   - Provide response content to validator.



