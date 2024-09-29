# dummyjson.com advanced example
Example of an advanced layout of an API using data from [https://dummyjson.com](https://dummyjson.com/docs)

# Follow along
To follow along you need ain and one of [curl](https://curl.se/), [wget](https://www.gnu.org/software/wget/) or [httpie](https://httpie.io/) installed. Then copy / clone out the *.ain files in this folder to your local drive.

Change the backend under the `[Backend]` section in the templates to whatever you have installed on your computer.

Getting an JWT token:
```bash
ain base.ain get-token.ain # Gets the JWT token which is then used when calling auth endpoints
```

Products:
```bash
ain base.ain products/get.ain # Get the 30 first products

LIMIT=10 SKIP=30 ain base.ain.ain products/get.ain paginate.ain # Get the next 10 products after the initial 30

ain base.ain products/add.ain!            # Add a product via one-off edited data
ID=1 ain base.ain products/get-by-id.ain  # Get product with ID 1
ID=2 ain base.ain products/update.ain!    # Update product 2 with one-off data (adding a ! after the file to edit it in-place)
ID=3 ain base.ain products/delete.ain     # Delete product 3

ain base.ain auth.ain products/get.ain # Get the first 30 products, with an Authorization Bearer: <token> header
```

# Layout explanation
The layout have been selected to contain minimal repetition.

## base.ain
Contains everything that is common to all calls to this API: a base-url, the preferred backend (curl in this example), any backend-options, common headers and timeouts.


## get-token.ain
Most REST APIs will contain calls that require authentication typically via an `Authorization: Bearer <token>`.

Working with ain you can work out the proper call the authorization endpoint and then extract the token out as a separate step, before you integrate it into the authorized API call.

`ain base.ain get-token.ain` will return the whole JWT payload.

## auth.ain
Now that we have a way of getting the JWT token, we can invoke ain from ain and use jq to extract the Bearer token. Then we insert it into an `Authorization: Bearer` header.

This can be made a simple or advanced as you'd like. Since it's returned via an executable you can easily make a shell-script to check the expiration of any existing JWT, or request a new token via a refresh token, before calling a token endpoint. Or you can hit the token endpoint every time.

## paginate.ain
Most REST endpoints have some pagination and these are usually supplied as query-parameters. This file contains both an limit and an offset and can be included with the call to any endpoint. Since query parameters are applied after the URL has been assembled the file itself can go anywhere file-list.

Supported parameters are:
```bash
LIMIT=n # LIMIT mandatory
SKIP=n  # SKIP is optional via bash if / else
```

## products/
All files concerning products are grouped into a folder called products/.

```bash
products/add.ain
products/delete.ain
products/get-by-id.ain
products/get.ain
products/update.ain
```

Composing these files with base.ain (and auth.ain and paginate.ain if need be) makes the resulting command-line readable:

```bash
ID=1 ain base.ain products/get-by-id.ain # Gets product with id 1
```
