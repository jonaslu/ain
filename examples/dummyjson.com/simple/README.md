Simple example of when just starting out with an API using data from [https://dummyjson.com](https://dummyjson.com/docs).

# Follow along
To follow along you need ain and one of [curl](https://curl.se/), [wget](https://www.gnu.org/software/wget/) or [httpie](https://httpie.io/) installed. Then copy / clone out the *.ain files in this folder to your local drive.

Change the backend under the `[Backend]` section in the templates to whatever you have installed on your computer.

```bash
ain get-users.ain # Get the first 10 users
ID=1 ain get-users-by-id.ain # Get user with id 1

ain get-products.ain # Get the first 10 products
SKIP=10 get-products.ain # Get the next 10 products

ID=3 ain get-product-by-id.ain # Get product with id 3
```

# Layout explanation
The layout is intentionally flat and contains duplication of
the base URL, backend and backend options. There is one file per API-call: getting all users, one user by id, all products and one product by id.

Using ain should be a gradual process. There is plenty of duplication between the files which is ok since it's more important to try out the API than to structure it properly.

You modify the API-call by changing values in the files or commenting out lines. Once the duplication starts to get troublesome ain let's you gradually refactor and extract common parts.
