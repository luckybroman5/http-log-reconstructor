## Overview

In order to implement wildcards, the `comment` field will be looked at in the javascript generation. The exact specification for this comment field in the `.har` files for headers, parameters, response bodies, etc, can be found here: https://w3c.github.io/web-performance/specs/HAR/Overview.html#sec-object-types-headers

### Goals:
* Provide flexibility to change / modify the behavior of certain parameters based off config, and avoid hardcoding
* Allow parameters from other requests to be used for subsequent requests
* Allow parameters to be pulled from a list randomly
* Allow parmeters to be randomly generated
* Allow a WHOLE URL to be dynamically read from another request

### Implementation:
Some form of Configuration data must be fed through for particular parameters / endpoints to different domains.

##### The command is invoked with a --hookFile argument
```
--hookFile somefile.js
```

The hookfile is then used to wrap all requests being called by k6. In this file, you can apply any rule you want applying to any specific domain.

```
// somefile.js

// global variables
authToken = ''
email = ''

// Example function to be called in the hooks, note that it's not exported
function getRandomEmail() {
    return 'loadtestuser+' + getRandomNumber + '@fox.com'
}

// The before function is called before each cycle of the VU
exports default before function() {
    // Sets the global email variable before the test runs
    email = getRandomEmail()
}

// The requestHook is called just before a request is made by k6 in order to modify params as wanted
exports default requestHook function(request) {
    /**
        request.headers [string]headers
        request.query []query
        request.method string
        request.body string
        request.url string
    */
    
    // Only do the following logic if the domain is `api3.fox.com`
    if (request.url.indexOf("api3.fox.com") !== -1) {
        // Add the authorization header to be whatever it needs to be for the session, but only if it's present in the request
        if (request.headers.authorization) request.headers.authorization = authToken

        // If there is an email in the request body, set it to something
        request.body.email = email

    }
    
    // Always have to return the result
    return request
}

// The response hook is called whenever a response from k6 is recieved, so values can be set to be used for subsequent requests
exports default responseHook function(response) {
    /**
        response.headers []headers
        response.query []query
        response.method string
        response.body string
    */

    // If an accessToken is returned from a login request, set the global variable
    if (response.request.url.indexOf('/login') !== -1 && response.body.accessToken) {
        authToken = response.body.accessToken
    }
}

// The done hook is called after all the requests for the VU Cylce are complete to clean up any unwanted variables
exports default done function() {
    // Reset the auth token to be nothing so the next time a test is ran, it'll be 
    authToken = ''
}

```

