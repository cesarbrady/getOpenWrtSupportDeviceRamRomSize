The official list of devices supported by openwrt is here: https://openwrt.org/toh/start

Although you can see the list with a browser, you can't see this list when viewing the source code of the page, because the list is rendered by javascript.

So you first need a javascript engine to run javascript to render the table, and then get the html page to do the analysis.

I checked my previous notes and found that splash can do this.

Run splash server with this command, with ip address specify

```bash
docker run --rm --network mynet --ip 172.18.0.10 --name splash --hostname splash scrapinghub/splash
```

If you do not have the network "mynet", you can run this command command to create

```bash
docker network create --subnet=172.18.0.0/16 mynet
```

Test to render the web page with the splash server, you should get the html source with the list inside.

```bash
curl 'http://172.18.0.1:8050/render.html?url=https://openwrt.org/toh/start'
```

Here is the golang program souce code

```go
package main

import "fmt"

func main() {
	var devcontent string

	db := getSQLite("device.db")
	db.createTable("device").
		addColumn("brand", "string").
		addColumn("model", "string").
		addColumn("version", "string").
		addColumn("release", "string").
		addColumn("flash", "int").
		addColumn("ram", "int")

	content := httpGet("http://your.splash.host/render.html?url=https://openwrt.org/toh/start").content
	for _, i := range reFindAll("<td class=\"leftalign brand\">(.+?)</td><td class=\"leftalign model\">(.+?)</td><td class=\"leftalign version\">(.*?)</td><td class=\"centeralign supported_current_rel\"><a.+?>(.+?)</a></td>.+?<td class=\"align device_techdata\"><a href=\"(.+?)\".+?</td>", content) {
		try(func() {
			brand := strStrip(i[1])
			model := strStrip(i[2])
			version := strStrip(i[3])
			release := strStrip(i[4])
			techdataurl := strStrip(i[5])
			for {
				if err := try(func() {
					devcontent = httpGet("https://openwrt.org/" + techdataurl).content
				}).Error; err == nil {
					break
				}
			}
			flash := reFindAll("<dd class=\"flash_mb\">([0-9]+).*?</dd>", devcontent)
			ram := reFindAll("<dd class=\"ram_mb\">([0-9]+).*?</dd>", devcontent)
			fmt.Println(brand, "\t", model, "\t", version, "\t", release, "\t", flash[0][1]+"mb", "\t", ram[0][1]+"mb")
			db.table("device").data(map[string]interface{}{
				"brand":   brand,
				"model":   model,
				"version": version,
				"release": release,
				"flash":   flash[0][1],
				"ram":     ram[0][1],
			}).insert()
		}).except(func(e eee) {
			lg.error(e)
		})
	}
}
```

It will use regular expressions to parse the content of the list, get the link to device detailed information, get the source code of this link and find the flash and ram size fields, and finally store the result in the sqlite database.

The following is part of the database

```
sqlite> select * from device limit 20;
id          brand       model         version     release     flash       ram   
----------  ----------  ------------  ----------  ----------  ----------  ----------
1           3Com        3CRWER100-75              14.07       4           16  
2           3Com        3CRWER100-75              14.07       4           16  
3           4G Systems  AccessCube (              8.09.2      32          64  
4           7Links      PX-4885                   19.07.2     4           32  
5           8devices    Carambola 1               19.07.6     8           32  
6           8devices    Carambola 2               19.07.6     16          64  
7           8devices    Lima                      19.07.6     32          64  
8           8devices    Rambutan                  19.07.6     128         128   
9           8devices    Jalapeno                  19.07.6     8           256   
10          8devices    Habanero DVK              snapshot    32          512   
11          Abicom Int  Freedom CPE   Rev 05      10.03       8           32  
12          Abicom Int  Scorpion SC4  Rev 02      19.07.6     16          256   
13          Abicom Int  Scorpion SC1  4           19.07.6     16          256   
14          Abicom Int  Scorpion SC3              19.07.6     16          256   
15          Accton      MR3201A                   8.09        8           32  
16          Accton      WR6202                    19.07.6     8           32  
17          Actiontec   GT701         C, D        10.03.1     4           16  
18          Actiontec   GT704WG       1A          10.03.1     4           16  
19          Actiontec   GT784WNV      5A          19.07.6     16          64  
20          Actiontec   MI424WR       A           18.06.1     8           32  
Run Time: real 0.001 user 0.000000 sys 0.000578
sqlite>
```

GitHub: https://github.com/cesarbrady/getOpenWrtSupportDeviceRamRomSize
