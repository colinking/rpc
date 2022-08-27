# Example

An example Go API. Based on the [Airplane external API](https://docs.airplane.dev/api/introduction).

To run the API:

```sh
# First, generate the API structs:
go run ./cmd

# Next, run the API:
cd example && go run ./main.go

# Finally, hit the API with a request:
curl -v localhost:4000/v0/runs/get
```