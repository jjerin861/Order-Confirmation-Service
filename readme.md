# Order confirmation server

Order confirmation server.

## Steps to Execute 

#### 1. export envs 
```bash
export PORT=8080
```
#### 2. start the application
```bash
go run main.go
```


### payment_confirmation [POST]

> http://localhost:8080/order_status/webhook/payment_confirmation

Sample request body
```bash
{
	"order_id": "de-ber-76898",
	"amount": 2000,
	"payment_reference": "h87j87y9q34rjo8rweqjo9fdckjhdslkmsdf",
	"payment_status": "confirmed"
}
```

### fraud_check [POST]

> http://localhost:8080/order_status/webhook/fraud_check

Sample request body
```bash
{
	"reference_id": "de-ber-76898",
	"risk_points": 10 
}
```

### vendor_confirmation [POST]

> http://localhost:8080/order_status/webhook/vendor_confirmation

Sample request body
```bash
{
	"order": "de-ber-76898",
	"status": "confirmed"
}
```
