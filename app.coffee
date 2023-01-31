axios    = require 'axios'
cheerio  = require 'cheerio'
storage  = require 'node-persist'
schedule = require 'node-schedule'
_        = require 'underscore'

search = (meta) ->
  console.log "#{ new Date() }: checking #{ meta.title }"
  url = meta.url.replace('www.mapcamera.com/search', 'www.mapcamera.com/ec/api/itemsearch') + '&siteid=1&limit=100&page=1&devicetype=pc&format=searchresult'
  options = {
    headers: {
      'Referer': meta.url
    }
  }
  try
    body = await axios url, options
    $ = cheerio.load body.data.itemSearchHtml
    items = $(".itembox p.txt")
      .map (i, e) ->
        price = $("span.price span.txtred", e)
        if price.length > 0
          price = price.text()
        else
          price = $("span.price", e).text()
        link = $("a", e).attr("href")

        return {
          title: $("a", e).text(),
          price: price,
          link: link
        }
      .filter (i, e) ->
        return e.price != 'SOLD OUT'
      .toArray()
    return items
  catch e
    console.log "Fetch error: " + e
    return null

getChecker = (meta) ->
  return (items) ->
    await storage.init()
    lastCheck = _.map await storage.getItem(meta.title), (e) -> return e.title
    titles = _.map items, (e) -> return e.title
    newly = _.difference titles, lastCheck
    if newly.length > 0
      console.log "#{ new Date() }: #{ newly.length } new #{ meta.title }s"
      for n in newly
        e = _.filter items, (e) -> return e.title == n
        getNotifier(meta) "New item: #{ n }, price #{ e[0].price }", e[0].link
    console.log "#{ new Date() }: checked #{ meta.title }"
    await storage.setItem meta.title, items

getNotifier = (meta) ->
  return (message, link) ->
    data = {
      "topic": meta.ntfy_topic,
      "message": message,
      "title": "New item!",
      "click": link
    }
    await axios.post "https://ntfy.sh", JSON.stringify(data)

startFunc =  ->
  await storage.init({ dir: 'persist' })
  keywords = await storage.getItem "keywords"
  for keyword in keywords
    do (keyword) ->
      check = getChecker keyword
      items = await search keyword
      await check items

cron = process.argv.slice 2
if cron.length == 0
  startFunc()
else
  cronStr = cron.join " "
  cronStr += " *" for i in [1..(5 - cron.length)]
  console.log "#{ new Date() }: scheduled job #{ cronStr }"
  schedule.scheduleJob cronStr, startFunc
