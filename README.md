# redditoauth
* A reddit oauth library written in go
* Reddit is not my copyright, They are an amazing company and deserve all the kutos for their apis and oauth work

## Instructions
* Before you can use this library, you need to preload the client secret, clientID, and user agent string as follows:

```
import "github.com/allonsy/redditoauth"

redditoauth.SetClientID("clientid")
redditoauth.SetClientSecret("clientsecret")
redditoauth.SetUserAgent("useragent")
```

* Failure to do this will return errors in the library functions that describe which fields have not been filled out correctly
* Then, you can call `performHandshake` as follows:

```
redditoauth.PerformHandshake("http://localhost", []string{"identity"}, true))
```

  * The first arg is the redirect_uri (localhost is a good idea for installed apps)
  * The second arg is a slice of strings that correspond to scopes
  * the last arg is a boolean that is true if we want to get a refresh token and false otherwise
  * The function will output a url to the console, visit this url, click agree and then take the code from the query string that it redirects to and paste it into the console (the program will wait for the code)
  * the function auto populates the access code and refresh token when a response from reddit is received
  * The function returns an access token and refresh token and an error if any occurs
  * IT IS YOUR RESPONSIBILITY TO SAVE THESE VALUES
  * On startup of the app, you will need to populate access token and refresh token value with `SetAccessToken(string)` and `SetRefreshToken(string)`
  * So long as the app is on, it will cache these tokens so you don't need to keep on re calling these functions or asking reddit for more tokens
  * If you forget to populate the access token and refresh token fields, my library will error, asking you to perform a handshake again
* Making requests:
  * After you have performed a handshake or set access/refresh tokens properly call `MakeApiReq(method, url, body, interface{})`
  * The method is the http method (GET, POST, etc...)
  * The url is the full url with all query parameters like `https://oauth.reddit.com/api/v1/me`
  * The optional body is an io.Reader that has the body of the request (for posts). If no body is needed, set this to `nil`
  * The last argument is a map or object for golang's json library to put the parsed json into (see golang's json library documentation for more info)
  * The function will set the error value if an error occurs
  * If the access token has expired, the code will auto refresh it with the refresh token. If no refresh token is provided, an error will be returned.

## Contributing
* If you have any feature requests, issues, or bugs, please don't be afraid to fill out issues
* All pull requests are welcome, just fork and submit a PR. Please format all go sources with `gofmt`
