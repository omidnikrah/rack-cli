package tui

import "rack-cli/internal/entity"

const deletedServerLabel = "deleted server"

type serverIndex struct {
	ips map[int]string
}

func (x *serverIndex) set(rows []entity.Server) {
	x.ips = make(map[int]string, len(rows))
	for _, s := range rows {
		x.ips[s.ID] = s.IP
	}
}

func (x *serverIndex) label(id int) string {
	if ip, ok := x.ips[id]; ok {
		return ip
	}

	return deletedServerLabel
}

func (x *serverIndex) exists(id int) bool {
	_, ok := x.ips[id]
	return ok
}
