### Overview

This tool is used to take a form of http archive (charles log, .har, etc) and generate a k6 load test.

### Design Goals
* Assume "Worst Case Scenarios" where assumptions can be made
* Make the process of creating a load test based off real app behavior as quick as possible. For example, actually interacting with the application and capturing charles logs, and generating a realistic load test from it should require little to no user intervention.
* Achieve a high level of "Realistic Entropy", meaning that the generated tests have a level of randomness in order to simulate real world traffic as much as possible


# Dev Setup

```
./bin/setup
```

# Dev Usage

```
./bin/run-dev go run main.go create test-data/Android.chls -f your.domain.com
```

The above command will then output the load test to stdout. You can then copy it and save it to a file.

From there, just run it with K6

```
k6 yourFile.js
```

## The Nitty Gritty...

Essentially what this repo allows someone to do, is record a charles log, then auto generate a k6 load test.

###### What about my custom business logic??

You may ask.. Well this tool will let you apply that, via a `hookFile`. A hookFile is just a simple javascript file. You can start creating your own by running:

```
./bin/run-dev go run main.go default-template
```

The `default-template` is just a basic starting point. The idea is if you have the same business logic for a given set of API endpoints, you can just create a `hookFile`, which can then be applied to all Charles / .har captures, which will then allow you to literally auto-generate load tests completely automatically!

You should run the command above if you haven't yet. Here is a reference for it:

`k6Options` - An object that you can use to set the k6Options as defined [here](https://docs.k6.io/docs/options)

`before()` - Called before each time the "default" k6 function is called. This could be used to generate a random email for a user registration request

`requestHook()` - Called before any http call. The raw object that would be used on the call is passed in, and you can modify it in any way you choose

`responseHook()` - Called after every http response. This is so you can extract values as needed, then set them to a global state to be used in further reqeusts

`after()` - Called each time after the "default" k6 function is called. This is where you can reset any previous state you might have set in the `responseHook()`

A typical Scenario for using each of the functions is as follows:

* Assume you are load testing a set of apis that all require an authorization header obtained from a login endpoint.
* The test consists of hitting a register endpoint, login endppoint, a get user info endpoint, and a friend request endpoint.

Using this tool, you can:
* construct a random email to be used for register in the `before()` function
* save the response from login with the `responseHook()` function
* send the auth header on all subsequent api requests 

From this point on, you can capture charles log after charles log of requests, and then apply specific and generic business rules to ALL generated files!

### but what if I need to make a very specific change to just one test? ..

SIMPLE! just edit the produced javascript file in anyway you need!

---

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