request  = require 'request'
cheerio  = require 'cheerio'
storage  = require 'node-persist'
schedule = require 'node-schedule'
_        = require 'underscore'

search = (meta, callback) ->
  console.log "#{ new Date() }: checking #{ meta.title }"
  url = meta.url
  request url, (err, resp, body) ->
    if err then return
    $ = cheerio.load body
    items = $(".itembox p.txt")
      .map (i, e) ->
        price = $("span.price span.txtred", e)
        if price.length > 0
          price = price.text()
        else
          price = $("span.price", e).text()

        return {
          title: $("a", e).text(),
          price: price
        }
      .filter (i, e) ->
        return e.price != 'SOLD OUT'
      .toArray()
    callback items

getChecker = (keyword) ->
  return (items) ->
    storage.initSync()
    lastCheck = _.map storage.getItem(keyword.title), (e) -> return e.title
    titles = _.map items, (e) -> return e.title
    newly = _.difference titles, lastCheck
    if newly.length > 0
      console.log "#{ new Date() }: #{ newly.length } new #{ keyword.title }s"
      for n in newly
        p = _.filter items, (e) -> return e.title == n
        getNotifier(keyword) "New Item: #{ n }, price #{ p[0].price }"
    console.log "#{ new Date() }: checked #{ keyword.title }"
    storage.setItem keyword.title, items

getNotifier = (meta) ->
  return (message) ->
    data = { user_credentials: meta.user_credentials, "notification[title]": message }
    request.post "https://new.boxcar.io/api/notifications", { form: data }

startFunc =  ->
  storage.initSync()
  keywords = storage.getItem "keywords"
  search keyword, getChecker(keyword) for keyword in keywords

cron = process.argv.slice 2
if cron.length == 0
  startFunc()
else
  cronStr = cron.join " "
  cronStr += " *" for i in [1..(5 - cron.length)]
  console.log "#{ new Date() }: scheduled job #{ cronStr }"
  schedule.scheduleJob cronStr, startFunc
