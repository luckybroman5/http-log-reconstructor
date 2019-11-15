
const k6Options = {
    // place your K6 options here
}

function before() {
    // Place code here to be called before each cycle of the VU
}

// The requestHook is called just before a request is made by k6 in order to modify params as wanted
function requestHook(request) {
    /**
        request.headers <string>headers
        request.query []query
        request.method string
        request.body string
        request.url string
    */
    
    /** EXAMPLE: Only do the following logic if the domain is `example.com`
    if (request.url.indexOf("example.com") !== -1) {
        // Add the authorization header to be whatever it needs to be for the session, but only if it's present in the request
        if (request.headers.authorization) request.headers.authorization = authToken

        // If there is an email in the request body, set it to something
        request.body.email = email

    }*/

    // Always have to return!
    return request
}

// The response hook is called whenever a response from k6 is recieved, so values can be set to be used for subsequent requests
function responseHook(response) {
    /**
        response.headers []headers
        response.query []query
        response.method string
        response.body string
    */

    /** EXAMPLE: Setting a variable from a login call
    if (response.request.url.indexOf('/login') !== -1 && response.body.accessToken) {
        authToken = response.body.accessToken
    }*/

    // Always have to return!
    return response
}

function done() {
    // Put code here to be called after all the requests for the VU Cylce are complete
}
