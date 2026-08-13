-- routing.lua — osm2pgsql flex config untuk jaringan jalan yang dapat dilalui kendaraan.
-- Hanya menyimpan jalan berkategori drivable ke tabel `java_ways` (geometry 4326).
-- Tabel ini kemudian diproses pgRouting (pgr_nodeNetwork + pgr_createTopology).

local osm2pgsql = require 'osm2pgsql'

local ways = osm2pgsql.define_way_table('java_ways', {
    { column = 'osm_id', type = 'bigint' },
    { column = 'name',   type = 'text' },
    { column = 'highway', type = 'text' },
    { column = 'oneway', type = 'text' },
    { column = 'geom', type = 'linestring', projection = 4326 },
})

-- Jalan yang boleh dimasukkan ke graf routing (bisa dilewati mobil/motor).
local drivable = {
    motorway = true,         motorway_link = true,
    trunk = true,            trunk_link = true,
    primary = true,          primary_link = true,
    secondary = true,        secondary_link = true,
    tertiary = true,         tertiary_link = true,
    unclassified = true,
    residential = true,
    service = true,
    living_street = true,
    road = true,
}

function osm2pgsql.process_way(object)
    local highway = object.tags.highway
    if not (highway and drivable[highway]) then
        return
    end
    local geom = object:as_linestring()
    if geom == nil then
        return
    end
    ways:insert({
        osm_id  = object.id,
        name    = object.tags.name,
        highway = highway,
        oneway  = object.tags.oneway,
        geom    = geom,
    })
end