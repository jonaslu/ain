# Simple layout using dummyjson.com
This folder is an example of what it can look like when just starting out with an API. We are using data from [https://dummyjson.com](https://dummyjson.com/docs).

Using ain should be a gradual process. You start out with 2-3 files for same API. There is plenty of duplication between the files which is ok since it's more important to try out the API initially. You modify the API-call by changing values in the files or comment out lines.

Once the duplication starts to get troublesome ain let's you gradually refactor and extract common parts.

To follow along you need ain and one of [curl](https://curl.se/), [wget](https://www.gnu.org/software/wget/) or [httpie](https://httpie.io/) installed.

Change the backend under the `[Backend]` section in the templates to whatever you have installed on your computer.

```bash
ain get-users.ain # Get the first 10 users
ID=1 ain get-users-by-id.ain # Get user with id 1

ain get-products.ain # Get the first 10 products
SKIP=10 get-products.ain # Get the next 10 products

ID=3 ain get-product-by-id.ain # Get product with id 3
```
