
Instructions to run code

1. run main.go
1. make a GET request to localhost:8080/split-bill
1. send json body, template below

```json
{
   "people":[
      {
         "name":"William",
         "items":[
            {
               "itemName":"",
               "price":0.00
            }
         ]
      },
      {
         "name":"John",
         "items":[
            {
               "itemName":"",
               "price":0.00
            }
         ]
      },
      {
         "name":"Aden",
         "items":[
            {
               "itemName":"",
               "price":0.00
            }
         ]
      },
      {
         "name":"Dan",
         "items":[
            {
               "itemName":"",
               "price":0.00
            }
         ]
      },
      {
         "name":"Joel",
         "items":[
            {
               "itemName":"",
               "price":0.00
            }
         ]
      }
   ],
   "sharedItems":[
      {
         "people":[
            {
               "name":"name1"
            },
            {
               "name":"name2"
            }
         ],
         "items":[
            {
               "itemName":"",
               "price":0.00
            }
         ]
      }
   ],
   "tipPaid": 0.0,
   "taxPaid": 0.0,
}
```