package mapobject

import (
	"mapserver/app"
	"mapserver/eventbus"
	"mapserver/mapblockparser"
	"mapserver/mapobjectdb"

	"github.com/sirupsen/logrus"
)

type MapObjectListener interface {
	onMapObject(x, y, z int, block *mapblockparser.MapBlock) *mapobjectdb.MapObject
}

type Listener struct {
	ctx             *app.App
	objectlisteners map[string]MapObjectListener
}

func (this *Listener) AddMapObject(blockname string, ol MapObjectListener) {
	this.objectlisteners[blockname] = ol
}

func (this *Listener) OnEvent(eventtype string, o interface{}) {
	if eventtype != eventbus.MAPBLOCK_RENDERED {
		return
	}

	block := o.(*mapblockparser.MapBlock)

	err := this.ctx.Objectdb.RemoveMapData(block.Pos)
	if err != nil {
		panic(err)
	}

	this.ctx.WebEventbus.Emit("mapobjects-cleared", block.Pos)

	for id, name := range block.BlockMapping {
		for k, v := range this.objectlisteners {
			if k == name {
				//block matches
				mapblockparser.IterateMapblock(func(x,y,z int){
					nodeid := block.GetNodeId(x, y, z)
					if nodeid == id {
						fields := logrus.Fields{
							"mbpos":  block.Pos,
							"x":      x,
							"y":      y,
							"z":      z,
							"type":   name,
							"nodeid": nodeid,
						}
						log.WithFields(fields).Debug("OnEvent()")

						obj := v.onMapObject(x, y, z, block)

						if obj != nil {
							this.ctx.Objectdb.AddMapData(obj)
							this.ctx.WebEventbus.Emit("mapobject-created", obj)
						}
					}
				})
			} // k==name
		} //for k,v
	} //for id, name
}

func Setup(ctx *app.App) {
	l := Listener{
		ctx:             ctx,
		objectlisteners: make(map[string]MapObjectListener),
	}

	//mapserver stuff
	l.AddMapObject("mapserver:poi", &PoiBlock{})
	l.AddMapObject("mapserver:train", &TrainBlock{})
	l.AddMapObject("mapserver:border", &BorderBlock{})
	l.AddMapObject("mapserver:label", &LabelBlock{})

	//travelnet
	l.AddMapObject("travelnet:travelnet", &TravelnetBlock{})

	//protections
	l.AddMapObject("protector:protect", &ProtectorBlock{})
	l.AddMapObject("protector:protect2", &ProtectorBlock{})
	l.AddMapObject("xp_redo:protector", &XPProtectorBlock{})

	//builtin
	l.AddMapObject("bones:bones", &BonesBlock{})

	//technic
	l.AddMapObject("technic:quarry", &QuarryBlock{})
	l.AddMapObject("technic:hv_nuclear_reactor_core_active", &NuclearReactorBlock{})
	l.AddMapObject("technic:admin_anchor", &TechnicAnchorBlock{})

	//digilines
	l.AddMapObject("digilines:lcd", &DigilineLcdBlock{})

	//mesecons
	luac := &LuaControllerBlock{}
	// mesecons_luacontroller:luacontroller0000 2^4=16
	l.AddMapObject("mesecons_luacontroller:luacontroller1111", luac)
	l.AddMapObject("mesecons_luacontroller:luacontroller1110", luac)
	l.AddMapObject("mesecons_luacontroller:luacontroller1100", luac)
	l.AddMapObject("mesecons_luacontroller:luacontroller1010", luac)
	l.AddMapObject("mesecons_luacontroller:luacontroller1000", luac)
	l.AddMapObject("mesecons_luacontroller:luacontroller1101", luac)
	l.AddMapObject("mesecons_luacontroller:luacontroller1001", luac)
	l.AddMapObject("mesecons_luacontroller:luacontroller1011", luac)
	l.AddMapObject("mesecons_luacontroller:luacontroller0111", luac)
	l.AddMapObject("mesecons_luacontroller:luacontroller0110", luac)
	l.AddMapObject("mesecons_luacontroller:luacontroller0100", luac)
	l.AddMapObject("mesecons_luacontroller:luacontroller0010", luac)
	l.AddMapObject("mesecons_luacontroller:luacontroller0000", luac)
	l.AddMapObject("mesecons_luacontroller:luacontroller0101", luac)
	l.AddMapObject("mesecons_luacontroller:luacontroller0001", luac)
	l.AddMapObject("mesecons_luacontroller:luacontroller0011", luac)
	l.AddMapObject("mesecons_luacontroller:luacontroller_burnt", luac)

	//digiterms
	digiterms := &DigitermsBlock{}
	l.AddMapObject("digiterms:lcd_monitor", digiterms)
	l.AddMapObject("digiterms:cathodic_beige_monitor", digiterms)
	l.AddMapObject("digiterms:cathodic_white_monitor", digiterms)
	l.AddMapObject("digiterms:cathodic_black_monitor", digiterms)
	l.AddMapObject("digiterms:scifi_glassscreen", digiterms)
	l.AddMapObject("digiterms:scifi_widescreen", digiterms)
	l.AddMapObject("digiterms:scifi_tallscreen", digiterms)
	l.AddMapObject("digiterms:scifi_keysmonitor", digiterms)

	//missions
	l.AddMapObject("missions:mission", &MissionBlock{})

	//jumpdrive, TODO: fleet controller
	l.AddMapObject("jumpdrive:engine", &JumpdriveBlock{})

	//TODO: atm, digiterms, signs/banners, spacecannons, shops (smart, fancy)

	ctx.BlockAccessor.Eventbus.AddListener(&l)
}
