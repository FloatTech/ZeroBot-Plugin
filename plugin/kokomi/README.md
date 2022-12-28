文档说明 kokomi_v2
安装教程
群文件下载压缩包,保证版本一致
解压文件两个压缩包,得到两个名为kokomi的文件夹(不要合并)
将小(占内存)的那个放在plugin文件夹,大的放在data文件夹,完成后保证文件夹名为kokomi

//新建图层,实现阴影400*510
		bg := Yinying(400, 510, 16)

"好感度"+strconv.Itoa(alldata.AvatarInfoList[t].FetterInfo.ExpLevel
"昵称:"+alldata.PlayerInfo.Nickname
"uid:"+suid
角色"LV"+strconv.Itoa(alldata.PlayerInfo.ShowAvatarInfoList[t].Level)
命之座strconv.Itoa(len(alldata.AvatarInfoList[t].TalentIDList))+"命"
        one.DrawString("生命值:"+strconv.Itoa(int(alldata.AvatarInfoList[t].FightPropMap.Num2000)), 5, 65)
		one.DrawString("攻击力:"+strconv.Itoa(int(alldata.AvatarInfoList[t].FightPropMap.Num2001)), 5, 125)
		one.DrawString("防御力:"+strconv.Itoa(int(alldata.AvatarInfoList[t].FightPropMap.Num2002)), 5, 185)
		one.DrawString("元素精通:"+strconv.Itoa(int(alldata.AvatarInfoList[t].FightPropMap.Num28)), 5, 245)
		one.DrawString("暴击率:"+strconv.Itoa(int(alldata.AvatarInfoList[t].FightPropMap.Num20*100))+"%", 5, 305)
		one.DrawString("暴击伤害:"+strconv.Itoa(int(alldata.AvatarInfoList[t].FightPropMap.Num22*100))+"%", 5, 365)
		one.DrawString("元素充能:"+strconv.Itoa(int(alldata.AvatarInfoList[t].FightPropMap.Num23*100))+"%", 5, 425)
        // 元素加伤判断
		switch {
		case alldata.AvatarInfoList[t].FightPropMap.Num30*100 > 0:
			one.DrawString("物理加伤:"+strconv.Itoa(int(alldata.AvatarInfoList[t].FightPropMap.Num30*100))+"%", 5, 485)
		case alldata.AvatarInfoList[t].FightPropMap.Num40*100 > 0:
			one.DrawString("火元素加伤:"+strconv.Itoa(int(alldata.AvatarInfoList[t].FightPropMap.Num40*100))+"%", 5, 485)
		case alldata.AvatarInfoList[t].FightPropMap.Num41*100 > 0:
			one.DrawString("雷元素加伤:"+strconv.Itoa(int(alldata.AvatarInfoList[t].FightPropMap.Num41*100))+"%", 5, 485)
		case alldata.AvatarInfoList[t].FightPropMap.Num42*100 > 0:
			one.DrawString("水元素加伤:"+strconv.Itoa(int(alldata.AvatarInfoList[t].FightPropMap.Num42*100))+"%", 5, 485)
		case alldata.AvatarInfoList[t].FightPropMap.Num44*100 > 0:
			one.DrawString("风元素加伤:"+strconv.Itoa(int(alldata.AvatarInfoList[t].FightPropMap.Num44*100))+"%", 5, 485)
		case alldata.AvatarInfoList[t].FightPropMap.Num45*100 > 0:
			one.DrawString("岩元素加伤:"+strconv.Itoa(int(alldata.AvatarInfoList[t].FightPropMap.Num45*100))+"%", 5, 485)
		case alldata.AvatarInfoList[t].FightPropMap.Num46*100 > 0:
			one.DrawString("冰元素加伤:"+strconv.Itoa(int(alldata.AvatarInfoList[t].FightPropMap.Num46*100))+"%", 5, 485)
		default: //草或者无
			one.DrawString("元素加伤:"+strconv.Itoa(int(alldata.AvatarInfoList[t].FightPropMap.Num43*100))+"%", 5, 485)
		}

        // 天赋等级
		if err := dc.LoadFontFace(FontFile, 65); err != nil { // 字体大小
			panic(err)
		}
		var link = []int{0, 0, 0}
		var i = 0
		for k, _ := range alldata.AvatarInfoList[t].SkillLevelMap {
			link[i] = k
			i++
		}
		sort.Ints(link)
		lin1, _ := alldata.AvatarInfoList[t].SkillLevelMap[link[0]]
		lin2, _ := alldata.AvatarInfoList[t].SkillLevelMap[link[1]]
		lin3, _ := alldata.AvatarInfoList[t].SkillLevelMap[link[2]]
    "天赋等级:"+strconv.Itoa(lin1)+"--"+strconv.Itoa(lin2)+"--"+strconv.Itoa(lin3)

    武器名字wq, _ := IdforNamemap[alldata.AvatarInfoList[t].EquipList[5].Flat.NameTextHash]
    "精炼:"+strconv.Itoa(int(alldata.AvatarInfoList[t].EquipList[5].Flat.RankLevel))
    "攻击力:"+strconv.FormatFloat(alldata.AvatarInfoList[t].EquipList[5].Flat.WeaponStat[0].Value
    "Lv:"+strconv.Itoa(alldata.AvatarInfoList[t].EquipList[5].Weapon.Level
    //副词条
		fucitiao, _ := IdforNamemap[alldata.AvatarInfoList[t].EquipList[5].Flat.WeaponStat[1].SubPropId] //名称
		var baifen = "%"
		if fucitiao == "元素精通" {
			baifen = ""
		}
        fucitiao+":"+strconv.Itoa(int(alldata.AvatarInfoList[t].EquipList[5].Flat.WeaponStat[1].Value


        //圣遗物
        名字sywname, _ := IdforNamemap[alldata.AvatarInfoList[t].EquipList[i].Flat.SetNameTextHash]