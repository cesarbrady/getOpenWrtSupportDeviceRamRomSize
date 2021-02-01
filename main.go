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
