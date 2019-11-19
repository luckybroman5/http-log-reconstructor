### Overview

This tool is used to take a form of http archive (charles log, .har, etc) and generate a k6 load test.

### Design Goals
* Assume "Worst Case Scenarios" where assumptions can be made
* Make the process of creating a load test based off real app behavior as quick as possible. For example, actually interacting with the application and capturing charles logs, and generating a realistic load test from it should require little to no user intervention.
* Achieve a high level of "Realistic Entropy", meaning that the generated tests have a level of randomness in order to simulate real world traffic as much as possible

#### Current Work:
* Cleanup Implementation for the first release
* Add random variation between sleep cylces
* Import helper functions and run browserify for a more rich feature set of load tests

#### Very Cool to have work:
* Create a Web UI

## Technologies:
* Docker
* Golang
* javascript
* [K6](https://docs.k6.io/docs/)
* [Charles Proxy Cli](https://www.charlesproxy.com/documentation/tools/command-line-tools/)


# Completed Work:

#### ~~MVP Requirements:~~
* ~~Parse a charles log and extract whitlisted requests~~
* ~~output some form of load test to be ran~~
* ~~Obtain and document “Wildcard values”, meaning things like dma, authorization, apikey, geo lat and long, etc are not captured for security / privacy and the ability variate the values in a load test~~