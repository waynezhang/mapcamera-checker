mapcamera-checker
=================

MapCamera Checker

### Pre

`npm install`  
`npm install -g forever coffee-script`

Copy `persist/keywords.sample` to `persist/keywords` and make the changes.  
_`user_credentials` is the access token of BoxCar for push notifications_

### Run

`forever start -c coffee app.coffee "cronjob schedule string"`  

e.g. `forever start -c coffee app.coffee "0,30 * * * *"`
