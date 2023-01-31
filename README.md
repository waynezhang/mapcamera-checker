mapcamera-checker
=================

MapCamera Checker

### Pre

`npm install`  
`npm install -g forever coffee-script`

Copy `persist/59aeb2c9970b7b25be2fab2317e31fcb.sample` to `persist/59aeb2c9970b7b25be2fab2317e31fcb` and make the changes.  
_`ntfy_topic` is the topic of ntfy.sh for push notifications_

### Run

`forever start -c coffee app.coffee "cronjob schedule string"`  

e.g. `forever start -c coffee app.coffee "0,30 * * * *"`
