# go rest api
____
### go version 1.19.2
### mysql version 8.0.31

____
### How to use
- Clone this repository 
- Run command: docker-compose up -d
- Navigate to browser and use endpoints (see below)

____
```
There are two endpoints available:
127.0.0.1:8080/api/upload - for upload .csv file
127.0.0.1:8080/api/getdata - for getting data from database using filters

For upload file:
use POST request with form-data body: key = uploadfile (type = file), value = .csv file

For getting data:

use GET request with parameters:
1. key = transaction_id, value = integer number (example: 127.0.0.1:8080/api/getdata?transaction_id=70)
2. key = status, value = accepted or declined (example: 127.0.0.1:8080/api/getdata?status=declined)
3. key = payment_type, value = cash or card (example: 127.0.0.1:8080/api/getdata?payment_type=card)
4. key = payment_narrative, value = full or partial string (example: 127.0.0.1:8080/api/getdata?payment_narrative=27123)
5. key = date_post, value = date from, date to comma separeted (example: 127.0.0.1:8080/api/getdata?date_post=2022-08-12,2022-08-13)
6. key = terminal_id, value = one or several terminals id comma separeted (example: 127.0.0.1:8080/api/getdata?terminal_id=3515,3602)

NOTE: only one key can be used!

For GET request response will be data in JSON format

```
